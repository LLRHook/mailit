package engine

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"io"
	"log/slog"
	"mime"
	"mime/multipart"
	"mime/quotedprintable"
	"net"
	"net/smtp"
	"net/textproto"
	"strings"
	"time"
)

// SenderMetrics is an optional interface for recording SMTP metrics.
// Pass nil to disable metrics.
type SenderMetrics interface {
	ObserveEmailSendDuration(seconds float64)
	IncSMTPConnection(mxHost, result string)
}

// Sender handles sending emails via SMTP directly to recipient MX servers.
type Sender struct {
	hostname       string
	heloDomain     string
	tlsPolicy      string // "opportunistic" or "enforce"
	connectTimeout time.Duration
	sendTimeout    time.Duration
	maxRecipients  int
	resolver       *DNSResolver
	logger         *slog.Logger
	circuitBreaker *CircuitBreaker
	metrics        SenderMetrics

	// Relay mode (SES)
	relayMode     string // "direct" or "relay"
	relayHost     string
	relayPort     int
	relayUsername string
	relayPassword string
	relayTLS      string // "starttls" or "tls"
}

// SenderConfig configures the SMTP sender.
type SenderConfig struct {
	Hostname       string
	HeloDomain     string
	TLSPolicy      string
	ConnectTimeout time.Duration
	SendTimeout    time.Duration
	MaxRecipients  int
	Metrics        SenderMetrics

	// Relay mode fields
	RelayMode     string
	RelayHost     string
	RelayPort     int
	RelayUsername string
	RelayPassword string
	RelayTLS      string
}

// OutgoingMessage holds all the data needed to build and send an email.
type OutgoingMessage struct {
	From         string
	To           []string
	Cc           []string
	Bcc          []string
	ReplyTo      string
	Subject      string
	HTMLBody     string
	TextBody     string
	Headers      map[string]string
	Attachments  []MessageAttachment
	MessageID    string
	DKIMDomain   string
	DKIMSelector string
	DKIMKey      string // decrypted PEM private key
}

// MessageAttachment represents a file attached to an email.
type MessageAttachment struct {
	Filename    string
	Content     []byte
	ContentType string
}

// SendResult holds the outcome of a send operation.
type SendResult struct {
	MessageID  string
	Recipients map[string]RecipientResult
}

// RecipientResult holds the delivery result for a single recipient.
type RecipientResult struct {
	Status    string // "sent", "failed", "deferred"
	Code      int    // SMTP response code
	Message   string // SMTP response message
	Permanent bool   // true for 5xx errors
}

// NewSender creates a new SMTP sender with the given configuration.
func NewSender(cfg SenderConfig, resolver *DNSResolver, logger *slog.Logger) *Sender {
	if cfg.ConnectTimeout == 0 {
		cfg.ConnectTimeout = 30 * time.Second
	}
	if cfg.SendTimeout == 0 {
		cfg.SendTimeout = 5 * time.Minute
	}
	if cfg.MaxRecipients == 0 {
		cfg.MaxRecipients = 50
	}
	if cfg.TLSPolicy == "" {
		cfg.TLSPolicy = "opportunistic"
	}
	if cfg.HeloDomain == "" {
		cfg.HeloDomain = cfg.Hostname
	}

	relayMode := cfg.RelayMode
	if relayMode == "" {
		relayMode = "direct"
	}
	relayPort := cfg.RelayPort
	if relayPort == 0 {
		relayPort = 587
	}
	relayTLS := cfg.RelayTLS
	if relayTLS == "" {
		relayTLS = "starttls"
	}

	return &Sender{
		hostname:       cfg.Hostname,
		heloDomain:     cfg.HeloDomain,
		tlsPolicy:      cfg.TLSPolicy,
		connectTimeout: cfg.ConnectTimeout,
		sendTimeout:    cfg.SendTimeout,
		maxRecipients:  cfg.MaxRecipients,
		resolver:       resolver,
		logger:         logger,
		circuitBreaker: NewCircuitBreaker(defaultFailureThreshold, defaultResetTimeout),
		metrics:        cfg.Metrics,
		relayMode:      relayMode,
		relayHost:      cfg.RelayHost,
		relayPort:      relayPort,
		relayUsername:  cfg.RelayUsername,
		relayPassword:  cfg.RelayPassword,
		relayTLS:       relayTLS,
	}
}

