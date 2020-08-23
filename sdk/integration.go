package sdk

// Integration is the interface that integrations implement
type Integration interface {
	// Start is called when the integration is starting up
	Start(logger Logger, config Config, manager Manager) error
	// Validate is called before a new integration instance is added to determine
	// if the config is valid and the integration can properly communicate with the
	// source system. The result and the error will both be delivered to the App.
	// Returning a nil error is considered a successful validation.
	Validate(validate Validate) (result map[string]interface{}, err error)
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
	// AutoConfigure is called when a cloud integration has requested to be auto configured
	AutoConfigure(autoconfig AutoConfigure) (*Config, error)
	// Stop is called when the integration is shutting down for cleanup
	Stop() error
}
