package dev

import (
	"time"

	"github.com/pinpt/agent/v4/sdk"
	"github.com/pinpt/go-common/v10/log"
)

type export struct {
	logger                log.Logger
	config                sdk.Config
	state                 sdk.State
	jobID                 string
	customerID            string
	integrationInstanceID string
	refType               string
	pipe                  sdk.Pipe
	historical            bool
	stats                 sdk.Stats
}

var _ sdk.Export = (*export)(nil)

// Config is any customer specific configuration for this customer
func (e *export) Config() sdk.Config {
	return e.config
}

// State is any customer specific state for this customer
func (e *export) State() sdk.State {
	return e.state
}

// JobID will return a specific job id for this export which can be used in logs, etc
func (e *export) JobID() string {
	return e.jobID
}

// CustomerID will return the customer id for the export
func (e *export) CustomerID() string {
	return e.customerID
}

// IntegrationInstanceID will return the unique instance id for this integration for a customer
func (e *export) IntegrationInstanceID() string {
	return e.integrationInstanceID
}

// RefType for the integration
func (e *export) RefType() string {
	return e.refType
}

// Stats is the stats object that an integration can use to track integration specific stats for the export
func (e *export) Stats() sdk.Stats {
	return e.stats
}

//  Pipe should be called to get the pipe for streaming data back to pinpoint
func (e *export) Pipe() sdk.Pipe {
	return e.pipe
}

// Paused must be called when the integration is paused for any reason such as rate limiting
func (e *export) Paused(resetAt time.Time) error {
	log.Info(e.logger, "paused", "reset", resetAt, "duration", time.Until(resetAt))
	return nil
}

// Resumed must be called when a paused integration is resumed
func (e *export) Resumed() error {
	log.Info(e.logger, "pause resumed")
	return nil
}

// Historical if true, the integration should perform a full historical export
func (e *export) Historical() bool {
	return e.historical
}

// New will return an sdk.Export
func New(logger log.Logger, config sdk.Config, state sdk.State, jobID string, customerID string, integrationInstanceID string, refType string, historical bool, pipe sdk.Pipe) (sdk.Export, error) {
	return &export{
		logger:                logger,
		config:                config,
		state:                 state,
		jobID:                 jobID,
		customerID:            customerID,
		refType:               refType,
		pipe:                  pipe,
		integrationInstanceID: integrationInstanceID,
		historical:            historical,
		stats:                 sdk.NewStats(),
	}, nil
}