// SendEmail builds a MIME message, signs it with DKIM, and delivers it directly
// to each recipient's MX server. Recipients are grouped by domain so each domain
// gets a single SMTP session where possible.
func (s *Sender) SendEmail(ctx context.Context, msg *OutgoingMessage) (*SendResult, error) {
	if len(msg.To) == 0 && len(msg.Cc) == 0 && len(msg.Bcc) == 0 {
		return nil, fmt.Errorf("no recipients specified")
	}

	allRecipients := collectRecipients(msg)
	if len(allRecipients) > s.maxRecipients {
		return nil, fmt.Errorf("too many recipients: %d exceeds maximum %d", len(allRecipients), s.maxRecipients)
	}

	// Build the RFC 5322 MIME message.
	rawMessage, err := BuildMessage(msg)
	if err != nil {
		return nil, fmt.Errorf("building message: %w", err)
	}

	// Sign with DKIM if configured.
	signedMessage := rawMessage
	if msg.DKIMKey != "" && msg.DKIMDomain != "" && msg.DKIMSelector != "" {
		signedMessage, err = SignMessage(rawMessage, msg.DKIMDomain, msg.DKIMSelector, msg.DKIMKey)
		if err != nil {
			return nil, fmt.Errorf("DKIM signing: %w", err)
		}
	}

	result := &SendResult{
		MessageID:  msg.MessageID,
		Recipients: make(map[string]RecipientResult),
	}

	if s.relayMode == "relay" {
		// Deliver through a relay (e.g. SES) instead of direct MX delivery.
		s.deliverViaRelay(ctx, allRecipients, msg.From, signedMessage, result)
	} else {
		// Direct delivery: group recipients by domain.
		domainRecipients := groupByDomain(allRecipients)
		for domain, recipients := range domainRecipients {
			s.deliverToDomain(ctx, domain, recipients, msg.From, signedMessage, result)
		}
	}

	return result, nil
}

// collectRecipients gathers all unique recipient addresses from To, Cc, and Bcc.
func collectRecipients(msg *OutgoingMessage) []string {
	seen := make(map[string]bool)
	var recipients []string

	for _, lists := range [][]string{msg.To, msg.Cc, msg.Bcc} {
		for _, addr := range lists {
			lower := strings.ToLower(strings.TrimSpace(addr))
			if lower != "" && !seen[lower] {
				seen[lower] = true
				recipients = append(recipients, lower)
			}
		}
	}

	return recipients
}

// groupByDomain groups email addresses by their domain part.
func groupByDomain(recipients []string) map[string][]string {
	groups := make(map[string][]string)
	for _, addr := range recipients {
		parts := strings.SplitN(addr, "@", 2)
		if len(parts) != 2 {
			continue
		}
		domain := strings.ToLower(parts[1])
		groups[domain] = append(groups[domain], addr)
	}
	return groups
}

// deliverViaRelay sends email through a configured SMTP relay (e.g. SES, Mailgun).
func (s *Sender) deliverViaRelay(
	ctx context.Context,
	recipients []string,
	from string,
	message []byte,
	result *SendResult,
) {
	addr := fmt.Sprintf("%s:%d", s.relayHost, s.relayPort)
	for _, rcpt := range recipients {
		s.deliverToHost(ctx, addr, from, []string{rcpt}, message, result)
	}
}

