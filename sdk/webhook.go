package sdk

// WebHook is a control interafce for web hook data received by pinpoint on behalf of the integration
type WebHook interface {
	Control
	// Config is any customer specific configuration for this customer
	Config() Config
	// State is a customer specific state object for this integration and customer
	State() State
	// CustomerID will return the customer id for the web hook
	CustomerID() string
	// IntegrationInstanceID will return the unique instance id for this integration for a customer
	IntegrationInstanceID() string
	// RefID will return the ref id from when the hook was created
	RefID() string
	// Pipe returns a pipe for sending data back to pinpoint from the web hook data
	Pipe() Pipe
	// Data returns the payload of a webhook decoded from json into a map
	Data() (map[string]interface{}, error)
	// Bytes will return the underlying data as bytes
	Bytes() []byte
	// Headers are the headers that came from the web hook
	Headers() map[string]string
}
