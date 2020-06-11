package server

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/pinpt/agent.next/internal/export/eventapi"
	pipe "github.com/pinpt/agent.next/internal/pipe/eventapi"
	redisState "github.com/pinpt/agent.next/internal/state/redis"
	"github.com/pinpt/agent.next/sdk"
	"github.com/pinpt/go-common/v10/api"
	"github.com/pinpt/go-common/v10/event"
	"github.com/pinpt/go-common/v10/graphql"
	"github.com/pinpt/go-common/v10/hash"
	"github.com/pinpt/go-common/v10/log"
	"github.com/pinpt/integration-sdk/agent"
)

// IntegrationContext is the details for each integration
type IntegrationContext struct {
	Integration sdk.Integration
	Descriptor  *sdk.Descriptor
}

// Config is the configuration for the server
type Config struct {
	Ctx         context.Context
	Dir         string // temp dir for files
	Logger      log.Logger
	State       sdk.State     // can be nil
	RedisClient *redis.Client // can be nil
	Integration *IntegrationContext
	UUID        string
	Channel     string
	APIKey      string
	Secret      string
	GroupID     string
	DevMode     bool
	DevPipe     sdk.Pipe
	DevExport   sdk.Export
}

// Server is the event loop server portion of the agent
type Server struct {
	logger   log.Logger
	config   Config
	dbchange *Subscriber
	event    *Subscriber
}

var _ io.Closer = (*Server)(nil)

// Close the server
func (s *Server) Close() error {
	if s.dbchange != nil {
		s.dbchange.Close()
		s.dbchange = nil
	}
	if s.event != nil {
		s.event.Close()
		s.event = nil
	}
	return nil
}

func (s *Server) newState(customerID string, integrationID string) (sdk.State, error) {
	state := s.config.State
	if state == nil {
		// if no state provided, we use redis state in this case
		st, err := redisState.New(s.config.Ctx, s.config.RedisClient, customerID+":"+s.config.Integration.Descriptor.RefType+":"+integrationID)
		if err != nil {
			return nil, err
		}
		state = st
	}
	return state, nil
}

func (s *Server) newPipe(logger sdk.Logger, dir string, customerID string, jobID string, integrationID string) sdk.Pipe {
	var p sdk.Pipe
	if s.config.DevMode {
		p = s.config.DevPipe
	} else {
		p = pipe.New(pipe.Config{
			Ctx:           s.config.Ctx,
			Logger:        logger,
			Dir:           dir,
			CustomerID:    customerID,
			UUID:          s.config.UUID,
			JobID:         jobID,
			IntegrationID: integrationID,
			Channel:       s.config.Channel,
			APIKey:        s.config.APIKey,
			Secret:        s.config.Secret,
			RefType:       s.config.Integration.Descriptor.RefType,
		})
	}
	return p
}

func (s *Server) newTempDir(jobID string) string {
	if jobID == "" {
		jobID = strconv.Itoa(int(time.Now().Unix()))
	}
	dir := filepath.Join(s.config.Dir, jobID)
	os.MkdirAll(dir, 0700)
	return dir
}

func (s *Server) newConfig(configstr *string, kv map[string]interface{}) sdk.Config {
	var sdkconfig sdk.Config
	if s.config.DevMode {
		sdkconfig = s.config.DevExport.Config()
	} else {
		if configstr != nil && *configstr != "" {
			json.Unmarshal([]byte(*configstr), &kv)
		}
		sdkconfig = sdk.NewConfig(kv)
	}
	return sdkconfig
}

type cleanupFunc func()

func (s *Server) toInstance(integration *agent.IntegrationInstance) (*sdk.Instance, cleanupFunc, error) {
	state, err := s.newState(integration.CustomerID, integration.ID)
	if err != nil {
		return nil, nil, err
	}
	dir := s.newTempDir("")
	pipe := s.newPipe(s.logger, dir, integration.CustomerID, "", integration.ID)
	cleanup := func() {
		pipe.Close()
		os.RemoveAll(dir)
	}
	config := s.newConfig(integration.Config, make(map[string]interface{}))
	instance := sdk.NewInstance(config, state, pipe, integration.CustomerID, integration.ID)
	return instance, cleanup, nil
}