// deliverToDomain resolves MX records for the domain and attempts delivery
// through each MX host in priority order until one succeeds.
func (s *Sender) deliverToDomain(
	ctx context.Context,
	domain string,
	recipients []string,
	from string,
	message []byte,
	result *SendResult,
) {
	mxRecords, err := s.resolver.LookupMX(domain)
	if err != nil {
		s.logger.Error("MX lookup failed", "domain", domain, "error", err)
		for _, rcpt := range recipients {
			result.Recipients[rcpt] = RecipientResult{
				Status:    "failed",
				Code:      0,
				Message:   fmt.Sprintf("MX lookup failed: %v", err),
				Permanent: false,
			}
		}
		return
	}

	// Try each MX host in priority order.
	var lastErr error
	for _, mx := range mxRecords {
		select {
		case <-ctx.Done():
			for _, rcpt := range recipients {
				if _, ok := result.Recipients[rcpt]; !ok {
					result.Recipients[rcpt] = RecipientResult{
						Status:  "failed",
						Code:    0,
						Message: "context cancelled",
					}
				}
			}
			return
		default:
		}

		// Circuit breaker: skip MX hosts that are in open state.
		if !s.circuitBreaker.Allow(mx.Host) {
			s.logger.Warn("circuit breaker open, skipping MX host",
				"domain", domain,
				"mx_host", mx.Host,
			)
			continue
		}

		s.logger.Debug("attempting delivery",
			"domain", domain,
			"mx_host", mx.Host,
			"mx_priority", mx.Priority,
			"recipients", len(recipients),
		)

		err := s.deliverToHost(ctx, mx.Host, from, recipients, message, result)
		if err == nil {
			s.circuitBreaker.RecordSuccess(mx.Host)
			return // Successfully delivered.
		}
		s.circuitBreaker.RecordFailure(mx.Host)
		lastErr = err
		s.logger.Warn("delivery attempt failed",
			"mx_host", mx.Host,
			"error", err,
		)
	}

	// All MX hosts failed. Mark undelivered recipients as deferred.
	for _, rcpt := range recipients {
		if _, ok := result.Recipients[rcpt]; !ok {
			result.Recipients[rcpt] = RecipientResult{
				Status:    "deferred",
				Code:      0,
				Message:   fmt.Sprintf("all MX hosts failed: %v", lastErr),
				Permanent: false,
			}
		}
	}
}

// deliverToHost connects to a single MX host and attempts SMTP delivery.
func (s *Sender) deliverToHost(
	ctx context.Context,
	host string,
	from string,
	recipients []string,
	message []byte,
	result *SendResult,
) error {
	start := time.Now()
	addr := host + ":25"

	// Connect with timeout.
	dialer := net.Dialer{Timeout: s.connectTimeout}
	conn, err := dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		s.recordSMTPConnection(host, "connect_error")
		return fmt.Errorf("connecting to %s: %w", addr, err)
	}

	// Set an overall deadline for the SMTP session.
	if err := conn.SetDeadline(time.Now().Add(s.sendTimeout)); err != nil {
		_ = conn.Close()
		return fmt.Errorf("setting deadline: %w", err)
	}

	client, err := smtp.NewClient(conn, host)
	if err != nil {
		_ = conn.Close()
		return fmt.Errorf("creating SMTP client for %s: %w", host, err)
	}
	defer func() { _ = client.Close() }()

	// Send EHLO.
	if err := client.Hello(s.heloDomain); err != nil {
		return fmt.Errorf("EHLO to %s: %w", host, err)
	}

	// Attempt STARTTLS.
	if ok, _ := client.Extension("STARTTLS"); ok {
		tlsConfig := &tls.Config{
			ServerName: host,
		}
		if err := client.StartTLS(tlsConfig); err != nil {
			if s.tlsPolicy == "enforce" {
				return fmt.Errorf("STARTTLS required but failed for %s: %w", host, err)
			}
			s.logger.Warn("STARTTLS failed, continuing without TLS",
				"host", host,
				"error", err,
			)
		}
	} else if s.tlsPolicy == "enforce" {
		return fmt.Errorf("STARTTLS required but not offered by %s", host)
	}

	// MAIL FROM.
	if err := client.Mail(from); err != nil {
		return fmt.Errorf("MAIL FROM to %s: %w", host, err)
	}

	// RCPT TO for each recipient. Track per-recipient failures.
	var validRecipients []string
	for _, rcpt := range recipients {
		if err := client.Rcpt(rcpt); err != nil {
			code, msg := parseSmtpError(err)
			bounce := ClassifyBounce(code, msg)
			result.Recipients[rcpt] = RecipientResult{
				Status:    statusFromBounce(bounce),
				Code:      code,
				Message:   msg,
				Permanent: bounce.Permanent,
			}
			s.logger.Warn("RCPT TO rejected",
				"recipient", rcpt,
				"host", host,
				"code", code,
				"message", msg,
			)
		} else {
			validRecipients = append(validRecipients, rcpt)
		}
	}

	if len(validRecipients) == 0 {
		_ = client.Reset()
		return nil // All recipients rejected; results already recorded.
	}

	// DATA.
	wc, err := client.Data()
	if err != nil {
		code, msg := parseSmtpError(err)
		for _, rcpt := range validRecipients {
			result.Recipients[rcpt] = RecipientResult{
				Status:    "failed",
				Code:      code,
				Message:   msg,
				Permanent: code >= 500,
			}
		}
		return fmt.Errorf("DATA to %s: %w", host, err)
	}

	if _, err := wc.Write(message); err != nil {
		_ = wc.Close()
		return fmt.Errorf("writing message data to %s: %w", host, err)
	}

	if err := wc.Close(); err != nil {
		code, msg := parseSmtpError(err)
		for _, rcpt := range validRecipients {
			result.Recipients[rcpt] = RecipientResult{
				Status:    "failed",
				Code:      code,
				Message:   msg,
				Permanent: code >= 500,
			}
		}
		return fmt.Errorf("closing DATA to %s: %w", host, err)
	}

	// Mark all valid recipients as sent.
	for _, rcpt := range validRecipients {
		result.Recipients[rcpt] = RecipientResult{
			Status:  "sent",
			Code:    250,
			Message: "OK",
		}
	}

	_ = client.Quit()
	s.recordSMTPConnection(host, "success")
	s.recordEmailSendDuration(time.Since(start).Seconds())
	return nil
}

