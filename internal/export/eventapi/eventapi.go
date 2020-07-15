package eventapi

import (
	"context"
	"sync"
	"time"

	"github.com/pinpt/agent.next/sdk"
	"github.com/pinpt/go-common/v10/api"
	"github.com/pinpt/go-common/v10/datetime"
	"github.com/pinpt/go-common/v10/event"
	gql "github.com/pinpt/go-common/v10/graphql"
	"github.com/pinpt/go-common/v10/log"
	"github.com/pinpt/integration-sdk/agent"
)

// Completion event
type export struct {
	ctx                   context.Context
	logger                log.Logger
	config                sdk.Config
	state                 sdk.State
	subscriptionChannel   *event.SubscriptionChannel
	customerID            string
	jobID                 string
	integrationInstanceID string
	uuid                  string
	channel               string
	apikey                string
	secret                string
	pipe                  sdk.Pipe
	paused                bool
	historical            bool
	mu                    sync.Mutex
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

//  Pipe should be called to get the pipe for streaming data back to pinpoint
func (e *export) Pipe() sdk.Pipe {
	return e.pipe
}

func (e *export) createGraphql() gql.Client {
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

func (e *export) updateIntegration(vars gql.Variables) error {
	// update the db with our new integration state
	cl := e.createGraphql()
	if err := agent.ExecIntegrationInstanceSilentUpdateMutation(cl, e.integrationInstanceID, vars, false); err != nil {
		return err
	}
	return nil
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
	var dt agent.IntegrationInstanceThrottledUntil
	sdk.ConvertTimeToDateModel(resetAt, &dt)
	return e.updateIntegration(gql.Variables{
		agent.IntegrationInstanceModelThrottledColumn:      true,
		agent.IntegrationInstanceModelThrottledUntilColumn: dt,
		agent.IntegrationInstanceModelUpdatedAtColumn:      datetime.EpochNow(),
	})
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
	var dt agent.IntegrationInstanceThrottledUntil
	return e.updateIntegration(gql.Variables{
		agent.IntegrationModelThrottledColumn:      false,
		agent.IntegrationModelThrottledUntilColumn: dt,
		agent.IntegrationModelUpdatedAtColumn:      datetime.EpochNow(),
	})
}

// Historical if true, the integration should perform a full historical export
func (e *export) Historical() bool {
	return e.historical
}

// Config is details for the configuration
type Config struct {
	Ctx                   context.Context
	Logger                log.Logger
	Config                sdk.Config
	State                 sdk.State
	SubscriptionChannel   *event.SubscriptionChannel
	CustomerID            string
	JobID                 string
	IntegrationInstanceID string
	UUID                  string
	Pipe                  sdk.Pipe
	Channel               string
	APIKey                string
	Secret                string
	Historical            bool
}

// New will return an sdk.Export
func New(config Config) (sdk.Export, error) {
	ctx := config.Ctx
	if ctx == nil {
		ctx = context.Background()
	}
	return &export{
		ctx:                   ctx,
		logger:                config.Logger,
		config:                config.Config,
		state:                 config.State,
		customerID:            config.CustomerID,
		jobID:                 config.JobID,
		integrationInstanceID: config.IntegrationInstanceID,
		uuid:                  config.UUID,
		pipe:                  config.Pipe,
		subscriptionChannel:   config.SubscriptionChannel,
		historical:            config.Historical,
	}, nil
}
