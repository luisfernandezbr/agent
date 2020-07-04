package dev

import (
	"time"

	"github.com/pinpt/agent.next/sdk"
	"github.com/pinpt/go-common/v10/log"
)

type mutation struct {
	logger                log.Logger
	config                sdk.Config
	state                 sdk.State
	customerID            string
	integrationInstanceID string
	refID                 string
	pipe                  sdk.Pipe
	id                    string
	model                 string
	action                sdk.MutationAction
	payload               interface{}
	user                  sdk.MutationUser
}

var _ sdk.Mutation = (*mutation)(nil)

// Config is any customer specific configuration for this customer
func (e *mutation) Config() sdk.Config {
	return e.config
}

// State is any customer specific state for this customer
func (e *mutation) State() sdk.State {
	return e.state
}

// CustomerID will return the customer id for the mutation
func (e *mutation) CustomerID() string {
	return e.customerID
}

// IntegrationInstanceID will return the unique instance id for this integration for a customer
func (e *mutation) IntegrationInstanceID() string {
	return e.integrationInstanceID
}

// RefID will return the ref id from when the hook was created
func (e *mutation) RefID() string {
	return e.refID
}

//  Pipe should be called to get the pipe for streaming data back to pinpoint
func (e *mutation) Pipe() sdk.Pipe {
	return e.pipe
}

// Paused must be called when the integration is paused for any reason such as rate limiting
func (e *mutation) Paused(resetAt time.Time) error {
	return nil
}

// Resumed must be called when a paused integration is resumed
func (e *mutation) Resumed() error {
	return nil
}

// ID is the primary key of the payload
func (e *mutation) ID() string {
	return e.id
}

// Model is the name of the model of the payload
func (e *mutation) Model() string {
	return e.model
}

// Action is the mutation action
func (e *mutation) Action() sdk.MutationAction {
	return e.action
}

// Payload is the payload of the mutation which can be either a sdk.Model for create, sdk.PartialModel for update or nil for delete
func (e *mutation) Payload() interface{} {
	return e.payload
}

// User is the user that is requesting the mutation and any authorization details that might be required
func (e *mutation) User() sdk.MutationUser {
	return e.user
}

// New will return an sdk.mutation
func New(
	logger log.Logger,
	config sdk.Config,
	state sdk.State,
	customerID string,
	refID string,
	integrationInstanceID string,
	pipe sdk.Pipe,
	id string,
	model string,
	action sdk.MutationAction,
	payload interface{},
	user sdk.MutationUser,
) sdk.Mutation {
	return &mutation{
		logger:                logger,
		config:                config,
		state:                 state,
		customerID:            customerID,
		refID:                 refID,
		integrationInstanceID: integrationInstanceID,
		pipe:                  pipe,
		user:                  user,
	}
}