// recordSMTPConnection records an SMTP connection metric if metrics are configured.
func (s *Sender) recordSMTPConnection(host, result string) {
	if s.metrics != nil {
		s.metrics.IncSMTPConnection(host, result)
	}
}

// recordEmailSendDuration records email send duration if metrics are configured.
func (s *Sender) recordEmailSendDuration(seconds float64) {
	if s.metrics != nil {
		s.metrics.ObserveEmailSendDuration(seconds)
	}
}

// BuildMessage constructs an RFC 5322 MIME message from the outgoing message.
// It produces a multipart/mixed message when attachments are present, with a
// multipart/alternative body for text and HTML parts.
func BuildMessage(msg *OutgoingMessage) ([]byte, error) {
	var buf bytes.Buffer
	headers := textproto.MIMEHeader{}

	// Required headers.
	headers.Set("From", msg.From)
	headers.Set("Subject", encodeSubject(msg.Subject))
	headers.Set("Date", time.Now().UTC().Format("Mon, 02 Jan 2006 15:04:05 -0700"))
	headers.Set("MIME-Version", "1.0")

	if msg.MessageID != "" {
		headers.Set("Message-ID", "<"+msg.MessageID+">")
	}

	if len(msg.To) > 0 {
		headers.Set("To", strings.Join(msg.To, ", "))
	}
	if len(msg.Cc) > 0 {
		headers.Set("Cc", strings.Join(msg.Cc, ", "))
	}
	// Bcc is intentionally omitted from headers.

	if msg.ReplyTo != "" {
		headers.Set("Reply-To", msg.ReplyTo)
	}

	// Custom headers.
	for key, value := range msg.Headers {
		headers.Set(key, value)
	}

	hasText := msg.TextBody != ""
	hasHTML := msg.HTMLBody != ""
	hasAttachments := len(msg.Attachments) > 0

	switch {
	case hasAttachments:
		// multipart/mixed wrapping a multipart/alternative body + attachments.
		if err := buildMultipartMixed(&buf, headers, msg); err != nil {
			return nil, err
		}
	case hasText && hasHTML:
		// multipart/alternative with text and HTML parts.
		if err := buildMultipartAlternative(&buf, headers, msg.TextBody, msg.HTMLBody); err != nil {
			return nil, err
		}
	case hasHTML:
		buildSinglePart(&buf, headers, "text/html; charset=utf-8", msg.HTMLBody)
	case hasText:
		buildSinglePart(&buf, headers, "text/plain; charset=utf-8", msg.TextBody)
	default:
		buildSinglePart(&buf, headers, "text/plain; charset=utf-8", "")
	}

	return buf.Bytes(), nil
}

// writeHeaders writes MIME headers to the buffer.
func writeHeaders(buf *bytes.Buffer, headers textproto.MIMEHeader) {
	// Write headers in a consistent order for DKIM reproducibility.
	orderedKeys := []string{
		"From", "To", "Cc", "Reply-To", "Subject",
		"Date", "Message-Id", "Mime-Version", "Content-Type",
	}
	written := make(map[string]bool)

	for _, key := range orderedKeys {
		if values, ok := headers[textproto.CanonicalMIMEHeaderKey(key)]; ok {
			for _, v := range values {
				fmt.Fprintf(buf, "%s: %s\r\n", textproto.CanonicalMIMEHeaderKey(key), v)
			}
			written[textproto.CanonicalMIMEHeaderKey(key)] = true
		}
	}

	// Write any remaining headers not in the ordered list.
	for key, values := range headers {
		if written[key] {
			continue
		}
		for _, v := range values {
			fmt.Fprintf(buf, "%s: %s\r\n", key, v)
		}
	}

	buf.WriteString("\r\n")
}

