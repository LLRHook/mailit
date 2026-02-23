package smtp

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/mail"
	"strings"
	"time"

	gosmtp "github.com/emersion/go-smtp"
	"github.com/google/uuid"
	"github.com/hibiken/asynq"

	"github.com/mailit-dev/mailit/internal/model"
	"github.com/mailit-dev/mailit/internal/worker"
)

// DomainLookup is the interface the SMTP backend needs for domain resolution.
// It looks up a domain by name across all teams and returns it only if verified.
type DomainLookup interface {
	GetVerifiedByName(ctx context.Context, name string) (*model.Domain, error)
}

// InboundEmailStore is the interface the SMTP backend needs for persisting inbound emails.
type InboundEmailStore interface {
	Create(ctx context.Context, email *model.InboundEmail) error
}

// Backend implements the go-smtp Backend interface for receiving inbound emails.
type Backend struct {
	domainLookup     DomainLookup
	inboundEmailRepo InboundEmailStore
	asynqClient      *asynq.Client
	maxMessageBytes  int64
	logger           *slog.Logger
}

// NewBackend creates a new inbound SMTP backend.
func NewBackend(
	domainLookup DomainLookup,
	inboundEmailRepo InboundEmailStore,
	asynqClient *asynq.Client,
	maxMessageBytes int64,
	logger *slog.Logger,
) *Backend {
	return &Backend{
		domainLookup:     domainLookup,
		inboundEmailRepo: inboundEmailRepo,
		asynqClient:      asynqClient,
		maxMessageBytes:  maxMessageBytes,
		logger:           logger,
	}
}

// NewSession is called when a new SMTP connection is established.
func (b *Backend) NewSession(c *gosmtp.Conn) (gosmtp.Session, error) {
	return &Session{
		backend: b,
		logger:  b.logger,
	}, nil
}

// Session represents an SMTP session for receiving an inbound email.
type Session struct {
	backend *Backend
	from    string
	to      []string
	domain  *model.Domain // resolved on first valid Rcpt
	logger  *slog.Logger
}

// Mail is called with the MAIL FROM address.
func (s *Session) Mail(from string, opts *gosmtp.MailOptions) error {
	s.from = from
	return nil
}

// Rcpt is called for each RCPT TO address.
// It validates that the domain part of the recipient is registered and verified.
func (s *Session) Rcpt(to string, opts *gosmtp.RcptOptions) error {
	domainName, err := extractDomain(to)
	if err != nil {
		return &gosmtp.SMTPError{
			Code:         550,
			EnhancedCode: gosmtp.EnhancedCode{5, 1, 1},
			Message:      "invalid recipient address",
		}
	}

	// Look up the domain across all teams.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	domain, err := s.backend.domainLookup.GetVerifiedByName(ctx, domainName)
	if err != nil {
		s.logger.Warn("inbound SMTP: domain lookup failed",
			"domain", domainName,
			"recipient", to,
			"error", err,
		)
		return &gosmtp.SMTPError{
			Code:         550,
			EnhancedCode: gosmtp.EnhancedCode{5, 1, 2},
			Message:      fmt.Sprintf("no such domain: %s", domainName),
		}
	}

	// Keep a reference to the first resolved domain for the team association.
	if s.domain == nil {
		s.domain = domain
	}

	s.to = append(s.to, to)
	return nil
}

