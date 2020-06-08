package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
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
	Ctx                 context.Context
	Dir                 string // temp dir for files
	Logger              log.Logger
	State               sdk.State // can be nil
	SubscriptionChannel *event.SubscriptionChannel
	RedisClient         *redis.Client // can be nil
	Integration         *IntegrationContext
	UUID                string
	Channel             string
	APIKey              string
	Secret              string
	DevMode             bool
	DevPipe             sdk.Pipe
	DevExport           sdk.Export
}

// Server is the event loop server portion of the agent
type Server struct {
	config Config
}

var _ io.Closer = (*Server)(nil)

// Close the server
func (s *Server) Close() error {
	return nil
}

func (s *Server) handleExport(logger log.Logger, req agent.ExportRequest) error {
	dir := filepath.Join(s.config.Dir, req.JobID)
	os.MkdirAll(dir, 0700)
	started := time.Now()
	var found bool
	var integrationAuth agent.ExportRequestIntegrationsAuthorization
	// run our integrations in parallel
	for _, i := range req.Integrations {
		if i.Name == s.config.Integration.Descriptor.RefType {
			found = true
			integrationAuth = i.Authorization
			break
		}
	}
	if !found {
		return fmt.Errorf("incorrect agent.ExportRequest event, didn't match our integration")
	}
	// build the sdk config for the integration
	var sdkconfig sdk.Config
	if s.config.DevMode {
		sdkconfig = s.config.DevExport.Config()
	} else {
		sdkconfig = sdk.NewConfig(integrationAuth.ToMap())
	}
	state := s.config.State
	if state == nil {
		// if no state provided, we use redis state in this case
		st, err := redisState.New(s.config.Ctx, s.config.RedisClient, req.CustomerID)
		if err != nil {
			return err
		}
		state = st
	}
	// start the integration in it's own thread
	var p sdk.Pipe
	var e sdk.Export
	if s.config.DevMode {
		p = s.config.DevPipe
		e = s.config.DevExport
	} else {
		p = pipe.New(pipe.Config{
			Ctx:        s.config.Ctx,
			Logger:     logger,
			Dir:        dir,
			CustomerID: req.CustomerID,
			UUID:       s.config.UUID,
			JobID:      req.JobID,
			Channel:    s.config.Channel,
			APIKey:     s.config.APIKey,
			Secret:     s.config.Secret,
			RefType:    s.config.Integration.Descriptor.RefType,
		})
		c, err := eventapi.New(eventapi.Config{
			Ctx:        s.config.Ctx,
			Logger:     logger,
			Config:     sdkconfig,
			State:      state,
			CustomerID: req.CustomerID,
			JobID:      req.JobID,
			UUID:       s.config.UUID,
			Pipe:       p,
			Channel:    s.config.Channel,
			APIKey:     s.config.APIKey,
			Secret:     s.config.Secret,
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
	if err := p.Flush(); err != nil {
		log.Error(logger, "error flushing pipe", "err", err)
	}
	if err := state.Flush(); err != nil {
		log.Error(logger, "error flushing state", "err", err)
	}
	log.Info(logger, "export completed", "duration", time.Since(started), "jobid", req.JobID, "customer_id", req.CustomerID)
	return nil
}

func (s *Server) run() {
	logger := log.With(s.config.Logger, "pkg", "server")
	log.Debug(logger, "starting subscription channel")
	for evt := range s.config.SubscriptionChannel.Channel() {
		log.Debug(logger, "received event", "evt", evt)
		switch evt.Model {
		case agent.ExportRequestModelName.String():
			var req agent.ExportRequest
			if err := json.Unmarshal([]byte(evt.Data), &req); err != nil {
				// FIXME
			}
			err := s.handleExport(logger, req)
			if err != nil {
				// FIXME: send export response with err
			}
		}
		evt.Commit()
	}
}

// New returns a new server instance
func New(config Config) (*Server, error) {
	server := &Server{config}
	if config.DevMode {
		if err := server.handleExport(config.Logger, agent.ExportRequest{
			JobID:      config.DevExport.JobID(),
			CustomerID: config.DevExport.CustomerID(),
			Integrations: []agent.ExportRequestIntegrations{
				agent.ExportRequestIntegrations{
					Name: config.Integration.Descriptor.RefType,
				},
			},
		}); err != nil {
			return nil, err
		}
		return nil, nil
	}
	go server.run()
	return server, nil
}
