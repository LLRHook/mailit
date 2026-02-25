package mock

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/mailit-dev/mailit/internal/worker"
)

// MockEmailSender mocks the worker.EmailSender interface.
type MockEmailSender struct{ mock.Mock }

func (m *MockEmailSender) SendEmail(ctx context.Context, msg *worker.OutboundMessage) ([]worker.RecipientResult, error) {
	args := m.Called(ctx, msg)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]worker.RecipientResult), args.Error(1)
}
