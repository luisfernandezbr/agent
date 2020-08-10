package validate

import "github.com/pinpt/agent.next/sdk"

// if this gets complicated and needs a pipe or something make a dev/eventapi implementation
type validate struct {
	config                sdk.Config
	integrationInstanceID string
	customerID            string
	refType               string
}

func (v *validate) IntegrationInstanceID() string {
	return v.integrationInstanceID
}

func (v *validate) CustomerID() string {
	return v.customerID
}

func (v *validate) Config() sdk.Config {
	return v.config
}

func (v *validate) RefType() string {
	return v.refType
}

// NewValidate will return a validate
func NewValidate(config sdk.Config, refType string, customerID string, integrationInstanceID string) sdk.Validate {
	return &validate{
		customerID:            customerID,
		refType:               refType,
		integrationInstanceID: integrationInstanceID,
		config:                config,
	}
}
