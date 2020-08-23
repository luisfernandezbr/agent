package sdk

// AutoConfigure is the control interface when auto configure is called
type AutoConfigure interface {
	Control
	// Config is any customer specific configuration for this customer
	Config() Config
	// State is a customer specific state object for this integration and customer
	State() State
	// Pipe should be called to get the pipe for streaming data back to pinpoint
	Pipe() Pipe
}
