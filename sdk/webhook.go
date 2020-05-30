package sdk

// WebHook is a control interafce for web hook data received by pinpoint on behalf of the integration
type WebHook interface {
	// Config is any customer specific configuration for this customer
	Config() Config
	// CustomerID will return the customer id for the web hook
	CustomerID() string
	// Pipe returns a pipe for sending data back to pinpoint from the webhook data
	Pipe() (Pipe, error)
}
