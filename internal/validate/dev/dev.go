package dev

import (
	"time"

	"github.com/pinpt/agent/v4/sdk"
)

// if this gets complicated and needs a pipe or something make a dev/eventapi implementation
type validate struct {
	config                sdk.Config
	logger                sdk.Logger
	integrationInstanceID string
	customerID            string
	refType               string
}

func (v *validate) State() sdk.State {
	return nil
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

// Paused must be called when the integration is paused for any reason such as rate limiting
func (v *validate) Paused(resetAt time.Time) error {
	return nil
}

// Resumed must be called when a paused integration is resumed
func (v *validate) Resumed() error {
	return nil
}

// Logger the logger object to use in the integration
func (e *validate) Logger() sdk.Logger {
	return e.logger
}

// NewValidate will return a validate
func NewValidate(config sdk.Config, logger sdk.Logger, refType string, customerID string, integrationInstanceID string) sdk.Validate {
	return &validate{
		customerID:            customerID,
		refType:               refType,
		integrationInstanceID: integrationInstanceID,
		config:                config,
		logger:                logger,
	}
}
