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

func (s *Server) newPipe(logger sdk.Logger, dir string, customerID string, jobID string) sdk.Pipe {
	var p sdk.Pipe
	if s.config.DevMode {
		p = s.config.DevPipe
	} else {
		p = pipe.New(pipe.Config{
			Ctx:        s.config.Ctx,
			Logger:     logger,
			Dir:        dir,
			CustomerID: customerID,
			UUID:       s.config.UUID,
			JobID:      jobID,
			Channel:    s.config.Channel,
			APIKey:     s.config.APIKey,
			Secret:     s.config.Secret,
			RefType:    s.config.Integration.Descriptor.RefType,
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

func (s *Server) handleAddIntegration(logger log.Logger, req agent.IntegrationRequest) error {
	if s.config.Integration.Descriptor.RefType == req.Integration.RefType {
		state, err := s.newState(req.CustomerID, req.Integration.ID)
		if err != nil {
			return err
		}
		dir := s.newTempDir("")
		pipe := s.newPipe(logger, dir, req.CustomerID, "")
		defer func() {
			pipe.Close()
			os.RemoveAll(dir)
		}()
		instance := sdk.NewInstance(state, pipe, req.CustomerID, req.Integration.ID)
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
	// build the sdk config for the integration
	var sdkconfig sdk.Config
	if s.config.DevMode {
		sdkconfig = s.config.DevExport.Config()
	} else {
		sdkconfig = sdk.NewConfig(integrationAuth.ToMap())
	}
	state, err := s.newState(req.CustomerID, integration.ID)
	if err != nil {
		return err
	}
	p := s.newPipe(logger, dir, req.CustomerID, req.JobID)
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

func toResponseIntegrations(integrations []agent.ExportRequestIntegrations, exportType agent.ExportResponseIntegrationsExportType) []agent.ExportResponseIntegrations {
	// FIXME: need to deal with entity errors somehow
	resp := make([]agent.ExportResponseIntegrations, 0)
	for _, i := range integrations {
		resp = append(resp, agent.ExportResponseIntegrations{
			IntegrationID: i.ID,
			Name:          i.Name,
			ExportType:    exportType,
			SystemType:    agent.ExportResponseIntegrationsSystemType(i.SystemType),
		})
	}
	return resp
}

func (s *Server) toHeaders(customerID string, refType string) map[string]string {
	headers := map[string]string{
		"customer_id": customerID,
		"ref_type":    refType,
	}
	if s.config.UUID != "" {
		headers["uuid"] = s.config.UUID
	}
	return headers
}

func (s *Server) run() {
	logger := log.With(s.config.Logger, "pkg", "server")
	log.Debug(logger, "starting subscription channel")
	opts := []event.Option{
		event.WithLogger(logger),
	}
	if s.config.Secret != "" {
		opts = append(opts, event.WithHeaders(map[string]string{"x-api-key": s.config.Secret}))
	}
	// TODO: need an event to remove an integration

	for evt := range s.config.SubscriptionChannel.Channel() {
		log.Debug(logger, "received event", "evt", evt)
		switch evt.Model {
		case agent.IntegrationRequestModelName.String():
			var req agent.IntegrationRequest
			if err := json.Unmarshal([]byte(evt.Data), &req); err != nil {
				log.Fatal(logger, "error parsing integration request event", "err", err)
			}
			var errmessage *string
			err := s.handleAddIntegration(logger, req)
			if err != nil {
				log.Error(logger, "error running add integration request", "err", err)
				errmessage = sdk.StringPointer(err.Error())
			}
			dt, err := sdk.NewDateWithTime(time.Now())
			if err != nil {
				log.Fatal(logger, "error parsing date time", "err", err)
			}
			resp := &agent.IntegrationResponse{
				CustomerID: req.CustomerID,
				Success:    err == nil,
				Error:      errmessage,
				UUID:       req.UUID,
				EventDate: agent.IntegrationResponseEventDate{
					Epoch:   dt.Epoch,
					Offset:  dt.Offset,
					Rfc3339: dt.Rfc3339,
				},
				Type:    agent.IntegrationResponseTypeIntegration,
				RefType: req.RefType,
			}
			if s.config.SubscriptionChannel.Publish(event.PublishEvent{
				Object:  resp,
				Headers: s.toHeaders(req.CustomerID, req.RefType),
				Logger:  logger,
			}, opts...); err != nil {
				log.Fatal(logger, "error sending add integration response", "err", err)
			}
		case agent.ExportRequestModelName.String():
			var req agent.ExportRequest
			if err := json.Unmarshal([]byte(evt.Data), &req); err != nil {
				log.Fatal(logger, "error parsing export request event", "err", err)
			}
			var errmessage *string
			err := s.handleExport(logger, req)
			if err != nil {
				log.Error(logger, "error running export request", "err", err)
				errmessage = sdk.StringPointer(err.Error())
			}
			exportType := agent.ExportResponseIntegrationsExportTypeIncremental
			if req.ReprocessHistorical {
				exportType = agent.ExportResponseIntegrationsExportTypeHistorical
			}
			dt, err := sdk.NewDateWithTime(time.Now())
			if err != nil {
				log.Fatal(logger, "error parsing date time", "err", err)
			}
			resp := &agent.ExportResponse{
				CustomerID:   req.CustomerID,
				Success:      err == nil,
				Error:        errmessage,
				JobID:        req.JobID,
				Integrations: toResponseIntegrations(req.Integrations, exportType),
				UUID:         req.UUID,
				EventDate: agent.ExportResponseEventDate{
					Epoch:   dt.Epoch,
					Offset:  dt.Offset,
					Rfc3339: dt.Rfc3339,
				},
				State: agent.ExportResponseStateCompleted,
				Type:  agent.ExportResponseTypeExport,
			}
			if s.config.SubscriptionChannel.Publish(event.PublishEvent{
				Object:  resp,
				Headers: s.toHeaders(req.CustomerID, req.RefType),
				Logger:  logger,
			}, opts...); err != nil {
				log.Fatal(logger, "error sending export response", "err", err)
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
					ID:   "1",
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
