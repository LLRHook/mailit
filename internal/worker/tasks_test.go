package worker

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewEmailSendTask(t *testing.T) {
	emailID := uuid.New()
	teamID := uuid.New()

	task, err := NewEmailSendTask(emailID, teamID)
	require.NoError(t, err)
	require.NotNil(t, task)

	assert.Equal(t, TaskEmailSend, task.Type())

	var payload EmailSendPayload
	err = json.Unmarshal(task.Payload(), &payload)
	require.NoError(t, err)
	assert.Equal(t, emailID, payload.EmailID)
	assert.Equal(t, teamID, payload.TeamID)
}

func TestNewEmailBatchSendTask(t *testing.T) {
	emailIDs := []uuid.UUID{uuid.New(), uuid.New(), uuid.New()}
	teamID := uuid.New()

	task, err := NewEmailBatchSendTask(emailIDs, teamID)
	require.NoError(t, err)
	require.NotNil(t, task)

	assert.Equal(t, TaskEmailBatchSend, task.Type())

	var payload EmailBatchSendPayload
	err = json.Unmarshal(task.Payload(), &payload)
	require.NoError(t, err)
	assert.Equal(t, emailIDs, payload.EmailIDs)
	assert.Equal(t, teamID, payload.TeamID)
}

func TestNewBroadcastSendTask(t *testing.T) {
	broadcastID := uuid.New()
	teamID := uuid.New()

	task, err := NewBroadcastSendTask(broadcastID, teamID)
	require.NoError(t, err)
	require.NotNil(t, task)

	assert.Equal(t, TaskBroadcastSend, task.Type())

	var payload BroadcastSendPayload
	err = json.Unmarshal(task.Payload(), &payload)
	require.NoError(t, err)
	assert.Equal(t, broadcastID, payload.BroadcastID)
	assert.Equal(t, teamID, payload.TeamID)
}

func TestNewDomainVerifyTask(t *testing.T) {
	domainID := uuid.New()
	teamID := uuid.New()

	task, err := NewDomainVerifyTask(domainID, teamID)
	require.NoError(t, err)
	require.NotNil(t, task)

	assert.Equal(t, TaskDomainVerify, task.Type())

	var payload DomainVerifyPayload
	err = json.Unmarshal(task.Payload(), &payload)
	require.NoError(t, err)
	assert.Equal(t, domainID, payload.DomainID)
	assert.Equal(t, teamID, payload.TeamID)
}

func TestNewWebhookDeliverTask(t *testing.T) {
	webhookEventID := uuid.New()

	task, err := NewWebhookDeliverTask(webhookEventID)
	require.NoError(t, err)
	require.NotNil(t, task)

	assert.Equal(t, TaskWebhookDeliver, task.Type())

	var payload WebhookDeliverPayload
	err = json.Unmarshal(task.Payload(), &payload)
	require.NoError(t, err)
	assert.Equal(t, webhookEventID, payload.WebhookEventID)
}

func TestNewBounceProcessTask(t *testing.T) {
	emailID := uuid.New()
	code := 550
	message := "User unknown"
	recipient := "bounce@example.com"

	task, err := NewBounceProcessTask(emailID, code, message, recipient)
	require.NoError(t, err)
	require.NotNil(t, task)

	assert.Equal(t, TaskBounceProcess, task.Type())

	var payload BounceProcessPayload
	err = json.Unmarshal(task.Payload(), &payload)
	require.NoError(t, err)
	assert.Equal(t, emailID, payload.EmailID)
	assert.Equal(t, code, payload.Code)
	assert.Equal(t, message, payload.Message)
	assert.Equal(t, recipient, payload.Recipient)
}

