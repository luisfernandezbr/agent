package server

import (
	"context"
	"encoding/json"
	"fmt"
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
	"github.com/pinpt/go-common/v10/event"
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

func (s *Server) handleAddIntegration(logger log.Logger, req agent.IntegrationRequest) error {
	if s.config.Integration.Descriptor.RefType == req.Integration.RefType {
		state, err := s.newState(req.CustomerID, req.Integration.ID)
		if err != nil {
			return err
		}
		dir := s.newTempDir("")
		pipe := s.newPipe(logger, dir, req.CustomerID, "", req.Integration.ID)
		defer func() {
			pipe.Close()
			os.RemoveAll(dir)
		}()
		config := s.newConfig(req.Integration.Config, req.Integration.Authorization.ToMap())
		instance := sdk.NewInstance(config, state, pipe, req.CustomerID, req.Integration.ID)
		log.Info(logger, "running add integration")
		if err := s.config.Integration.Integration.Enroll(*instance); err != nil {
			return err
		}
	}
	return nil
}

func (s *Server) handleExport(logger log.Logger, req agent.ExportRequest) error {
	dir := s.newTempDir(req.JobID)
	defer os.RemoveAll(dir)
	started := time.Now()
	var found bool
	var integration agent.ExportRequestIntegrations
	var integrationAuth agent.ExportRequestIntegrationsAuthorization
	// run our integrations in parallel
	for _, i := range req.Integrations {
		if i.Name == s.config.Integration.Descriptor.RefType {
			found = true
			integration = i
			integrationAuth = i.Authorization
			break
		}
	}
	if !found {
		return fmt.Errorf("incorrect agent.ExportRequest event, didn't match our integration")
	}
	sdkconfig := s.newConfig(integration.Config, integrationAuth.ToMap())
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

func (s *Server) onDBChange(evt event.SubscriptionEvent) error {
	ch, err := createDBChangeEvent(evt.Data)
	if err != nil {
		return err
	}
	// FIXME: handle db change
	log.Debug(s.logger, "received db change", ch)
	evt.Commit()
	return nil
}

func (s *Server) onEvent(evt event.SubscriptionEvent) error {
	log.Debug(s.logger, "received event", "evt", evt)
	switch evt.Model {
	case agent.ExportRequestModelName.String():
		var req agent.ExportRequest
		if err := json.Unmarshal([]byte(evt.Data), &req); err != nil {
			log.Fatal(s.logger, "error parsing export request event", "err", err)
		}
		// var errmessage *string
		err := s.handleExport(s.logger, req)
		if err != nil {
			log.Error(s.logger, "error running export request", "err", err)
			// errmessage = sdk.StringPointer(err.Error())
		}
		// FIXME: update the db
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
		if err := server.handleExport(config.Logger, agent.ExportRequest{
			JobID:      config.DevExport.JobID(),
			CustomerID: config.DevExport.CustomerID(),
			Integrations: []agent.ExportRequestIntegrations{
				agent.ExportRequestIntegrations{
					ID:   "1",
					Name: config.Integration.Descriptor.RefType,
				},
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
	server.event, err = NewEventSubscriber(config, []string{agent.ExportRequestModelName.String()}, server.onEvent)
	if err != nil {
		return nil, err
	}
	return server, nil
}