// Data is called when the full message body is received.
func (s *Session) Data(r io.Reader) error {
	if s.domain == nil {
		return &gosmtp.SMTPError{
			Code:         503,
			EnhancedCode: gosmtp.EnhancedCode{5, 5, 1},
			Message:      "no valid recipients",
		}
	}

	// Read the message body with a size limit.
	limited := io.LimitReader(r, s.backend.maxMessageBytes)
	body, err := io.ReadAll(limited)
	if err != nil {
		s.logger.Error("inbound SMTP: failed to read message body", "error", err)
		return &gosmtp.SMTPError{
			Code:         451,
			EnhancedCode: gosmtp.EnhancedCode{4, 3, 0},
			Message:      "failed to read message",
		}
	}

	// Parse basic headers from the raw message.
	msg, parseErr := mail.ReadMessage(bytes.NewReader(body))

	var subject string
	var headerFrom, headerTo, headerCc string
	if parseErr == nil {
		subject = msg.Header.Get("Subject")
		headerFrom = msg.Header.Get("From")
		headerTo = msg.Header.Get("To")
		headerCc = msg.Header.Get("Cc")
	} else {
		s.logger.Warn("inbound SMTP: failed to parse message headers", "error", parseErr)
	}

	// Use envelope From if header From is empty.
	fromAddr := headerFrom
	if fromAddr == "" {
		fromAddr = s.from
	}

	// Build the headers map from parsed headers.
	headers := make(model.JSONMap)
	if parseErr == nil {
		for key, values := range msg.Header {
			if len(values) == 1 {
				headers[key] = values[0]
			} else {
				headers[key] = values
			}
		}
	}

	// Parse Cc addresses from header.
	var ccAddresses []string
	if headerCc != "" {
		ccList, err := mail.ParseAddressList(headerCc)
		if err == nil {
			for _, addr := range ccList {
				ccAddresses = append(ccAddresses, addr.Address)
			}
		}
	}

	// Use parsed To addresses from header, fall back to envelope.
	toAddresses := s.to
	if headerTo != "" {
		toList, err := mail.ParseAddressList(headerTo)
		if err == nil {
			parsed := make([]string, 0, len(toList))
			for _, addr := range toList {
				parsed = append(parsed, addr.Address)
			}
			toAddresses = parsed
		}
	}

	rawMessage := string(body)
	now := time.Now().UTC()
	domainID := s.domain.ID

	inbound := &model.InboundEmail{
		ID:          uuid.New(),
		TeamID:      s.domain.TeamID,
		DomainID:    &domainID,
		FromAddress: fromAddr,
		ToAddresses: toAddresses,
		CcAddresses: ccAddresses,
		Subject:     ptrString(subject),
		RawMessage:  &rawMessage,
		Headers:     headers,
		Attachments: make(model.JSONArray, 0),
		Processed:   false,
		CreatedAt:   now,
	}

	// Persist the inbound email.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := s.backend.inboundEmailRepo.Create(ctx, inbound); err != nil {
		s.logger.Error("inbound SMTP: failed to save inbound email",
			"error", err,
			"from", fromAddr,
			"to", toAddresses,
		)
		return &gosmtp.SMTPError{
			Code:         451,
			EnhancedCode: gosmtp.EnhancedCode{4, 3, 0},
			Message:      "temporary error storing message",
		}
	}

	// Enqueue the inbound:process worker task.
	task, err := worker.NewInboundProcessTask(inbound.ID)
	if err != nil {
		s.logger.Error("inbound SMTP: failed to create inbound process task",
			"error", err,
			"inbound_email_id", inbound.ID,
		)
		// The email is already saved; the worker can pick it up later.
		return nil
	}

	if _, err := s.backend.asynqClient.Enqueue(task); err != nil {
		s.logger.Error("inbound SMTP: failed to enqueue inbound process task",
			"error", err,
			"inbound_email_id", inbound.ID,
		)
		// Non-fatal: the email is persisted and can be reprocessed.
	}

	s.logger.Info("inbound SMTP: email received and queued",
		"inbound_email_id", inbound.ID,
		"from", fromAddr,
		"to", toAddresses,
		"subject", subject,
	)

	return nil
}

// Reset is called between messages in the same SMTP session.
func (s *Session) Reset() {
	s.from = ""
	s.to = nil
	s.domain = nil
}

// Logout is called when the SMTP session ends.
func (s *Session) Logout() error {
	return nil
}

// extractDomain returns the domain part of an email address.
func extractDomain(email string) (string, error) {
	at := strings.LastIndex(email, "@")
	if at < 0 || at == len(email)-1 {
		return "", fmt.Errorf("invalid email address: %s", email)
	}
	return strings.ToLower(email[at+1:]), nil
}

// ptrString returns a pointer to the given string, or nil if empty.
func ptrString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
