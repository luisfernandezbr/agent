package dev

import (
	"time"

	"github.com/pinpt/agent.next/sdk"
	"github.com/pinpt/go-common/v10/log"
)

// Completion event
type Completion struct {
	Error error
}

type export struct {
	logger     log.Logger
	config     sdk.Config
	state      sdk.State
	jobID      string
	customerID string
	pipe       sdk.Pipe
	completion chan Completion
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

// Start must be called to begin an export and receive a pipe for sending data
func (e *export) Start() (sdk.Pipe, error) {
	return e.pipe, nil
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

// Completed must be called when an export is completed and can include an optional error or nil if no error
func (e *export) Completed(err error) {
	e.completion <- Completion{err}
}

// New will return an sdk.Export
func New(logger log.Logger, config sdk.Config, state sdk.State, jobID string, customerID string, pipe sdk.Pipe, completion chan Completion) (sdk.Export, error) {
	return &export{
		logger:     logger,
		config:     config,
		state:      state,
		jobID:      jobID,
		customerID: customerID,
		pipe:       pipe,
		completion: completion,
	}, nil
}
