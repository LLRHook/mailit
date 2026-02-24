package engine

import (
	"bufio"
	"bytes"
	"fmt"
	"mime"
	"mime/multipart"
	"net/mail"
	"strconv"
	"strings"
)

// BounceType classifies an SMTP error response.
type BounceType string

const (
	BounceHard      BounceType = "hard"      // 5xx - permanent, suppress address
	BounceSoft      BounceType = "soft"       // 4xx - temporary, retry later
	BounceComplaint BounceType = "complaint"  // spam complaint from recipient
)

// BounceInfo contains details about a bounced email.
type BounceInfo struct {
	Type      BounceType
	Code      int
	Message   string
	Recipient string
	Permanent bool
}

// ClassifyBounce analyzes an SMTP error code and message to determine the
// bounce type and whether the failure is permanent. Specific SMTP enhanced
// status codes are handled for more precise classification.
func ClassifyBounce(code int, message string) BounceInfo {
	info := BounceInfo{
		Code:    code,
		Message: message,
	}

	lowerMsg := strings.ToLower(message)

	// Check for spam/complaint indicators regardless of code.
	if containsAny(lowerMsg, "spam", "unsolicited", "abuse", "complaint", "blocked for spam") {
		info.Type = BounceComplaint
		info.Permanent = true
		return info
	}

	switch {
	case code >= 500 && code < 600:
		info.Type = BounceHard
		info.Permanent = true

		switch code {
		case 550:
			// Mailbox not found, does not exist, or rejected.
		case 551:
			// User not local; sometimes a forward reference.
		case 552:
			// Mailbox full / quota exceeded: treat as soft bounce since it
			// may clear up when the recipient frees space.
			if containsAny(lowerMsg, "quota", "mailbox full", "over quota", "storage") {
				info.Type = BounceSoft
				info.Permanent = false
			}
		case 553:
			// Mailbox name not allowed (syntax error in address).
		case 554:
			// Transaction failed. Could be policy or content rejection.
		}

	case code >= 400 && code < 500:
		info.Type = BounceSoft
		info.Permanent = false

		switch code {
		case 421:
			// Service not available, closing connection (temporary).
		case 450:
			// Mailbox unavailable (busy or temporarily blocked).
		case 451:
			// Local error in processing; try again.
		case 452:
			// Insufficient storage; try again later.
		}

	default:
		// Unknown code range: default to soft bounce to avoid suppressing
		// addresses on unexpected codes.
		info.Type = BounceSoft
		info.Permanent = false
	}

	return info
}

// ClassifyDSN parses a Delivery Status Notification (bounce email) per
// RFC 3464 and extracts bounce information. DSN messages use Content-Type
// multipart/report with report-type=delivery-status.
func ClassifyDSN(rawMessage []byte) (*BounceInfo, error) {
	msg, err := mail.ReadMessage(bytes.NewReader(rawMessage))
	if err != nil {
		return nil, fmt.Errorf("parsing DSN message: %w", err)
	}

	contentType := msg.Header.Get("Content-Type")
	if contentType == "" {
		return nil, fmt.Errorf("missing Content-Type header")
	}

	mediaType, params, err := mime.ParseMediaType(contentType)
	if err != nil {
		return nil, fmt.Errorf("parsing Content-Type: %w", err)
	}

	// DSN messages should be multipart/report.
	if mediaType != "multipart/report" {
		return nil, fmt.Errorf("unexpected Content-Type %q, expected multipart/report", mediaType)
	}

	reportType := params["report-type"]
	if reportType != "" && reportType != "delivery-status" {
		return nil, fmt.Errorf("unexpected report-type %q, expected delivery-status", reportType)
	}

	boundary := params["boundary"]
	if boundary == "" {
		return nil, fmt.Errorf("missing boundary in Content-Type")
	}

	reader := multipart.NewReader(msg.Body, boundary)

	var info BounceInfo
	foundStatus := false

	for {
		part, err := reader.NextPart()
		if err != nil {
			break
		}

		partType := part.Header.Get("Content-Type")
		partMedia, _, _ := mime.ParseMediaType(partType)

		// The delivery-status part contains the structured bounce data.
		if partMedia == "message/delivery-status" {
			if err := parseDSNStatus(part, &info); err != nil {
				return nil, fmt.Errorf("parsing delivery-status: %w", err)
			}
			foundStatus = true
		}

		_ = part.Close()
	}

	if !foundStatus {
		return nil, fmt.Errorf("no message/delivery-status part found in DSN")
	}

	return &info, nil
}

