package eventapi

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/pinpt/agent/v4/sdk"
	"github.com/pinpt/go-common/v10/api"
	"github.com/pinpt/go-common/v10/datetime"
	gql "github.com/pinpt/go-common/v10/graphql"
	"github.com/pinpt/go-common/v10/log"
	"github.com/pinpt/integration-sdk/agent"
)

type autoconfig struct {
	ctx                   context.Context
	logger                log.Logger
	config                sdk.Config
	state                 sdk.State
	customerID            string
	integrationInstanceID string
	refType               string
	channel               string
	apikey                string
	secret                string
	pipe                  sdk.Pipe
	paused                bool
	mu                    sync.Mutex
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

// Logger the logger object to use in the integration
func (e *autoconfig) Logger() sdk.Logger {
	return e.logger
}

func (e *autoconfig) createGraphql() gql.Client {
	url := api.BackendURL(api.GraphService, e.channel)
	client, err := gql.NewClient(e.customerID, "", e.secret, url)
	if err != nil {
		panic(err)
	}
	if e.apikey != "" {
		client.SetHeader("Authorization", e.apikey)
	}
	return client
}

func (e *autoconfig) updateIntegration(vars gql.Variables) error {
	// update the db with our new integration state
	return agent.ExecIntegrationInstanceSilentUpdateMutation(e.createGraphql(), e.integrationInstanceID, vars, false)
}

// Paused must be called when the integration is paused for any reason such as rate limiting
func (e *autoconfig) Paused(resetAt time.Time) error {
	e.mu.Lock()
	if e.paused {
		e.mu.Unlock()
		return nil
	}
	e.paused = true
	e.mu.Unlock()
	e.pipe.Flush() // flush the pipe once we're paused to go ahead and send any pending data
	log.Info(e.logger, "paused", "reset", resetAt, "duration", time.Until(resetAt))
	var dt agent.IntegrationInstanceThrottledUntil
	sdk.ConvertTimeToDateModel(resetAt, &dt)
	return e.updateIntegration(gql.Variables{
		agent.IntegrationInstanceModelThrottledColumn:      true,
		agent.IntegrationInstanceModelThrottledUntilColumn: dt,
		agent.IntegrationInstanceModelUpdatedAtColumn:      datetime.EpochNow(),
	})
}

// Resumed must be called when a paused integration is resumed
func (e *autoconfig) Resumed() error {
	e.mu.Lock()
	if !e.paused {
		e.mu.Unlock()
		return nil
	}
	e.paused = false
	e.mu.Unlock()
	log.Info(e.logger, "pause resumed")
	var dt agent.IntegrationInstanceThrottledUntil
	return e.updateIntegration(gql.Variables{
		agent.IntegrationInstanceModelThrottledColumn:      false,
		agent.IntegrationInstanceModelThrottledUntilColumn: dt,
		agent.IntegrationInstanceModelUpdatedAtColumn:      datetime.EpochNow(),
	})
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
	Channel               string
	APIKey                string
	Secret                string
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
		channel:               config.Channel,
		apikey:                config.APIKey,
		secret:                config.Secret,
	}, nil
}
