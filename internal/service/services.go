package service

// Services aggregates all service implementations.
type Services struct {
	Auth            AuthService
	Email           EmailService
	Domain          DomainService
	APIKey          APIKeyService
	Audience        AudienceService
	Contact         ContactService
	ContactProperty ContactPropertyService
	Topic           TopicService
	Segment         SegmentService
	Template        TemplateService
	Broadcast       BroadcastService
	Webhook         WebhookService
	InboundEmail    InboundEmailService
	Log             LogService
	Metrics         MetricsService
	Settings        SettingsService
}
