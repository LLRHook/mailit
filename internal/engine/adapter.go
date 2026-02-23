package engine

import (
	"context"

	"github.com/mailit-dev/mailit/internal/worker"
)

// WorkerAdapter adapts the engine.Sender to the worker.EmailSender interface.
type WorkerAdapter struct {
	sender *Sender
}

// NewWorkerAdapter creates a new WorkerAdapter wrapping the given Sender.
func NewWorkerAdapter(s *Sender) *WorkerAdapter {
	return &WorkerAdapter{sender: s}
}

// SendEmail converts a worker.OutboundMessage to an engine.OutgoingMessage,
// calls the engine sender, and converts the result back.
func (a *WorkerAdapter) SendEmail(ctx context.Context, msg *worker.OutboundMessage) ([]worker.RecipientResult, error) {
	outgoing := &OutgoingMessage{
		From:         msg.From,
		To:           msg.To,
		Cc:           msg.Cc,
		Bcc:          msg.Bcc,
		ReplyTo:      msg.ReplyTo,
		Subject:      msg.Subject,
		HTMLBody:     msg.HTMLBody,
		TextBody:     msg.TextBody,
		Headers:      msg.Headers,
		MessageID:    msg.MessageID,
		DKIMDomain:   msg.DKIMDomain,
		DKIMSelector: msg.DKIMSelector,
		DKIMKey:      string(msg.DKIMKey),
	}

	result, err := a.sender.SendEmail(ctx, outgoing)
	if err != nil {
		return nil, err
	}

	var results []worker.RecipientResult
	for recipient, r := range result.Recipients {
		results = append(results, worker.RecipientResult{
			Recipient: recipient,
			Success:   r.Status == "sent",
			Code:      r.Code,
			Message:   r.Message,
			Permanent: r.Permanent,
		})
	}

	return results, nil
}
