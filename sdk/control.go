package sdk

import "time"

// Control is an interface for notifying of control states
type Control interface {
	Identifier
	// Paused must be called when the integration is paused for any reason such as rate limiting
	Paused(resetAt time.Time) error
	// Resumed must be called when a paused integration is resumed
	Resumed() error
}

// Identifier is an interface for getting the ids of the current execution
type Identifier interface {
	// CustomerID will return the customer id for this instance
	CustomerID() string
	// IntegrationInstanceID will return the unique instance id for this integration for a customer
	IntegrationInstanceID() string
	// RefType for the integration
	RefType() string
}

type simpleIdentifier struct {
	customerID            string
	integrationInstanceID string
	refType               string
}

func (i *simpleIdentifier) CustomerID() string            { return i.customerID }
func (i *simpleIdentifier) IntegrationInstanceID() string { return i.integrationInstanceID }
func (i *simpleIdentifier) RefType() string               { return i.refType }

// NewSimpleIdentifier will return an identifier, should only be used in rare occasions
// for calling sdk apis that require an Identifier but are outside the scope of a
// webhook, mutation, enroll, dismiss, or validate since all those implement Identifier
func NewSimpleIdentifier(customerID string, integrationInstanceID string, refType string) Identifier {
	return &simpleIdentifier{
		customerID:            customerID,
		integrationInstanceID: integrationInstanceID,
		refType:               refType,
	}
}