func TestNewInboundProcessTask(t *testing.T) {
	inboundEmailID := uuid.New()

	task, err := NewInboundProcessTask(inboundEmailID)
	require.NoError(t, err)
	require.NotNil(t, task)

	assert.Equal(t, TaskInboundProcess, task.Type())

	var payload InboundProcessPayload
	err = json.Unmarshal(task.Payload(), &payload)
	require.NoError(t, err)
	assert.Equal(t, inboundEmailID, payload.InboundEmailID)
}

func TestNewCleanupExpiredTask(t *testing.T) {
	task, err := NewCleanupExpiredTask()
	require.NoError(t, err)
	require.NotNil(t, task)

	assert.Equal(t, TaskCleanupExpired, task.Type())
	// Cleanup task has nil payload.
	assert.Nil(t, task.Payload())
}

func TestTaskTypeConstants(t *testing.T) {
	// Verify task type constants are unique and well-formed.
	types := []string{
		TaskEmailSend,
		TaskEmailBatchSend,
		TaskBroadcastSend,
		TaskDomainVerify,
		TaskWebhookDeliver,
		TaskBounceProcess,
		TaskInboundProcess,
		TaskCleanupExpired,
	}

	seen := make(map[string]bool)
	for _, tt := range types {
		assert.NotEmpty(t, tt, "task type should not be empty")
		assert.False(t, seen[tt], "duplicate task type: %s", tt)
		seen[tt] = true
	}
}

func TestQueueConstants(t *testing.T) {
	assert.Equal(t, "critical", QueueCritical)
	assert.Equal(t, "default", QueueDefault)
	assert.Equal(t, "low", QueueLow)
}

func TestPayloadMarshalUnmarshal_Roundtrip(t *testing.T) {
	t.Run("EmailSendPayload roundtrip", func(t *testing.T) {
		original := EmailSendPayload{
			EmailID: uuid.New(),
			TeamID:  uuid.New(),
		}
		data, err := json.Marshal(original)
		require.NoError(t, err)

		var decoded EmailSendPayload
		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err)
		assert.Equal(t, original, decoded)
	})

	t.Run("BroadcastSendPayload roundtrip", func(t *testing.T) {
		original := BroadcastSendPayload{
			BroadcastID: uuid.New(),
			TeamID:      uuid.New(),
		}
		data, err := json.Marshal(original)
		require.NoError(t, err)

		var decoded BroadcastSendPayload
		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err)
		assert.Equal(t, original, decoded)
	})

	t.Run("DomainVerifyPayload roundtrip", func(t *testing.T) {
		original := DomainVerifyPayload{
			DomainID: uuid.New(),
			TeamID:   uuid.New(),
		}
		data, err := json.Marshal(original)
		require.NoError(t, err)

		var decoded DomainVerifyPayload
		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err)
		assert.Equal(t, original, decoded)
	})

	t.Run("WebhookDeliverPayload roundtrip", func(t *testing.T) {
		original := WebhookDeliverPayload{
			WebhookEventID: uuid.New(),
		}
		data, err := json.Marshal(original)
		require.NoError(t, err)

		var decoded WebhookDeliverPayload
		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err)
		assert.Equal(t, original, decoded)
	})

	t.Run("BounceProcessPayload roundtrip", func(t *testing.T) {
		original := BounceProcessPayload{
			EmailID:   uuid.New(),
			Code:      552,
			Message:   "Mailbox full",
			Recipient: "user@example.com",
		}
		data, err := json.Marshal(original)
		require.NoError(t, err)

		var decoded BounceProcessPayload
		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err)
		assert.Equal(t, original, decoded)
	})

	t.Run("EmailBatchSendPayload roundtrip", func(t *testing.T) {
		original := EmailBatchSendPayload{
			EmailIDs: []uuid.UUID{uuid.New(), uuid.New()},
			TeamID:   uuid.New(),
		}
		data, err := json.Marshal(original)
		require.NoError(t, err)

		var decoded EmailBatchSendPayload
		err = json.Unmarshal(data, &decoded)
		require.NoError(t, err)
		assert.Equal(t, original, decoded)
	})
}
