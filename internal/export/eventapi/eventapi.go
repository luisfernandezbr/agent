package eventapi

import (
	"context"
	"sync"
	"time"

	"github.com/pinpt/agent.next/sdk"
	"github.com/pinpt/go-common/v10/log"
)

// Completion event
type Completion struct {
	Error error
}

type export struct {
	ctx        context.Context
	logger     log.Logger
	config     sdk.Config
	state      sdk.State
	customerID string
	jobID      string
	uuid       string
	channel    string
	apikey     string
	secret     string
	pipe       sdk.Pipe
	completion chan Completion
	paused     bool
	mu         sync.Mutex
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
	// TODO: send agent.ExportResponse with start progress
	return e.pipe, nil
}

// Paused must be called when the integration is paused for any reason such as rate limiting
func (e *export) Paused(resetAt time.Time) error {
	e.mu.Lock()
	if e.paused {
		e.mu.Unlock()
		return nil
	}
	e.paused = true
	e.mu.Unlock()
	e.pipe.Flush() // flush the pipe once we're paused to go ahead and send any pending data
	log.Info(e.logger, "paused", "reset", resetAt, "duration", time.Until(resetAt))
	// FIXME: send agent.Pause
	return nil
}

// Resumed must be called when a paused integration is resumed
func (e *export) Resumed() error {
	e.mu.Lock()
	if !e.paused {
		e.mu.Unlock()
		return nil
	}
	e.paused = false
	e.mu.Unlock()
	log.Info(e.logger, "pause resumed")
	// FIXME: send agent.Resume
	return nil
}

// Completed must be called when an export is completed and can include an optional error or nil if no error
func (e *export) Completed(err error) {
	// FIXME: send agent.ExportResponse
	e.completion <- Completion{err}
}

// Config is details for the configuration
type Config struct {
	Ctx        context.Context
	Logger     log.Logger
	Config     sdk.Config
	State      sdk.State
	CustomerID string
	JobID      string
	UUID       string
	Pipe       sdk.Pipe
	Completion chan Completion
	Channel    string
	APIKey     string
	Secret     string
}

// New will return an sdk.Export
func New(config Config) (sdk.Export, error) {
	ctx := config.Ctx
	if ctx == nil {
		ctx = context.Background()
	}
	return &export{
		ctx:        ctx,
		logger:     config.Logger,
		config:     config.Config,
		state:      config.State,
		customerID: config.CustomerID,
		jobID:      config.JobID,
		uuid:       config.UUID,
		pipe:       config.Pipe,
		completion: config.Completion,
	}, nil
}
