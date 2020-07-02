package eventapi

import (
	"context"

	"github.com/pinpt/agent.next/sdk"
	"github.com/pinpt/go-common/v10/log"
)

// Completion event
type webhook struct {
	ctx                   context.Context
	logger                log.Logger
	config                sdk.Config
	state                 sdk.State
	customerID            string
	integrationInstanceID string
	refID                 string
	pipe                  sdk.Pipe
	headers               map[string]string
	data                  map[string]interface{}
}

var _ sdk.WebHook = (*webhook)(nil)

// Config is any customer specific configuration for this customer
func (e *webhook) Config() sdk.Config {
	return e.config
}

// State is any customer specific state for this customer
func (e *webhook) State() sdk.State {
	return e.state
}

// CustomerID will return the customer id for the webhook
func (e *webhook) CustomerID() string {
	return e.customerID
}

// IntegrationID will return the unique instance id for this integration for a customer
func (e *webhook) IntegrationID() string {
	return e.integrationInstanceID
}

// RefID will return the ref id from when the hook was created
func (e *webhook) RefID() string {
	return e.refID
}

//  Pipe should be called to get the pipe for streaming data back to pinpoint
func (e *webhook) Pipe() sdk.Pipe {
	return e.pipe
}

// Paused must be called when the integration is paused for any reason such as rate limiting
func (e *webhook) Data() map[string]interface{} {
	return e.data
}

// Resumed must be called when a paused integration is resumed
func (e *webhook) Headers() map[string]string {
	return e.headers
}

// Config is details for the configuration
type Config struct {
	Ctx                   context.Context
	Logger                log.Logger
	Config                sdk.Config
	State                 sdk.State
	CustomerID            string
	RefID                 string
	IntegrationInstanceID string
	Pipe                  sdk.Pipe
	Data                  map[string]interface{}
	Headers               map[string]string
}

// New will return an sdk.WebHook
func New(config Config) sdk.WebHook {
	ctx := config.Ctx
	if ctx == nil {
		ctx = context.Background()
	}
	return &webhook{
		ctx:                   ctx,
		logger:                config.Logger,
		config:                config.Config,
		state:                 config.State,
		customerID:            config.CustomerID,
		refID:                 config.RefID,
		integrationInstanceID: config.IntegrationInstanceID,
		pipe:                  config.Pipe,
		headers:               config.Headers,
		data:                  config.Data,
	}
}
