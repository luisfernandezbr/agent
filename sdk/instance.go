package sdk

// Instance is an instance of an integration for a specific customer and integration instance
type Instance struct {
	config                Config
	state                 State
	customerID            string
	integrationInstanceID string
	pipe                  Pipe
}

// Config is a customer specific config object for this integration and customer
func (i *Instance) Config() Config {
	return i.config
}

// State is a customer specific state object for this integration and customer
func (i *Instance) State() State {
	return i.state
}

// CustomerID will return the customer id for the export
func (i *Instance) CustomerID() string {
	return i.customerID
}

// IntegrationInstanceID will return the unique instance id for this integration for a customer
func (i *Instance) IntegrationInstanceID() string {
	return i.integrationInstanceID
}

// Pipe returns a pipe in the case the integration wants to send data back to pinpoint
func (i *Instance) Pipe() Pipe {
	return i.pipe
}

// NewInstance returns a new instance of the integration
func NewInstance(config Config, state State, pipe Pipe, customerID string, integrationInstanceID string) *Instance {
	return &Instance{
		config:                config,
		state:                 state,
		pipe:                  pipe,
		customerID:            customerID,
		integrationInstanceID: integrationInstanceID,
	}
}
