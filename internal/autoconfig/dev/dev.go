package dev

import (
	"context"
	"fmt"
	"time"

	"github.com/pinpt/agent/v4/sdk"
	"github.com/pinpt/go-common/v10/log"
)

type autoconfig struct {
	ctx                   context.Context
	logger                log.Logger
	config                sdk.Config
	state                 sdk.State
	customerID            string
	integrationInstanceID string
	refType               string
	pipe                  sdk.Pipe
}

var _ sdk.AutoConfigure = (*autoconfig)(nil)

// Config is any customer specific configuration for this customer
func (e *autoconfig) Config() sdk.Config {
	return e.config
}

// State is any customer specific state for this customer
func (e *autoconfig) State() sdk.State {
	return e.state
}

// CustomerID will return the customer id for this instance
func (e *autoconfig) CustomerID() string {
	return e.customerID
}

// IntegrationInstanceID will return the unique instance id for this integration for a customer
func (e *autoconfig) IntegrationInstanceID() string {
	return e.integrationInstanceID
}

// RefType for the integration
func (e *autoconfig) RefType() string {
	return e.refType
}

//  Pipe should be called to get the pipe for streaming data back to pinpoint
func (e *autoconfig) Pipe() sdk.Pipe {
	return e.pipe
}

// Paused must be called when the integration is paused for any reason such as rate limiting
func (e *autoconfig) Paused(resetAt time.Time) error {
	return nil
}

// Resumed must be called when a paused integration is resumed
func (e *autoconfig) Resumed() error {
	return nil
}

// Config is details for the configuration
type Config struct {
	Ctx                   context.Context
	Logger                log.Logger
	Config                sdk.Config
	State                 sdk.State
	CustomerID            string
	IntegrationInstanceID string
	RefType               string
	Pipe                  sdk.Pipe
}

// New will return an sdk.AutoConfigure
func New(config Config) (sdk.AutoConfigure, error) {
	ctx := config.Ctx
	if ctx == nil {
		ctx = context.Background()
	}
	if config.RefType == "" {
		return nil, fmt.Errorf("missing RefType")
	}
	return &autoconfig{
		ctx:                   ctx,
		logger:                config.Logger,
		config:                config.Config,
		state:                 config.State,
		integrationInstanceID: config.IntegrationInstanceID,
		refType:               config.RefType,
		pipe:                  config.Pipe,
		customerID:            config.CustomerID,
	}, nil
}
