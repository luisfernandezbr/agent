package dev

import (
	"time"

	"github.com/pinpt/agent.next/sdk"
	"github.com/pinpt/go-common/v10/log"
)

type webhook struct {
	logger                log.Logger
	config                sdk.Config
	state                 sdk.State
	customerID            string
	integrationInstanceID string
	refID                 string
	url                   string
	pipe                  sdk.Pipe
	headers               map[string]string
	buf                   []byte
	data                  map[string]interface{}
	scope                 sdk.WebHookScope
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

// Data returns the payload of a webhook decoded from json into a map
func (e *webhook) Data() (map[string]interface{}, error) {
	return e.data, nil
}

// Bytes will return the underlying data as bytes
func (e *webhook) Bytes() []byte {
	return e.buf
}

// URL the webhook callback url
func (e *webhook) URL() string {
	return e.url
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

// Scope is the registered webhook scope
func (e *webhook) Scope() sdk.WebHookScope {
	return e.scope
}

// New will return an sdk.WebHook
func New(
	logger log.Logger,
	config sdk.Config,
	state sdk.State,
	customerID string,
	webhookURL string,
	refID string,
	integrationInstanceID string,
	pipe sdk.Pipe,
	headers map[string]string,
	data map[string]interface{},
	buf []byte,
	scope sdk.WebHookScope,
) sdk.WebHook {
	return &webhook{
		logger:                logger,
		config:                config,
		state:                 state,
		customerID:            customerID,
		refID:                 refID,
		integrationInstanceID: integrationInstanceID,
		pipe:                  pipe,
		headers:               headers,
		data:                  data,
		buf:                   buf,
		scope:                 scope,
		url:                   webhookURL,
	}
}