// buildSinglePart writes a simple single-part message.
func buildSinglePart(buf *bytes.Buffer, headers textproto.MIMEHeader, contentType, body string) {
	headers.Set("Content-Type", contentType)
	headers.Set("Content-Transfer-Encoding", "quoted-printable")
	writeHeaders(buf, headers)

	w := quotedprintable.NewWriter(buf)
	_, _ = w.Write([]byte(body))
	_ = w.Close()
}

// buildMultipartAlternative writes a multipart/alternative message with text
// and HTML parts.
func buildMultipartAlternative(buf *bytes.Buffer, headers textproto.MIMEHeader, textBody, htmlBody string) error {
	w := multipart.NewWriter(buf)
	headers.Set("Content-Type", fmt.Sprintf("multipart/alternative; boundary=%s", w.Boundary()))
	writeHeaders(buf, headers)

	// Text part.
	textHeaders := textproto.MIMEHeader{}
	textHeaders.Set("Content-Type", "text/plain; charset=utf-8")
	textHeaders.Set("Content-Transfer-Encoding", "quoted-printable")
	textPart, err := w.CreatePart(textHeaders)
	if err != nil {
		return fmt.Errorf("creating text part: %w", err)
	}
	qw := quotedprintable.NewWriter(textPart)
	_, _ = qw.Write([]byte(textBody))
	_ = qw.Close()

	// HTML part.
	htmlHeaders := textproto.MIMEHeader{}
	htmlHeaders.Set("Content-Type", "text/html; charset=utf-8")
	htmlHeaders.Set("Content-Transfer-Encoding", "quoted-printable")
	htmlPart, err := w.CreatePart(htmlHeaders)
	if err != nil {
		return fmt.Errorf("creating HTML part: %w", err)
	}
	qw = quotedprintable.NewWriter(htmlPart)
	_, _ = qw.Write([]byte(htmlBody))
	_ = qw.Close()

	return w.Close()
}

// buildMultipartMixed writes a multipart/mixed message containing a
// multipart/alternative body and file attachments.
func buildMultipartMixed(buf *bytes.Buffer, headers textproto.MIMEHeader, msg *OutgoingMessage) error {
	mixedWriter := multipart.NewWriter(buf)
	headers.Set("Content-Type", fmt.Sprintf("multipart/mixed; boundary=%s", mixedWriter.Boundary()))
	writeHeaders(buf, headers)

	hasText := msg.TextBody != ""
	hasHTML := msg.HTMLBody != ""

	// Write the body part(s).
	if hasText && hasHTML {
		// Nested multipart/alternative inside the mixed message.
		// Create a temporary writer to generate a boundary, then use it for
		// both the Content-Type header and the actual nested writer.
		boundary := multipart.NewWriter(nil).Boundary()
		altHeaders := textproto.MIMEHeader{}
		altHeaders.Set("Content-Type", fmt.Sprintf("multipart/alternative; boundary=%s", boundary))
		altPart, err := mixedWriter.CreatePart(altHeaders)
		if err != nil {
			return fmt.Errorf("creating alternative part: %w", err)
		}

		nestedAlt := multipart.NewWriter(altPart)
		_ = nestedAlt.SetBoundary(boundary)

		// Text part.
		textHeaders := textproto.MIMEHeader{}
		textHeaders.Set("Content-Type", "text/plain; charset=utf-8")
		textHeaders.Set("Content-Transfer-Encoding", "quoted-printable")
		textPart, err := nestedAlt.CreatePart(textHeaders)
		if err != nil {
			return fmt.Errorf("creating text part: %w", err)
		}
		qw := quotedprintable.NewWriter(textPart)
		_, _ = qw.Write([]byte(msg.TextBody))
		_ = qw.Close()

		// HTML part.
		htmlHeaders := textproto.MIMEHeader{}
		htmlHeaders.Set("Content-Type", "text/html; charset=utf-8")
		htmlHeaders.Set("Content-Transfer-Encoding", "quoted-printable")
		htmlPart, err := nestedAlt.CreatePart(htmlHeaders)
		if err != nil {
			return fmt.Errorf("creating HTML part: %w", err)
		}
		qw = quotedprintable.NewWriter(htmlPart)
		_, _ = qw.Write([]byte(msg.HTMLBody))
		_ = qw.Close()

		if err := nestedAlt.Close(); err != nil {
			return fmt.Errorf("closing alternative writer: %w", err)
		}
	} else if hasHTML {
		htmlHeaders := textproto.MIMEHeader{}
		htmlHeaders.Set("Content-Type", "text/html; charset=utf-8")
		htmlHeaders.Set("Content-Transfer-Encoding", "quoted-printable")
		htmlPart, err := mixedWriter.CreatePart(htmlHeaders)
		if err != nil {
			return fmt.Errorf("creating HTML part: %w", err)
		}
		qw := quotedprintable.NewWriter(htmlPart)
		_, _ = qw.Write([]byte(msg.HTMLBody))
		_ = qw.Close()
	} else if hasText {
		textHeaders := textproto.MIMEHeader{}
		textHeaders.Set("Content-Type", "text/plain; charset=utf-8")
		textHeaders.Set("Content-Transfer-Encoding", "quoted-printable")
		textPart, err := mixedWriter.CreatePart(textHeaders)
		if err != nil {
			return fmt.Errorf("creating text part: %w", err)
		}
		qw := quotedprintable.NewWriter(textPart)
		_, _ = qw.Write([]byte(msg.TextBody))
		_ = qw.Close()
	}

	// Write attachment parts.
	for _, att := range msg.Attachments {
		contentType := att.ContentType
		if contentType == "" {
			contentType = "application/octet-stream"
		}

		attHeaders := textproto.MIMEHeader{}
		attHeaders.Set("Content-Type", contentType+"; name=\""+att.Filename+"\"")
		attHeaders.Set("Content-Transfer-Encoding", "base64")
		attHeaders.Set("Content-Disposition",
			mime.FormatMediaType("attachment", map[string]string{"filename": att.Filename}))

		attPart, err := mixedWriter.CreatePart(attHeaders)
		if err != nil {
			return fmt.Errorf("creating attachment part for %s: %w", att.Filename, err)
		}

		encoder := base64.NewEncoder(base64.StdEncoding, &lineWrapper{writer: attPart, lineLen: 76})
		if _, err := encoder.Write(att.Content); err != nil {
			return fmt.Errorf("encoding attachment %s: %w", att.Filename, err)
		}
		_ = encoder.Close()
	}

	return mixedWriter.Close()
}