func (s *Server) handleAddIntegration(integration *agent.IntegrationInstance) error {
	log.Info(s.logger, "running enroll integration", "id", integration.ID, "customer_id", integration.CustomerID)
	instance, cleanup, err := s.toInstance(integration)
	if err != nil {
		return err
	}
	defer cleanup()
	if err := s.config.Integration.Integration.Enroll(*instance); err != nil {
		return err
	}
	return nil
}

func (s *Server) handleRemoveIntegration(integration *agent.IntegrationInstance) error {
	instance, cleanup, err := s.toInstance(integration)
	if err != nil {
		return err
	}
	defer cleanup()
	log.Info(s.logger, "running dismiss integration", "id", integration.ID, "customer_id", integration.CustomerID)
	if err := s.config.Integration.Integration.Dismiss(*instance); err != nil {
		return err
	}
	return nil
}

func (s *Server) handleExport(logger log.Logger, req agent.Export) error {
	dir := s.newTempDir(req.JobID)
	defer os.RemoveAll(dir)
	started := time.Now()
	integration := req.Integration
	sdkconfig := s.newConfig(integration.Config, make(map[string]interface{}))
	state, err := s.newState(req.CustomerID, integration.ID)
	if err != nil {
		return err
	}
	p := s.newPipe(logger, dir, req.CustomerID, req.JobID, integration.ID)
	defer p.Close()
	var e sdk.Export
	if s.config.DevMode {
		e = s.config.DevExport
	} else {
		c, err := eventapi.New(eventapi.Config{
			Ctx:           s.config.Ctx,
			Logger:        logger,
			Config:        sdkconfig,
			State:         state,
			CustomerID:    req.CustomerID,
			JobID:         req.JobID,
			IntegrationID: integration.ID,
			UUID:          s.config.UUID,
			Pipe:          p,
			Channel:       s.config.Channel,
			APIKey:        s.config.APIKey,
			Secret:        s.config.Secret,
			Historical:    req.ReprocessHistorical,
		})
		if err != nil {
			return err
		}
		e = c
	}
	log.Info(logger, "running export")
	if err := s.config.Integration.Integration.Export(e); err != nil {
		return err
	}
	if err := state.Flush(); err != nil {
		log.Error(logger, "error flushing state", "err", err)
	}
	log.Info(logger, "export completed", "duration", time.Since(started), "jobid", req.JobID, "customer_id", req.CustomerID)
	return nil
}

func (s *Server) calculateIntegrationHashCode(integration *agent.IntegrationInstance) string {
	// since we get db change events each time we update the agent.integration table, we don't
	// want to thrash the integration with enrolls so we are going to check specific fields for changes
	// if any of these change, we don't need to send a enroll since the config is the same as before
	return hash.Values(
		integration.Config,
		integration.Active,
		integration.Interval,
	)
}

func (s *Server) onDBChange(evt event.SubscriptionEvent) error {
	ch, err := createDBChangeEvent(evt.Data)
	if err != nil {
		return err
	}
	log.Debug(s.logger, "received db change", ch)
	switch ch.Model {
	case agent.IntegrationModelName.String():
		// don't worry in dev mode
		if !s.config.DevMode {
			// integration has changed so we need to either enroll or dismiss
			if integration, ok := ch.Object.(*agent.IntegrationInstance); ok {
				cachekey := "agent:" + integration.CustomerID + ":" + integration.ID + ":hashcode"
				res, _ := s.config.RedisClient.Get(s.config.Ctx, cachekey).Result()
				hashcode := s.calculateIntegrationHashCode(integration)
				// check to see if this is a delete OR we've deactivated the integration
				if ch.Action == Delete || !integration.Active {
					// delete the integration cache key and then signal a removal
					s.config.RedisClient.Del(s.config.Ctx, cachekey)
					if err := s.handleRemoveIntegration(integration); err != nil {
						log.Error(s.logger, "error removing integration", "err", err, "id", integration.ID)
					}
				} else {
					if res != hashcode {
						// update our hash key and then signal an addition
						s.config.RedisClient.Set(s.config.Ctx, cachekey, hashcode, 0)
						if err := s.handleAddIntegration(integration); err != nil {
							log.Error(s.logger, "error adding integration", "err", err, "id", integration.ID)
						}
					}
				}
			}
		}
	}
	evt.Commit()
	return nil
}

