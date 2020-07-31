package sdk

import "time"

// Control is an interface for notifying of control states
type Control interface {
	// Paused must be called when the integration is paused for any reason such as rate limiting
	Paused(resetAt time.Time) error
	// Resumed must be called when a paused integration is resumed
	Resumed() error
	// CustomerID will return the customer id for this instance
	CustomerID() string
	// IntegrationInstanceID will return the unique instance id for this integration for a customer
	IntegrationInstanceID() string
	// RefType for the integration
	RefType() string
}
