package sdk

import "time"

// Export is a control interface for an export
type Export interface {
	// Config is any customer specific configuration for this customer
	Config() Config
	// State is a customer specific state object for this integration and customer
	State() State
	// JobID will return a specific job id for this export which can be used in logs, etc
	JobID() string
	// CustomerID will return the customer id for the export
	CustomerID() string
	// IntegrationID will return the unique instance id for this integration for a customer
	IntegrationID() string
	// Pipe should be called to get the pipe for streaming data back to pinpoint
	Pipe() (Pipe, error)
	// Paused must be called when the integration is paused for any reason such as rate limiting
	Paused(resetAt time.Time) error
	// Resumed must be called when a paused integration is resumed
	Resumed() error
	// Historical if true, the integration should perform a full historical export
	Historical() bool
}
