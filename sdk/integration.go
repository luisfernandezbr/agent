package sdk

// Integration is the interface that integrations implement
type Integration interface {
	// Start is called when the integration is starting up
	Start(logger Logger, config Config, manager Manager) error
	// Export is called to tell the integration to run an export
	Export(export Export) error
	// WebHook is called when a webhook is received on behalf of the integration
	WebHook(webhook WebHook) error
	// Stop is called when the integration is shutting down for cleanup
	Stop() error
}
