package eventapi

import (
	"context"
	"encoding/json"
	"time"

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
	buf                   []byte
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

// IntegrationInstanceID will return the unique instance id for this integration for a customer
func (e *webhook) IntegrationInstanceID() string {
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

// Bytes will return the underlying data as bytes
func (e *webhook) Bytes() []byte {
	return e.buf
}

// Data returns the payload of a webhook decoded from json into a map
func (e *webhook) Data() (map[string]interface{}, error) {
	data := make(map[string]interface{})
	if err := json.Unmarshal(e.buf, &data); err != nil {
		return nil, err
	}
	return data, nil
}

// Headers are the headers that came from the web hook
func (e *webhook) Headers() map[string]string {
	return e.headers
}

// Paused must be called when the integration is paused for any reason such as rate limiting
func (e *webhook) Paused(resetAt time.Time) error {
	return nil
}

// Resumed must be called when a paused integration is resumed
func (e *webhook) Resumed() error {
	return nil
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
	Buf                   []byte
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
		buf:                   config.Buf,
		headers:               config.Headers,
	}
}
