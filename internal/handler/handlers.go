package handler

import "github.com/mailit-dev/mailit/internal/service"

// Handlers aggregates all HTTP handlers.
type Handlers struct {
	Auth            *AuthHandler
	Email           *EmailHandler
	Domain          *DomainHandler
	APIKey          *APIKeyHandler
	Audience        *AudienceHandler
	Contact         *ContactHandler
	ContactProperty *ContactPropertyHandler
	Topic           *TopicHandler
	Segment         *SegmentHandler
	Template        *TemplateHandler
	Broadcast       *BroadcastHandler
	Webhook         *WebhookHandler
	InboundEmail    *InboundEmailHandler
	Log             *LogHandler
	Metrics         *MetricsHandler
}

func NewHandlers(svc *service.Services) *Handlers {
	return &Handlers{
		Auth:            NewAuthHandler(svc.Auth),
		Email:           NewEmailHandler(svc.Email),
		Domain:          NewDomainHandler(svc.Domain),
		APIKey:          NewAPIKeyHandler(svc.APIKey),
		Audience:        NewAudienceHandler(svc.Audience),
		Contact:         NewContactHandler(svc.Contact),
		ContactProperty: NewContactPropertyHandler(svc.ContactProperty),
		Topic:           NewTopicHandler(svc.Topic),
		Segment:         NewSegmentHandler(svc.Segment),
		Template:        NewTemplateHandler(svc.Template),
		Broadcast:       NewBroadcastHandler(svc.Broadcast),
		Webhook:         NewWebhookHandler(svc.Webhook),
		InboundEmail:    NewInboundEmailHandler(svc.InboundEmail),
		Log:             NewLogHandler(svc.Log),
		Metrics:         NewMetricsHandler(svc.Metrics),
	}
}