func (s *Server) onEvent(evt event.SubscriptionEvent) error {
	log.Debug(s.logger, "received event", "evt", evt)
	switch evt.Model {
	case agent.ExportModelName.String():
		var req agent.Export
		if err := json.Unmarshal([]byte(evt.Data), &req); err != nil {
			log.Fatal(s.logger, "error parsing export event", "err", err)
		}
		var cl graphql.Client
		// don't worry in dev mode
		if !s.config.DevMode {
			var err error
			cl, err = graphql.NewClient(
				req.CustomerID,
				"",
				s.config.Secret,
				api.BackendURL(api.GraphService, s.config.Channel),
			)
			if err != nil {
				log.Error(s.logger, "error creating graphql client", "err',err")
			}
			if s.config.APIKey != "" {
				cl.SetHeader("Authorization", s.config.APIKey)
			}
			// update the integration state to acknowledge that we are exporting
			vars := make(graphql.Variables)
			vars[agent.IntegrationInstanceModelExportAcknowledgedColumn] = true
			// TODO(robin): add last export acknowledged date
			if _, err := agent.ExecIntegrationInstanceUpdateMutation(cl, req.Integration.ID, vars, false); err != nil {
				log.Error(s.logger, "error updating agent integration", "err", err, "id", req.Integration.ID)
			}
		}
		var errmessage *string
		err := s.handleExport(s.logger, req)
		if err != nil {
			log.Error(s.logger, "error running export request", "err", err)
			errmessage = sdk.StringPointer(err.Error())
		}
		// don't worry in dev mode
		if !s.config.DevMode {
			// update the db with our new integration state
			vars := make(graphql.Variables)
			vars[agent.IntegrationInstanceModelStateColumn] = agent.IntegrationStateIdle
			if errmessage != nil {
				vars[agent.IntegrationInstanceModelErrorMessageColumn] = *errmessage
			}
			var dt agent.IntegrationInstanceLastExportCompletedDate
			sdk.ConvertTimeToDateModel(time.Now(), &dt)
			vars[agent.IntegrationInstanceModelLastExportCompletedDateColumn] = dt
			if _, err := agent.ExecIntegrationInstanceUpdateMutation(cl, req.Integration.ID, vars, false); err != nil {
				log.Error(s.logger, "error updating agent integration", "err", err, "id", req.Integration.ID)
			}
		}
	}
	evt.Commit()
	return nil
}

// New returns a new server instance
func New(config Config) (*Server, error) {
	server := &Server{
		config: config,
		logger: log.With(config.Logger, "pkg", "server"),
	}
	if config.DevMode {
		if err := server.handleExport(config.Logger, agent.Export{
			JobID:      config.DevExport.JobID(),
			CustomerID: config.DevExport.CustomerID(),
			Integration: agent.ExportIntegration{
				ID: "1",
			},
		}); err != nil {
			return nil, err
		}
		return nil, nil
	}
	var err error
	server.dbchange, err = NewDBChangeSubscriber(config, server.onDBChange)
	if err != nil {
		return nil, err
	}
	server.event, err = NewEventSubscriber(config, []string{agent.ExportModelName.String()}, server.onEvent)
	if err != nil {
		return nil, err
	}
	return server, nil
}
