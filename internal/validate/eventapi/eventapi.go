package eventapi

import (
	"sync"
	"time"

	"github.com/pinpt/agent/v4/sdk"
	"github.com/pinpt/go-common/v10/datetime"
	"github.com/pinpt/go-common/v10/graphql"
	"github.com/pinpt/go-common/v10/log"
	"github.com/pinpt/integration-sdk/agent"
)

// if this gets complicated and needs a pipe or something make a dev/eventapi implementation
type validate struct {
	logger                log.Logger
	config                sdk.Config
	integrationInstanceID string
	customerID            string
	refType               string
	client                graphql.Client
	mu                    sync.Mutex
	paused                bool
	state                 sdk.State
}

func (v *validate) State() sdk.State {
	return v.state
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

// Logger the logger object to use in the integration
func (e *validate) Logger() sdk.Logger {
	return e.logger
}

// Paused must be called when the integration is paused for any reason such as rate limiting
func (v *validate) Paused(resetAt time.Time) error {
	v.mu.Lock()
	if v.paused {
		v.mu.Unlock()
		return nil
	}
	v.paused = true
	v.mu.Unlock()
	log.Info(v.logger, "paused", "reset", resetAt, "duration", time.Until(resetAt))
	var dt agent.IntegrationInstanceThrottledUntil
	sdk.ConvertTimeToDateModel(resetAt, &dt)
	return agent.ExecIntegrationInstanceSilentUpdateMutation(v.client, v.integrationInstanceID, graphql.Variables{
		agent.IntegrationInstanceModelThrottledColumn:      true,
		agent.IntegrationInstanceModelThrottledUntilColumn: dt,
		agent.IntegrationInstanceModelUpdatedAtColumn:      datetime.EpochNow(),
	}, false)
}

// Resumed must be called when a paused integration is resumed
func (v *validate) Resumed() error {
	v.mu.Lock()
	if !v.paused {
		v.mu.Unlock()
		return nil
	}
	v.paused = false
	v.mu.Unlock()
	log.Info(v.logger, "pause resumed")
	var dt agent.IntegrationInstanceThrottledUntil
	return agent.ExecIntegrationInstanceSilentUpdateMutation(v.client, v.integrationInstanceID, graphql.Variables{
		agent.IntegrationInstanceModelThrottledColumn:      false,
		agent.IntegrationInstanceModelThrottledUntilColumn: dt,
		agent.IntegrationInstanceModelUpdatedAtColumn:      datetime.EpochNow(),
	}, false)
}

// NewValidate will return a validate
func NewValidate(config sdk.Config, logger log.Logger, refType string, customerID string, integrationInstanceID string, client graphql.Client, state sdk.State) sdk.Validate {
	return &validate{
		customerID:            customerID,
		refType:               refType,
		integrationInstanceID: integrationInstanceID,
		config:                config,
		client:                client,
		state:                 state,
		logger:                logger,
	}
}
