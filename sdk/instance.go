package sdk

// Instance is an instance of an integration for a specific customer and integration instance
type Instance struct {
	config                Config
	state                 State
	customerID            string
	integrationInstanceID string
	refType               string
	pipe                  Pipe
	logger                Logger
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

// RefType returns the integration ref_type
func (i *Instance) RefType() string {
	return i.refType
}

// Logger the logger object to use in the integration
func (i *Instance) Logger() Logger {
	return i.logger
}

// NewInstance returns a new instance of the integration
func NewInstance(config Config, logger Logger, state State, pipe Pipe, customerID string, refType string, integrationInstanceID string) *Instance {
	return &Instance{
		config:                config,
		state:                 state,
		pipe:                  pipe,
		customerID:            customerID,
		refType:               refType,
		integrationInstanceID: integrationInstanceID,
		logger:                logger,
	}
}
