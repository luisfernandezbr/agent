package sdk

// Integration is the interface that integrations implement
type Integration interface {
	// Start is called when the integration is starting up
	Start(logger Logger, config Config, manager Manager) error
	// Enroll is called when a new integration instance is added
	Enroll(instance Instance) error
	// Dismiss is called when an existing integration instance is removed
	Dismiss(instance Instance) error
	// Export is called to tell the integration to run an export
	Export(export Export) error
	// WebHook is called when a webhook is received on behalf of the integration
	WebHook(webhook WebHook) error
	// Mutation is called when a mutation request is received on behalf of the integration
	Mutation(mutation Mutation) error
	// Stop is called when the integration is shutting down for cleanup
	Stop() error
}
