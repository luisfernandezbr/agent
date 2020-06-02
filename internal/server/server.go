package server

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/go-redis/redis"
	"github.com/pinpt/agent.next/internal/export/eventapi"
	pipe "github.com/pinpt/agent.next/internal/pipe/eventapi"
	redisState "github.com/pinpt/agent.next/internal/state/redis"
	"github.com/pinpt/agent.next/sdk"
	"github.com/pinpt/go-common/event"
	"github.com/pinpt/go-common/log"
	pstr "github.com/pinpt/go-common/strings"
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
	Integrations        map[string]*IntegrationContext
	UUID                string
	Channel             string
	APIKey              string
	Secret              string
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

func (s *Server) handleExport(logger log.Logger, evt event.SubscriptionEvent) error {
	var req agent.ExportRequest
	if err := json.Unmarshal([]byte(evt.Data), &req); err != nil {
		return err
	}
	dir := filepath.Join(s.config.Dir, req.JobID)
	os.MkdirAll(dir, 0700)
	var wg sync.WaitGroup
	errors := make(chan error, len(req.Integrations))
	started := time.Now()
	// run our integrations in parallel
	for _, i := range req.Integrations {
		integrationdata := s.config.Integrations[i.Name]
		if integrationdata == nil {
			// FIXME: send an error response
			log.Error(logger, "couldn't find integration named", "name", i.Name)
			continue
		}
		integration := integrationdata.Integration
		descriptor := integrationdata.Descriptor
		// build the sdk config for the integration
		sdkconfig := sdk.Config{}
		for k, v := range i.Authorization.ToMap() {
			sdkconfig[k] = pstr.Value(v)
		}
		state := s.config.State
		if state == nil {
			// if no state provided, we use redis state in this case
			st, err := redisState.New(s.config.RedisClient, req.CustomerID)
			if err != nil {
				return err
			}
			state = st
		}
		wg.Add(1)
		// start the integration in it's own thread
		go func(integration sdk.Integration, descriptor *sdk.Descriptor, state sdk.State) {
			defer wg.Done()
			completion := make(chan eventapi.Completion, 1)
			p := pipe.New(pipe.Config{
				Ctx:        s.config.Ctx,
				Logger:     logger,
				Dir:        dir,
				CustomerID: req.CustomerID,
				UUID:       s.config.UUID,
				JobID:      req.JobID,
				Channel:    s.config.Channel,
				APIKey:     s.config.APIKey,
				Secret:     s.config.Secret,
				RefType:    descriptor.RefType,
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
				Completion: completion,
				Channel:    s.config.Channel,
				APIKey:     s.config.APIKey,
				Secret:     s.config.Secret,
			})
			if err != nil {
				errors <- err
				return
			}
			ts := time.Now()
			if err := integration.Export(c); err != nil {
				errors <- err
				return
			}
			// wait for the integration to complete
			comp := <-completion
			if err := p.Flush(); err != nil {
				log.Error(logger, "error flushing pipe", "err", err)
			}
			if err := state.Flush(); err != nil {
				log.Error(logger, "error flushing state", "err", err)
			}
			log.Debug(logger, "export completed", "integration", descriptor.RefType, "duration", time.Since(ts), "err", comp.Error)
			if comp.Error != nil {
				errors <- comp.Error
				return
			}
		}(integration, descriptor, state)
	}
	log.Debug(logger, "waiting for export to complete")
	wg.Wait()
	var err error
	select {
	case e := <-errors:
		err = e
		break
	default:
	}
	log.Info(logger, "export completed", "duration", time.Since(started), "jobid", req.JobID, "customer_id", req.CustomerID, "err", err)
	return err
}

func (s *Server) run() {
	logger := log.With(s.config.Logger, "pkg", "server")
	log.Debug(logger, "starting subscription channel")
	for evt := range s.config.SubscriptionChannel.Channel() {
		log.Debug(logger, "received event", "evt", evt)
		switch evt.Model {
		case agent.ExportRequestModelName.String():
			err := s.handleExport(logger, evt)
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
	go server.run()
	return server, nil
}