// parseDSNStatus reads a message/delivery-status MIME part and populates
// the BounceInfo from its fields. The delivery-status part contains groups
// of header-like fields separated by blank lines.
func parseDSNStatus(part *multipart.Part, info *BounceInfo) error {
	scanner := bufio.NewScanner(part)

	for scanner.Scan() {
		line := scanner.Text()

		// Skip blank lines (group separators).
		if strings.TrimSpace(line) == "" {
			continue
		}

		colonIdx := strings.Index(line, ":")
		if colonIdx < 0 {
			continue
		}

		key := strings.TrimSpace(strings.ToLower(line[:colonIdx]))
		value := strings.TrimSpace(line[colonIdx+1:])

		switch key {
		case "status":
			parseDSNStatusCode(value, info)
		case "final-recipient":
			// Format: rfc822;user@example.com
			if idx := strings.Index(value, ";"); idx >= 0 {
				info.Recipient = strings.TrimSpace(value[idx+1:])
			}
		case "original-recipient":
			// Use as fallback if final-recipient is missing.
			if info.Recipient == "" {
				if idx := strings.Index(value, ";"); idx >= 0 {
					info.Recipient = strings.TrimSpace(value[idx+1:])
				}
			}
		case "diagnostic-code":
			info.Message = value
			// Try to extract SMTP code from diagnostic.
			parseDiagnosticCode(value, info)
		case "action":
			action := strings.ToLower(value)
			switch action {
			case "failed":
				info.Permanent = true
				info.Type = BounceHard
			case "delayed", "relayed", "expanded":
				info.Permanent = false
				info.Type = BounceSoft
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("reading delivery-status: %w", err)
	}

	// Default classification if not set by action field.
	if info.Type == "" {
		info.Type = BounceSoft
	}

	return nil
}

// parseDSNStatusCode parses an enhanced status code (e.g., "5.1.1") from a
// DSN Status field and classifies the bounce.
func parseDSNStatusCode(status string, info *BounceInfo) {
	// Enhanced status codes are in the form class.subject.detail (e.g. 5.1.1).
	parts := strings.SplitN(status, ".", 3)
	if len(parts) < 1 {
		return
	}

	class, err := strconv.Atoi(parts[0])
	if err != nil {
		return
	}

	switch class {
	case 5:
		info.Type = BounceHard
		info.Permanent = true
		info.Code = 550 // Default 5xx code.

		// Refine code based on subject.detail if available.
		if len(parts) >= 3 {
			subject, _ := strconv.Atoi(parts[1])
			detail, _ := strconv.Atoi(parts[2])

			switch {
			case subject == 1 && detail == 1:
				// 5.1.1: Bad destination mailbox address.
				info.Code = 550
			case subject == 1 && detail == 2:
				// 5.1.2: Bad destination system address.
				info.Code = 550
			case subject == 2 && detail == 0:
				// 5.2.0: Other/undefined mailbox status.
				info.Code = 550
			case subject == 2 && detail == 1:
				// 5.2.1: Mailbox disabled, not accepting messages.
				info.Code = 550
			case subject == 2 && detail == 2:
				// 5.2.2: Mailbox full.
				info.Type = BounceSoft
				info.Permanent = false
				info.Code = 552
			case subject == 3:
				// 5.3.x: Mail system issues.
				info.Code = 554
			case subject == 4:
				// 5.4.x: Network/routing issues.
				info.Code = 554
			case subject == 7:
				// 5.7.x: Security/policy rejection.
				info.Code = 554
			}
		}

	case 4:
		info.Type = BounceSoft
		info.Permanent = false
		info.Code = 450

		if len(parts) >= 3 {
			subject, _ := strconv.Atoi(parts[1])
			switch subject {
			case 2:
				// 4.2.x: Mailbox issues (full, disabled temporarily).
				info.Code = 452
			case 4:
				// 4.4.x: Network/connection issues.
				info.Code = 421
			case 7:
				// 4.7.x: Temporary security/policy hold (greylisting, etc.).
				info.Code = 450
			}
		}

	case 2:
		// 2.x.x: Success. Not a bounce.
		info.Type = ""
		info.Permanent = false
		info.Code = 250
	}
}

// parseDiagnosticCode attempts to extract an SMTP response code from a
// diagnostic-code field (e.g., "smtp; 550 5.1.1 User unknown").
func parseDiagnosticCode(diagnostic string, info *BounceInfo) {
	// Strip the transport type prefix.
	if idx := strings.Index(diagnostic, ";"); idx >= 0 {
		diagnostic = strings.TrimSpace(diagnostic[idx+1:])
	}

	// Try to parse the leading SMTP code.
	if len(diagnostic) >= 3 {
		code, err := strconv.Atoi(diagnostic[:3])
		if err == nil && code >= 200 && code < 600 {
			info.Code = code
			// Reclassify based on the actual SMTP code if the DSN status
			// was ambiguous.
			classified := ClassifyBounce(code, info.Message)
			info.Type = classified.Type
			info.Permanent = classified.Permanent
		}
	}
}

// containsAny checks if s contains any of the given substrings.
func containsAny(s string, substrs ...string) bool {
	for _, sub := range substrs {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}
