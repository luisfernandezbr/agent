package sdk

// WebHook is a control interafce for web hook data received by pinpoint on behalf of the integration
type WebHook interface {
	// Config is any customer specific configuration for this customer
	Config() Config
	// CustomerID will return the customer id for the web hook
	CustomerID() string
	// RefID will return the ref id from when the hook was created
	RefID() string
	// Pipe returns a pipe for sending data back to pinpoint from the web hook data
	Pipe() (Pipe, error)
	// Data is the data payload for the web hook
	Data() map[string]interface{}
	// Headers are the headers that came from the web hook
	Headers() map[string]string
}