// lineWrapper wraps base64 output at the specified line length with CRLF.
type lineWrapper struct {
	writer io.Writer
	lineLen int
	current int
}

func (lw *lineWrapper) Write(p []byte) (int, error) {
	total := 0
	for len(p) > 0 {
		remaining := lw.lineLen - lw.current
		if remaining <= 0 {
			if _, err := lw.writer.Write([]byte("\r\n")); err != nil {
				return total, err
			}
			lw.current = 0
			remaining = lw.lineLen
		}

		chunk := p
		if len(chunk) > remaining {
			chunk = p[:remaining]
		}

		n, err := lw.writer.Write(chunk)
		total += n
		lw.current += n
		p = p[n:]
		if err != nil {
			return total, err
		}
	}
	return total, nil
}

// encodeSubject encodes a subject line using RFC 2047 if it contains non-ASCII.
func encodeSubject(subject string) string {
	for _, r := range subject {
		if r > 127 {
			return mime.QEncoding.Encode("utf-8", subject)
		}
	}
	return subject
}

// parseSmtpError extracts the SMTP response code and message from an error.
func parseSmtpError(err error) (int, string) {
	if err == nil {
		return 0, ""
	}

	msg := err.Error()

	// Try to parse "XXX Y.Y.Y message" or "XXX message" format.
	if len(msg) >= 3 {
		var code int
		if _, parseErr := fmt.Sscanf(msg[:3], "%d", &code); parseErr == nil && code >= 200 && code < 600 {
			return code, strings.TrimSpace(msg[3:])
		}
	}

	// Could not parse; guess from the error message.
	if strings.Contains(strings.ToLower(msg), "timeout") ||
		strings.Contains(strings.ToLower(msg), "connection refused") {
		return 421, msg
	}

	return 0, msg
}

// statusFromBounce maps a BounceInfo to a delivery status string.
func statusFromBounce(b BounceInfo) string {
	switch b.Type {
	case BounceHard:
		return "failed"
	case BounceSoft:
		return "deferred"
	case BounceComplaint:
		return "failed"
	default:
		return "deferred"
	}
}
