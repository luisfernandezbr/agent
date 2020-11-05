package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/jhaynie/oauth1"
	eventAPIautoconfig "github.com/pinpt/agent/v4/internal/autoconfig/eventapi"
	eventAPIexport "github.com/pinpt/agent/v4/internal/export/eventapi"
	eventAPImutation "github.com/pinpt/agent/v4/internal/mutation/eventapi"
	pipe "github.com/pinpt/agent/v4/internal/pipe/eventapi"
	redisState "github.com/pinpt/agent/v4/internal/state/redis"
	"github.com/pinpt/agent/v4/internal/util"
	eventAPIvalidate "github.com/pinpt/agent/v4/internal/validate/eventapi"
	eventAPIwebhook "github.com/pinpt/agent/v4/internal/webhook/eventapi"
	"github.com/pinpt/agent/v4/sdk"
	"github.com/pinpt/go-common/v10/api"
	"github.com/pinpt/go-common/v10/datamodel"
	"github.com/pinpt/go-common/v10/datetime"
	"github.com/pinpt/go-common/v10/event"
	"github.com/pinpt/go-common/v10/graphql"
	"github.com/pinpt/go-common/v10/hash"
	"github.com/pinpt/go-common/v10/log"
	"github.com/pinpt/go-common/v10/slack"
	pstrings "github.com/pinpt/go-common/v10/strings"
	"github.com/pinpt/integration-sdk/agent"
	"github.com/pinpt/integration-sdk/sourcecode"
	"github.com/pinpt/integration-sdk/web"
	"github.com/pinpt/integration-sdk/work"
)

// IntegrationContext is the details for each integration
type IntegrationContext struct {
	Integration sdk.Integration
	Descriptor  *sdk.Descriptor
}

// Config is the configuration for the server
type Config struct {
	Ctx          context.Context
	Dir          string // temp dir for files
	Logger       log.Logger
	State        sdk.State     // can be nil
	RedisClient  *redis.Client // can be nil
	Integration  *IntegrationContext
	UUID         string
	Channel      string
	Secret       string
	GroupID      string
	SelfManaged  bool
	APIKey       string
	EnrollmentID string
	SlackToken   string
	SlackChannel string
}

// Server is the event loop server portion of the agent
type Server struct {
	config   Config
	dbchange *Subscriber
	webhook  *Subscriber
	mutation *Subscriber
	validate *Subscriber
	export   *Subscriber
	location string
	ticker   *time.Ticker
	slack    slack.Client
}

var _ io.Closer = (*Server)(nil)

// Close the server
func (s *Server) Close() error {
	if s.dbchange != nil {
		s.dbchange.Close()
		s.dbchange = nil
	}
	if s.export != nil {
		s.export.Close()
		s.export = nil
	}
	if s.validate != nil {
		s.validate.Close()
		s.validate = nil
	}
	if s.webhook != nil {
		s.webhook.Close()
		s.webhook = nil
	}
	if s.mutation != nil {
		s.mutation.Close()
		s.mutation = nil
	}
	return nil
}

func (s *Server) customerSpecificStateKey(customerID string, integrationInstanceID string) string {
	return customerID + ":" + s.config.Integration.Descriptor.RefType + ":" + integrationInstanceID
}

func (s *Server) newState(customerID string, integrationInstanceID string) (sdk.State, error) {
	state := s.config.State
	if state == nil {
		// if no state provided, we use redis state in this case
		st, err := redisState.New(s.config.Ctx, s.config.RedisClient, s.customerSpecificStateKey(customerID, integrationInstanceID))
		if err != nil {
			return nil, err
		}
		state = st
	}
	return state, nil
}

func (s *Server) newPipe(logger sdk.Logger, dir string, customerID string, jobID string, integrationInstanceID string) sdk.Pipe {
	var p sdk.Pipe
	p = pipe.New(pipe.Config{
		Ctx:                   s.config.Ctx,
		Logger:                logger,
		Dir:                   dir,
		CustomerID:            customerID,
		UUID:                  s.config.UUID,
		JobID:                 jobID,
		IntegrationInstanceID: integrationInstanceID,
		Channel:               s.config.Channel,
		APIKey:                s.config.APIKey,
		Secret:                s.config.Secret,
		RefType:               s.config.Integration.Descriptor.RefType,
	})
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

func newConfig(configstr *string) (*sdk.Config, error) {
	sdkconfig := sdk.NewConfig(nil)
	if configstr != nil && *configstr != "" {
		if err := sdkconfig.Parse([]byte(*configstr)); err != nil {
			return nil, err
		}
	}
	return &sdkconfig, nil
}

// fetchConfig will get the config from pinpoint, should only be used for webhooks and mutations
func (s *Server) fetchConfig(client graphql.Client, integrationInstanceID string) (*sdk.Config, error) {
	var first int64 = 1
	integrationRes, err := agent.FindIntegrationInstances(client, &agent.IntegrationInstanceQueryInput{
		First: &first,
		Query: &agent.IntegrationInstanceQuery{
			Filters: []string{
				"_id = ?",
				"active = ?",
			},
			Params: []interface{}{
				integrationInstanceID,
				true,
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("error finding integration instance: %w", err)
	}
	if len(integrationRes.Edges) == 0 {
		return nil, nil
	}
	return newConfig(integrationRes.Edges[0].Node.Config)
}

type cleanupFunc func()

func (s *Server) toInstance(logger sdk.Logger, integration *agent.IntegrationInstance) (*sdk.Instance, cleanupFunc, error) {
	state, err := s.newState(integration.CustomerID, integration.ID)
	if err != nil {
		return nil, nil, err
	}
	dir := s.newTempDir("")
	pipe := s.newPipe(logger, dir, integration.CustomerID, "", integration.ID)
	cleanup := func() {
		pipe.Close()
		os.RemoveAll(dir)
	}
	config, err := newConfig(integration.Config)
	if err != nil {
		return nil, nil, err
	}
	instance := sdk.NewInstance(*config, logger, state, pipe, integration.CustomerID, integration.RefType, integration.ID)
	return instance, cleanup, nil
}

func (s *Server) handleAddIntegration(logger sdk.Logger, integration *agent.IntegrationInstance) error {
	log.Info(logger, "running enroll integration", "id", integration.ID)
	instance, cleanup, err := s.toInstance(logger, integration)
	if err != nil {
		return err
	}
	defer cleanup()
	err = s.config.Integration.Integration.Enroll(*instance)
	s.sendSlackMessage(logger, "enroll", instance.CustomerID(), instance.IntegrationInstanceID(), instance.RefType(), err)
	if err != nil {
		return err
	}
	return nil
}

func (s *Server) handleRemoveIntegration(logger sdk.Logger, integration *agent.IntegrationInstance) error {
	instance, cleanup, err := s.toInstance(logger, integration)
	if err != nil {
		return err
	}
	defer cleanup()
	log.Info(logger, "running dismiss integration", "id", integration.ID, "customer_id", integration.CustomerID)
	err = s.config.Integration.Integration.Dismiss(*instance)
	s.sendSlackMessage(logger, "dismiss", instance.CustomerID(), instance.IntegrationInstanceID(), instance.RefType(), err)
	if err != nil {
		return err
	}
	return nil
}

func (s *Server) handleExport(logger log.Logger, client graphql.Client, req agent.Export) error {
	if req.IntegrationInstanceID == nil {
		log.Error(logger, "received an export for an integration instance id that was nil, ignoring", "req", sdk.Stringify(req))
		return nil
	}
	if err := s.handleEnroll(logger, convertExportIntegrationInstance(req.Integration)); err != nil {
		return err
	}
	dir := s.newTempDir(req.JobID)
	defer os.RemoveAll(dir)
	started := time.Now()
	integration := req.Integration
	sdkconfig, err := newConfig(integration.Config)
	if err != nil {
		return err
	}
	stats := sdk.NewStats()
	if sdkconfig == nil {
		log.Info(logger, "received an export for an integration that no longer exists, ignoring", "id", *req.IntegrationInstanceID)
		return nil
	}
	state, err := s.newState(req.CustomerID, integration.ID)
	if err != nil {
		return err
	}
	p := s.newPipe(logger, dir, req.CustomerID, req.JobID, integration.ID)
	defer p.Close()
	e, err := eventAPIexport.New(eventAPIexport.Config{
		Ctx:                   s.config.Ctx,
		Logger:                logger,
		Config:                *sdkconfig,
		State:                 state,
		CustomerID:            req.CustomerID,
		JobID:                 req.JobID,
		IntegrationInstanceID: *req.IntegrationInstanceID,
		RefType:               s.config.Integration.Descriptor.RefType,
		UUID:                  s.config.UUID,
		Pipe:                  p,
		Channel:               s.config.Channel,
		APIKey:                s.config.APIKey,
		Secret:                s.config.Secret,
		Historical:            req.ReprocessHistorical,
		Stats:                 sdk.PrefixStats(stats, "export"),
	})
	if err != nil {
		return err
	}
	log.Info(logger, "running export")

	eerr := s.config.Integration.Integration.Export(e)
	if err := state.Flush(); err != nil {
		log.Error(logger, "error flushing state", "err", err)
	}
	s.sendSlackMessage(logger, "export", req.CustomerID, req.Integration.IntegrationID, req.Integration.RefType, eerr,
		"job_id", req.JobID,
	)
	var errmsg *string
	if eerr != nil {
		errmsg = pstrings.Pointer(eerr.Error())
	}
	completeEvent := &agent.ExportComplete{
		CustomerID:            req.CustomerID,
		JobID:                 req.JobID,
		IntegrationID:         req.Integration.IntegrationID,
		IntegrationInstanceID: *req.IntegrationInstanceID,
		CreatedAt:             datetime.TimeToEpoch(started),
		StartedAt:             datetime.TimeToEpoch(started),
		EndedAt:               datetime.EpochNow(),
		Historical:            req.ReprocessHistorical,
		Success:               eerr == nil,
		Error:                 errmsg,
		Stats:                 pstrings.Pointer(sdk.Stringify(stats)),
		RefType:               req.RefType,
	}
	id := agent.NewExportCompleteID(req.CustomerID, req.JobID, integration.ID)
	vars := completeEvent.ToMap()
	delete(vars, "id")
	delete(vars, "customer_id")
	delete(vars, "hashcode")
	if err := agent.ExecExportCompleteSilentUpdateMutation(client, id, vars, true); err != nil {
		return fmt.Errorf("error creating export complete: %w", err)
	}
	log.Info(logger, "export completed", "duration", time.Since(started), "jobid", req.JobID, "customer_id", req.CustomerID, "err", eerr)
	if eerr != nil {
		return fmt.Errorf("error running integration export: %w", eerr)
	}
	return nil
}

func (s *Server) handleWebhook(logger log.Logger, client graphql.Client, integrationInstanceID, customerID, webhookURL string, refID string, webhook web.Hook) error {
	buf := []byte(webhook.Data)
	jobID := fmt.Sprintf("webhook_%d", datetime.EpochNow())
	dir := s.newTempDir(jobID)
	defer os.RemoveAll(dir)
	started := time.Now()
	sdkconfig, err := s.fetchConfig(client, integrationInstanceID)
	if err != nil {
		return err
	}
	if sdkconfig == nil {
		log.Info(logger, "received a webhook for an integration that no longer exists, ignoring", "id", integrationInstanceID)
		return nil
	}
	state, err := s.newState(customerID, integrationInstanceID)
	if err != nil {
		return err
	}
	p := s.newPipe(logger, dir, customerID, jobID, integrationInstanceID)
	defer p.Close()
	e := eventAPIwebhook.New(eventAPIwebhook.Config{
		Ctx:                   s.config.Ctx,
		Logger:                logger,
		Config:                *sdkconfig,
		State:                 state,
		CustomerID:            customerID,
		RefID:                 refID,
		RefType:               s.config.Integration.Descriptor.RefType,
		IntegrationInstanceID: integrationInstanceID,
		Pipe:                  p,
		Headers:               webhook.Headers,
		Buf:                   buf,
		WebHookURL:            webhookURL,
	})
	log.Info(logger, "running webhook")
	err = s.config.Integration.Integration.WebHook(e)
	s.sendSlackMessage(logger, "webhook", customerID, integrationInstanceID, s.config.Integration.Descriptor.RefType, err)
	if err != nil {
		return fmt.Errorf("error running integration webhook: %w", err)
	}
	log.Debug(logger, "flushing state")
	if err := state.Flush(); err != nil {
		log.Error(logger, "error flushing state", "err", err)
	}
	log.Info(logger, "webhook completed", "duration", time.Since(started), "ref_id", refID, "customer_id", customerID)
	return nil
}

func (s *Server) handleMutation(logger log.Logger, client graphql.Client, integrationInstanceID, customerID string, refType string, mutation agent.Mutation) (*sdk.MutationResponse, error) {
	buf := []byte(mutation.Payload)
	var data sdk.MutationData
	if err := json.Unmarshal(buf, &data); err != nil {
		return nil, fmt.Errorf("error unmarshaling mutation data payload: %w", err)
	}
	payload, err := sdk.CreateMutationPayloadFromData(data.Model, data.Action, data.Payload)
	if err != nil {
		return nil, fmt.Errorf("error creating mutation payload. %w", err)
	}
	jobID := fmt.Sprintf("mutation_%d", datetime.EpochNow())
	dir := s.newTempDir(jobID)
	defer os.RemoveAll(dir)
	started := time.Now()
	sdkconfig, err := s.fetchConfig(client, integrationInstanceID)
	if err != nil {
		return nil, err
	}
	if sdkconfig == nil {
		log.Info(logger, "received a mutation for an integration that no longer exists, ignoring", "id", integrationInstanceID)
		return nil, nil
	}
	state, err := s.newState(customerID, integrationInstanceID)
	if err != nil {
		return nil, err
	}
	p := s.newPipe(logger, dir, customerID, jobID, integrationInstanceID)
	defer p.Close()
	e := eventAPImutation.New(eventAPImutation.Config{
		Ctx:                   s.config.Ctx,
		Logger:                logger,
		Config:                *sdkconfig,
		State:                 state,
		CustomerID:            customerID,
		RefID:                 data.RefID,
		RefType:               refType,
		IntegrationInstanceID: integrationInstanceID,
		Pipe:                  p,
		ID:                    data.RefID,
		Model:                 data.Model,
		Action:                data.Action,
		Payload:               payload,
		User:                  data.User,
	})
	log.Info(logger, "running mutation", "id", mutation.ID, "customer_id", customerID, "ref_id", data.RefID)
	mr, err := s.config.Integration.Integration.Mutation(e)
	s.sendSlackMessage(logger, "mutation", customerID, integrationInstanceID, refType, err)
	if err != nil {
		return nil, fmt.Errorf("error running integration mutation: %w", err)
	}
	if err := state.Flush(); err != nil {
		log.Error(logger, "error flushing state", "err", err)
	}
	log.Info(logger, "mutation completed", "duration", time.Since(started), "id", mutation.ID, "ref_id", data.RefID, "customer_id", customerID)
	return mr, nil
}

func calculateIntegrationHashCode(integration *agent.IntegrationInstance) string {
	// since we get db change events each time we update the agent.integration table, we don't
	// want to thrash the integration with enrolls so we are going to check specific fields for changes
	// if any of these change, we don't need to send a enroll since the config is the same as before
	return hash.Values(
		integration.Config,
		integration.Active,
		integration.Interval,
	)
}

type deletableState interface {
	DeleteAll() error
}

func makeEnrollCachekey(customerID string, integrationInstanceID string) string {
	return "agent:" + customerID + ":" + integrationInstanceID
}

func (s *Server) onDBChange(logger sdk.Logger, evt event.SubscriptionEvent, refType string, location string) error {
	if refType != s.config.Integration.Descriptor.RefType || location != s.location {
		// skip db changes we're not targeting
		log.Debug(logger, "skipping db change since we're not targeted", "need_location", s.location, "need_reftype", s.config.Integration.Descriptor.RefType, "location", location, "ref_type", refType)
		return nil
	}
	ch, err := createDBChangeEvent(evt.Data)
	if err != nil {
		return err
	}
	switch ch.Model {
	case agent.IntegrationInstanceModelName.String():
		// integration has changed so we need to either enroll or dismiss
		if integration, ok := ch.Object.(*agent.IntegrationInstance); ok {
			cachekey := makeEnrollCachekey(integration.CustomerID, integration.ID)
			// check to see if this is a delete OR we've deleted the integration
			if ch.Action == Delete || integration.Deleted {
				// check cache key or you will get into an infinite loop
				var val int64
				if integration.Location == agent.IntegrationInstanceLocationCloud {
					val = s.config.RedisClient.Exists(s.config.Ctx, cachekey).Val()
				} else {
					if exists := s.config.State.Exists(cachekey); exists {
						val = 1
					}
				}
				log.Debug(logger, "recieved db change for deleted integration", "id", integration.ID, "cachekey", cachekey, "val", val, "will_dismiss", val > 0)
				if val > 0 {
					if integration.Location == agent.IntegrationInstanceLocationCloud {
						// delete the integration cache key and then signal a removal
						defer s.config.RedisClient.Del(s.config.Ctx, cachekey)
					} else {
						defer s.config.State.Delete(cachekey)
					}
					log.Info(logger, "an integration instance has been deleted", "id", integration.ID, "customer_id", integration.CustomerID)
					if err := s.handleRemoveIntegration(logger, integration); err != nil {
						log.Error(logger, "error removing integration", "err", err, "id", integration.ID)
					}
					log.Info(logger, "after removing the integration", "id", integration.ID, "customer_id", integration.CustomerID)
					// check to see if this is redis based state (will be nil) and if so, try and cleanup
					if s.config.State == nil {
						state, err := s.newState(integration.CustomerID, integration.ID)
						if err != nil {
							return err
						}
						// check to see if this deletable state and if so, we delete all the keys
						if ds, ok := state.(deletableState); ok {
							log.Info(logger, "cleaning up state", "id", integration.ID, "customer_id", integration.CustomerID)
							if err := ds.DeleteAll(); err != nil {
								log.Error(logger, "error deleting the integration state", "err", err, "id", integration.ID)
							}
						}
					}
					// go through and cleanup some of our other tables
					gql, err := s.newGraphqlClient(integration.CustomerID)
					if err != nil {
						log.Error(logger, "error creating graphql client", "err", err)
					} else {
						log.Info(logger, "cleaning up integration repo/project errors", "id", integration.ID, "customer_id", integration.CustomerID)
						started := time.Now()
						async := sdk.NewAsync(5)
						var rq sourcecode.RepoErrorQuery
						sourcecode.FindRepoErrorsPaginated(gql, &rq, 100, func(conn *sourcecode.RepoErrorConnection) (bool, error) {
							for _, _edge := range conn.Edges {
								var edge = _edge
								async.Do(func() error {
									if err := sourcecode.ExecRepoErrorDeleteMutation(gql, edge.Node.ID); err != nil {
										log.Error(logger, "error deleting the integration repo error model", "err", err, "integration_instance_id", integration.ID, "id", edge.Node.ID)
									}
									return nil
								})
							}
							return true, nil
						})
						var pq work.ProjectErrorQuery
						work.FindProjectErrorsPaginated(gql, &pq, 100, func(conn *work.ProjectErrorConnection) (bool, error) {
							for _, _edge := range conn.Edges {
								var edge = _edge
								async.Do(func() error {
									if err := work.ExecProjectErrorDeleteMutation(gql, edge.Node.ID); err != nil {
										log.Error(logger, "error deleting the integration project error model", "err", err, "integration_instance_id", integration.ID, "id", edge.Node.ID)
									}
									return nil
								})
							}
							return true, nil
						})
						async.Wait()
						log.Info(logger, "completed clean up of integration repo/project errors", "duration", time.Since(started), "id", integration.ID, "customer_id", integration.CustomerID)
					}
				}
			} else if (ch.Action == Create || ch.Action == Update) &&
				integration.AutoConfigure && !integration.Deleted && !integration.Active && integration.Setup == agent.IntegrationInstanceSetupConfig {
				// this is an auto config for a cloud integration
				var config sdk.Config
				if integration.Config != nil {
					config.Parse([]byte(*integration.Config))
				}
				jobID := fmt.Sprintf("autoconfig_%d", datetime.EpochNow())
				dir := s.newTempDir(jobID)
				defer os.RemoveAll(dir)
				started := time.Now()
				customerID := integration.CustomerID
				integrationInstanceID := integration.ID
				state, err := s.newState(customerID, integrationInstanceID)
				if err != nil {
					return err
				}
				logger := detailLogger(logger, customerID, &integrationInstanceID)
				p := s.newPipe(logger, dir, customerID, jobID, integrationInstanceID)
				defer p.Close()
				e, err := eventAPIautoconfig.New(eventAPIautoconfig.Config{
					Ctx:                   s.config.Ctx,
					Logger:                logger,
					Config:                config,
					State:                 state,
					CustomerID:            customerID,
					RefType:               s.config.Integration.Descriptor.RefType,
					IntegrationInstanceID: integrationInstanceID,
					Pipe:                  p,
					Channel:               s.config.Channel,
					Secret:                s.config.Secret,
					APIKey:                s.config.APIKey,
				})
				log.Info(logger, "running auto configure")
				newconfig, err := s.config.Integration.Integration.AutoConfigure(e)
				if err != nil {
					return fmt.Errorf("error running integration auto configure: %w", err)
				}
				log.Debug(logger, "flushing state")
				if err := state.Flush(); err != nil {
					log.Error(logger, "error flushing state", "err", err)
				}
				gql, err := s.newGraphqlClient(integration.CustomerID)
				if err != nil {
					return fmt.Errorf("error creating graphql client: %w", err)
				}
				input := make(graphql.Variables)
				input[agent.IntegrationInstanceModelActiveColumn] = true
				input[agent.IntegrationInstanceModelSetupColumn] = agent.IntegrationInstanceSetupReady
				input[agent.IntegrationInstanceModelUpdatedAtColumn] = sdk.EpochNow()
				if newconfig != nil {
					input[agent.IntegrationInstanceModelConfigColumn] = sdk.Stringify(newconfig)
				}
				if err := agent.ExecIntegrationInstanceSilentUpdateMutation(gql, integration.ID, input, false); err != nil {
					return fmt.Errorf("error updating agent instance: %w", err)
				}
				log.Info(logger, "auto configure completed", "duration", time.Since(started), "customer_id", customerID)
			}
		}
	}
	return nil
}

func convertExportIntegrationInstance(submodel agent.ExportIntegration) *agent.IntegrationInstance {
	var integration agent.IntegrationInstance
	integration.FromMap(submodel.ToMap())
	return &integration
}

// handleEnroll will call enroll if the integration is new or has been reconfigured
func (s *Server) handleEnroll(logger log.Logger, integrationInstance *agent.IntegrationInstance) error {
	cachekey := makeEnrollCachekey(integrationInstance.CustomerID, integrationInstance.ID)
	hashcode := calculateIntegrationHashCode(integrationInstance)
	iscloud := integrationInstance.Location == agent.IntegrationInstanceLocationCloud
	var install bool
	var res string
	var err error
	// if it's active then check to see if we're updated
	if iscloud {
		res, err = s.config.RedisClient.Get(s.config.Ctx, cachekey).Result()
	} else {
		_, err = s.config.State.Get(cachekey, &res)
	}
	if err != nil {
		if err != redis.Nil {
			return fmt.Errorf("error getting cachekey for state: %w", err)
		}
		// not in cache so it must be new
		install = true
	}
	if !install {
		// check if its changed configuration
		install = res != hashcode
		log.Debug(logger, "comparing integration hashcode on integration change", "hashcode", hashcode, "res", res, "install", install, "id", integrationInstance.ID)
	}
	if install {
		// update our hash key and then signal an addition
		if iscloud {
			if err := s.config.RedisClient.Set(s.config.Ctx, cachekey, hashcode, 0).Err(); err != nil {
				return fmt.Errorf("error setting the cache key (%s) on the install: %w", cachekey, err)
			}
		} else {
			if err := s.config.State.Set(cachekey, hashcode); err != nil {
				return fmt.Errorf("error setting the cache key (%s) on the install: %w", cachekey, err)
			}
		}
		log.Info(logger, "an integration instance has been added", "id", integrationInstance.ID, "customer_id", integrationInstance.CustomerID, "cachekey", cachekey, "hashcode", hashcode)
		if err := s.handleAddIntegration(logger, integrationInstance); err != nil {
			return fmt.Errorf("error adding integration instance (%s): %w", integrationInstance.ID, err)
		}
	}
	return nil
}

func (s *Server) newGraphqlClient(customerID string) (graphql.Client, error) {
	cl, err := graphql.NewClient(
		customerID,
		"",
		s.config.Secret,
		api.BackendURL(api.GraphService, s.config.Channel),
	)
	if err != nil {
		return nil, err
	}
	if s.config.APIKey != "" {
		cl.SetHeader("Authorization", s.config.APIKey)
	}
	return cl, nil
}

// toResponseErr will return a pointer to the error's string value, or nil if the error is nil
func toResponseErr(err error) *string {
	if err == nil {
		return nil
	}
	var str string
	str = err.Error()
	return &str
}

func detailLogger(logger log.Logger, customerID string, integrationInstanceID *string) log.Logger {
	if integrationInstanceID != nil {
		return log.With(logger, "customer_id", customerID, "integration_instance_id", *integrationInstanceID)
	}
	return log.With(logger, "customer_id", customerID)
}

// eventPublish will publish a model on the subscription channel
func (s *Server) eventPublish(logger sdk.Logger, ch *event.SubscriptionChannel, model datamodel.Model, headers map[string]string) {
	// publish on another thread because we're inside s.event's cosumer loop
	go func() {
		log.Debug(logger, "publishing an event", "model", model, "headers", headers)
		ts := time.Now()
		if err := ch.Publish(event.PublishEvent{
			Object:  model,
			Headers: headers,
			Logger:  logger,
		}); err != nil {
			log.Error(logger, "error publishing %s: %w", model.GetModelName(), err)
		}
		log.Debug(logger, "publishing an event (sent)", "model", model, "headers", headers, "duration", time.Since(ts))
	}()
}

func (s *Server) onValidate(logger sdk.Logger, req agent.ValidateRequest) (*string, error) {
	cfg, err := newConfig(&req.Config)
	if err != nil {
		return nil, fmt.Errorf("error parsing config: %w", err)
	}
	if cfg == nil {
		return nil, fmt.Errorf("parse config was nil")
	}
	if req.IntegrationInstanceID == nil {
		return nil, fmt.Errorf("missing required integration_instance_id")
	}
	client, err := s.newGraphqlClient(req.CustomerID)
	if err != nil {
		return nil, err
	}
	state, err := s.newState(req.CustomerID, *req.IntegrationInstanceID)
	if err != nil {
		return nil, err
	}
	resp, err := s.config.Integration.Integration.Validate(eventAPIvalidate.NewValidate(
		*cfg,
		logger,
		req.RefType,
		req.CustomerID,
		*req.IntegrationInstanceID,
		client,
		state,
	))
	var result *string
	if err != nil {
		return nil, err
	}
	if resp != nil {
		buf, err := json.Marshal(resp)
		if err != nil {
			return nil, fmt.Errorf("error encoding validation result to json: %w", err)
		}
		result = pstrings.Pointer(string(buf))
	}
	return result, nil
}

// onOauth1 fetches the token and returns a requestToken and requestSecret
func (s *Server) onOauth1(logger sdk.Logger, req agent.Oauth1Request) (string, string, error) {
	log.Debug(logger, "on OAuth1 request", "req", sdk.Stringify(req))
	key, err := util.ParsePrivateKey(req.PrivateKey)
	if err != nil {
		return "", "", err
	}
	var endpoint oauth1.Endpoint
	switch req.Stage {
	case agent.Oauth1RequestStageRequestToken:
		endpoint.RequestTokenURL = req.URL
	case agent.Oauth1RequestStageAccessToken:
		endpoint.AccessTokenURL = req.URL
	}
	config := oauth1.Config{
		ConsumerKey:    req.ConsumerKey,
		ConsumerSecret: req.ConsumerSecret,
		Endpoint:       endpoint,
		Signer:         &oauth1.RSASigner{PrivateKey: key},
		CallbackURL:    req.CallbackURL,
	}
	switch req.Stage {
	case agent.Oauth1RequestStageRequestToken:
		return config.RequestToken()
	case agent.Oauth1RequestStageAccessToken:
		return config.AccessToken(*req.Token, *req.TokenSecret, *req.Code)
	}
	return "", "", fmt.Errorf("invalid stage requested")
}

// onOauth1Identity fetches the oauth1 identity of a user
func (s *Server) onOauth1Identity(logger sdk.Logger, req agent.Oauth1UserIdentityRequest) (*sdk.OAuth1Identity, error) {
	log.Debug(logger, "on OAuth1 identity request", "req", sdk.Stringify(req))
	key, err := util.ParsePrivateKey(req.PrivateKey)
	if err != nil {
		return nil, err
	}
	if fn, ok := s.config.Integration.Integration.(sdk.OAuth1Integration); ok {
		identifier := sdk.NewSimpleIdentifier(req.CustomerID, *req.IntegrationInstanceID, req.RefType)
		return fn.IdentifyOAuth1User(identifier, req.URL, key, req.ConsumerKey, req.ConsumerSecret, *req.Token, *req.TokenSecret)
	}
	return nil, fmt.Errorf("error fulfilling the oauth1 identity request because the integration doesn't support it")
}

func (s *Server) makeExportStat(logger sdk.Logger, integrationInstanceID, customerID, jobID string) {
	if client, err := s.newGraphqlClient(customerID); err == nil {
		id := agent.NewExportStatID(integrationInstanceID)
		err := agent.ExecExportStatSilentUpdateMutation(client, id, graphql.Variables{
			agent.ExportStatModelIntegrationInstanceIDColumn: integrationInstanceID,
			agent.ExportStatModelJobIDColumn:                 jobID,
			agent.ExportStatModelCreatedDateColumn:           agent.ExportStatCreatedDate(datetime.NewDateNow()),
		}, true)
		if err != nil {
			log.Error(logger, "error creating liveness record", "err", err)
		} else {
			log.Debug(logger, "created export liveness record")
		}
	} else {
		log.Error(logger, "error creating client for liveness", "err", err)
	}
}

func (s *Server) startExportLiveness(logger sdk.Logger, export agent.Export) {
	s.ticker = time.NewTicker(time.Minute * 1)
	s.makeExportStat(logger, *export.IntegrationInstanceID, export.CustomerID, export.JobID)
	go func(integrationInstanceID, customerID, jobID string) {
		for range s.ticker.C {
			s.makeExportStat(logger, integrationInstanceID, customerID, jobID)
		}
	}(*export.IntegrationInstanceID, export.CustomerID, export.JobID)
}

func (s *Server) stopExportLiveness() {
	s.ticker.Stop()
}

func (s *Server) onEvent(logger sdk.Logger, evt event.SubscriptionEvent, refType string, location string) error {
	logger = log.With(logger, "ref_type", refType)
	log.Debug(logger, "received event", "evt", evt, "refType", refType, "location", location)
	switch evt.Model {
	case agent.ValidateRequestModelName.String():
		var req agent.ValidateRequest
		if err := json.Unmarshal([]byte(evt.Data), &req); err != nil {
			// NOTE: This is a serious error because if we can't decode the body then
			// we can't set the sessionID on the resp so the ui won't recieve the message,
			// so it will hang forever.
			return fmt.Errorf("critical error parsing validation request: %w", err)
		}
		logger = detailLogger(logger, req.CustomerID, req.IntegrationInstanceID)
		result, err := s.onValidate(logger, req)
		res := &agent.ValidateResponse{
			CustomerID: req.CustomerID,
			Error:      toResponseErr(err),
			RefType:    req.RefType,
			SessionID:  req.SessionID,
			Result:     result,
			Success:    err == nil,
		}
		headers := map[string]string{
			"ref_type":    req.RefType,
			"session_id":  req.SessionID,
			"customer_id": req.CustomerID,
		}
		if req.EnrollmentID != nil {
			headers["enrollment_id"] = *req.EnrollmentID
		}
		s.eventPublish(logger, s.validate.ch, res, headers)
		if err != nil {
			log.Error(logger, "sent validation response with error", "result", result, "err", err.Error(), "headers", headers)
		} else {
			log.Info(logger, "sent validation response", "result", result, "headers", headers)
		}
	case agent.Oauth1RequestModelName.String():
		var req agent.Oauth1Request
		if err := json.Unmarshal([]byte(evt.Data), &req); err != nil {
			// NOTE: This is a serious error because if we can't decode the body then
			// we can't set the sessionID on the resp so the ui won't recieve the message,
			// so it will hang forever.
			return fmt.Errorf("critical error parsing oauth1 request: %w", err)
		}
		logger = detailLogger(logger, req.CustomerID, req.IntegrationInstanceID)
		token, secret, err := s.onOauth1(logger, req)
		res := &agent.Oauth1Response{
			CustomerID: req.CustomerID,
			Error:      toResponseErr(err),
			RefType:    req.RefType,
			SessionID:  req.SessionID,
			Success:    err == nil,
			Token:      sdk.StringPointer(token),
			Secret:     sdk.StringPointer(secret),
		}
		headers := map[string]string{
			"ref_type":    req.RefType,
			"session_id":  req.SessionID,
			"customer_id": req.CustomerID,
		}
		if req.EnrollmentID != nil {
			headers["enrollment_id"] = *req.EnrollmentID
		}
		s.eventPublish(logger, s.validate.ch, res, headers)
		if err != nil {
			log.Error(logger, "sent oauth1 response with error", "err", err.Error(), "headers", headers)
		} else {
			log.Info(logger, "sent oauth1 response", "headers", headers)
		}
	case agent.Oauth1UserIdentityRequestModelName.String():
		var req agent.Oauth1UserIdentityRequest
		if err := json.Unmarshal([]byte(evt.Data), &req); err != nil {
			// NOTE: This is a serious error because if we can't decode the body then
			// we can't set the sessionID on the resp so the ui won't recieve the message,
			// so it will hang forever.
			return fmt.Errorf("critical error parsing oauth1 identity request: %w", err)
		}
		logger = detailLogger(logger, req.CustomerID, req.IntegrationInstanceID)
		identity, err := s.onOauth1Identity(logger, req)
		var errmsg *string
		if err != nil {
			errmsg = sdk.StringPointer(err.Error())
		}
		var identityField *agent.Oauth1UserIdentityResponseIdentity
		if err == nil {
			identityField = &agent.Oauth1UserIdentityResponseIdentity{
				Name:      identity.Name,
				Email:     identity.Email,
				AvatarURL: identity.AvatarURL,
			}
		}
		res := &agent.Oauth1UserIdentityResponse{
			CustomerID:            req.CustomerID,
			Error:                 errmsg,
			Success:               err == nil,
			RefType:               req.RefType,
			Identity:              identityField,
			IntegrationInstanceID: req.IntegrationInstanceID,
			SessionID:             req.SessionID,
		}
		if identity != nil {
			res.RefID = identity.RefID
		}
		headers := map[string]string{
			"ref_type":    req.RefType,
			"customer_id": req.CustomerID,
			"session_id":  req.SessionID,
		}
		if req.EnrollmentID != nil {
			headers["enrollment_id"] = *req.EnrollmentID
		}
		s.eventPublish(logger, s.validate.ch, res, headers)
		if err != nil {
			log.Error(logger, "sent oauth1 identity response with error", "err", err.Error(), "headers", headers)
		} else {
			log.Info(logger, "sent oauth1 identity response", "headers", headers, "identity", identity)
		}
	case agent.ExportModelName.String():
		var req agent.Export
		if err := json.Unmarshal([]byte(evt.Data), &req); err != nil {
			log.Fatal(logger, "error parsing export event", "err", err)
		}
		logger = detailLogger(logger, req.CustomerID, req.IntegrationInstanceID)
		if req.Integration.Location.String() != location {
			log.Info(logger, "skipping export request, location of integration does not match agent location", "integration", req.Integration.Location.String(), "agent", location)
			break
		}
		if time.Since(evt.Timestamp) > time.Minute*5 {
			log.Info(logger, "skipping export request, too old", "age", time.Since(evt.Timestamp), "id", evt.ID)
			break
		}
		cl, err := s.newGraphqlClient(req.CustomerID)
		if err != nil {
			return fmt.Errorf("error creating graphql client: %w", err)
		}
		instanceID := *req.IntegrationInstanceID
		instance, err := agent.FindIntegrationInstance(cl, instanceID)
		if err != nil {
			return fmt.Errorf("error finding integration instance (%v): %w", instanceID, err)
		}
		if instance == nil {
			log.Info(logger, "skipping export request because the integration instance no longer exists in the db", "id", instanceID)
			return nil
		}
		if !instance.Active {
			log.Info(logger, "skipping export request because the integration instance is no longer active", "id", instanceID)
			return nil
		}
		// update the integration state to acknowledge that we are exporting
		vars := make(graphql.Variables)
		vars[agent.IntegrationInstanceModelStateColumn] = agent.IntegrationInstanceStateExporting
		if err := agent.ExecIntegrationInstanceSilentUpdateMutation(cl, instanceID, vars, false); err != nil {
			log.Error(logger, "error updating agent integration", "err", err, "id", instanceID)
		}
		s.startExportLiveness(logger, req)
		defer s.stopExportLiveness()
		var errmessage *string
		if err := s.handleExport(logger, cl, req); err != nil {
			log.Error(logger, "error running export request", "err", err)
			errmessage = sdk.StringPointer(err.Error())
		}
		// update the db with our new integration state
		vars = make(graphql.Variables)
		vars[agent.IntegrationInstanceModelStateColumn] = agent.IntegrationInstanceStateIdle
		if errmessage != nil {
			vars[agent.IntegrationInstanceModelErroredColumn] = true
			vars[agent.IntegrationInstanceModelErrorMessageColumn] = *errmessage
			vars[agent.IntegrationInstanceModelErrorDateColumn] = datetime.NewDateNow()
		} else {
			vars[agent.IntegrationInstanceModelErroredColumn] = false
			vars[agent.IntegrationInstanceModelErrorMessageColumn] = nil
			vars[agent.IntegrationInstanceModelErrorDateColumn] = datetime.NewDateFromEpoch(0)
		}
		if err := agent.ExecIntegrationInstanceSilentUpdateMutation(cl, instanceID, vars, false); err != nil {
			log.Error(logger, "error updating agent integration", "err", err, "id", instanceID)
		}
		vars = make(graphql.Variables)
		ts := time.Now()
		var dt agent.IntegrationInstanceStatLastExportCompletedDate
		sdk.ConvertTimeToDateModel(ts, &dt)
		vars[agent.IntegrationInstanceStatModelLastExportCompletedDateColumn] = dt
		if req.ReprocessHistorical {
			var dt agent.IntegrationInstanceStatLastHistoricalCompletedDate
			sdk.ConvertTimeToDateModel(ts, &dt)
			vars[agent.IntegrationInstanceStatModelLastHistoricalCompletedDateColumn] = dt
		}
		if err := agent.ExecIntegrationInstanceStatSilentUpdateMutation(cl, agent.NewIntegrationInstanceStatID(instanceID), vars, false); err != nil {
			log.Error(logger, "error updating agent integration stat", "err", err, "id", instanceID)
		}
	}
	return nil
}

func (s *Server) onWebhook(logger sdk.Logger, evt event.SubscriptionEvent, refType string, location string) error {
	log.Debug(logger, "received webhook event", "evt", evt)
	customerID := evt.Headers["customer_id"]
	if customerID == "" {
		return errors.New("webhook missing customer id")
	}
	integrationInstanceID := evt.Headers["integration_instance_id"]
	if integrationInstanceID == "" {
		return errors.New("webhook missing integration id")
	}
	logger = detailLogger(logger, customerID, &integrationInstanceID)
	switch evt.Model {
	case web.HookModelName.String():
		var wh web.Hook
		if err := json.Unmarshal([]byte(evt.Data), &wh); err != nil {
			log.Fatal(logger, "error parsing webhook", "err", err)
		}
		wehbookURL := evt.Headers["webhook_url"]
		if wehbookURL == "" {
			return errors.New("webhook missing webhook_url")
		}
		cl, err := graphql.NewClient(
			customerID,
			"",
			s.config.Secret,
			api.BackendURL(api.GraphService, s.config.Channel),
		)
		if err != nil {
			return fmt.Errorf("error creating graphql client: %w", err)
		}
		var errmessage *string
		// TODO(robin): maybe scrub some event-api related fields out of the headers
		if err := s.handleWebhook(logger, cl, integrationInstanceID, customerID, wehbookURL, evt.Headers["ref_id"], wh); err != nil {
			log.Error(logger, "error running webhook", "err", err)
			errmessage = sdk.StringPointer(err.Error())
		}
		// update the db with our new integration state
		if errmessage != nil {
			vars := make(graphql.Variables)
			vars[agent.IntegrationInstanceModelErroredColumn] = true
			vars[agent.IntegrationInstanceModelErrorMessageColumn] = *errmessage
			if err := agent.ExecIntegrationInstanceSilentUpdateMutation(cl, integrationInstanceID, vars, false); err != nil {
				log.Error(logger, "error updating agent integration", "err", err, "id", integrationInstanceID)
			}
		}
	}
	return nil
}

func (s *Server) onMutation(logger sdk.Logger, evt event.SubscriptionEvent, refType string, location string) error {
	logger = detailLogger(logger, evt.Headers["customer_id"], pstrings.Pointer(evt.Headers["integration_instance_id"]))
	log.Debug(logger, "received mutation event", "evt", evt)
	switch evt.Model {
	case agent.MutationModelName.String():
		var m agent.Mutation
		if err := json.Unmarshal([]byte(evt.Data), &m); err != nil {
			log.Fatal(logger, "error parsing muation", "err", err)
		}
		logger = detailLogger(logger, m.CustomerID, m.IntegrationInstanceID)
		if m.IntegrationInstanceID == nil {
			log.Error(logger, "mutation event is missing integration instance id, skipping")
			break
		}
		cl, err := s.newGraphqlClient(m.CustomerID)
		if err != nil {
			log.Error(logger, "error creating graphql client", "err',err")
		}
		var errmessage *string
		// TODO(robin): maybe scrub some event-api related fields out of the headers
		mr, err := s.handleMutation(logger, cl, *m.IntegrationInstanceID, m.CustomerID, refType, m)
		if err != nil {
			log.Error(logger, "error running mutation", "err", err)
			errmessage = sdk.StringPointer(err.Error())
		}
		// send the response to the mutation, but we need to do this on a separate go routine
		var resp agent.MutationResponse
		resp.ID = agent.NewMutationResponseID(m.CustomerID)
		resp.CustomerID = m.CustomerID
		resp.IntegrationInstanceID = m.IntegrationInstanceID
		resp.Error = errmessage
		resp.SessionID = m.SessionID
		resp.Success = errmessage == nil
		if mr == nil {
			resp.RefID = m.RefID
		} else {
			if mr.RefID != nil {
				resp.RefID = *mr.RefID
			} else {
				resp.RefID = m.RefID
			}
			if mr.Properties != nil {
				resp.Properties = sdk.StringPointer(sdk.Stringify(mr.Properties))
			}
			resp.URL = mr.URL
			resp.EntityID = mr.EntityID
			resp.RefType = m.RefType
		}
		s.eventPublish(logger, s.mutation.ch, &resp, map[string]string{
			"ref_type":                m.RefType,
			"ref_id":                  m.RefID,
			"integration_instance_id": *m.IntegrationInstanceID,
			"session_id":              m.SessionID,
		})
	}
	return nil
}

type noOpSlackClient struct{}

func (n noOpSlackClient) SendMessage(msg string) error {
	return nil
}

// New returns a new server instance
func New(config Config) (*Server, error) {
	location := agent.ExportIntegrationLocationPrivate
	if !config.SelfManaged {
		location = agent.ExportIntegrationLocationCloud
	}
	slackClient, err := slack.New(config.SlackToken, config.SlackChannel)
	if err != nil {
		log.Warn(config.Logger, "error creating slack client. Will continue without it", "err", err)
		slackClient = &noOpSlackClient{}
	}
	server := &Server{
		config:   config,
		location: location.String(),
		slack:    slackClient,
	}
	server.dbchange, err = NewDBChangeSubscriber(config, location, config.Integration.Descriptor.RefType, server.onDBChange, config.Integration.Descriptor.RefType, "integration")
	if err != nil {
		return nil, err
	}
	exportObjectExpr := fmt.Sprintf(`ref_type:"%s" AND integration.location:"%s"`, config.Integration.Descriptor.RefType, location.String())
	// validateObjectExpr also works for oauth1 request
	var validateObjectExpr string
	if !config.SelfManaged {
		validateObjectExpr = fmt.Sprintf(`ref_type:"%s" AND enrollment_id:null`, config.Integration.Descriptor.RefType)
	} else {
		validateObjectExpr = fmt.Sprintf(`ref_type:"%s" AND enrollment_id:"%s"`, config.Integration.Descriptor.RefType, config.EnrollmentID)
	}
	server.export, err = NewEventSubscriber(
		config,
		[]string{
			agent.ExportModelName.String(),
		},
		&event.SubscriptionFilter{
			ObjectExpr: exportObjectExpr,
		},
		location,
		server.onEvent,
		config.Integration.Descriptor.RefType,
		"export",
	)
	if err != nil {
		return nil, err
	}
	server.validate, err = NewEventSubscriber(
		config,
		[]string{
			agent.Oauth1RequestModelName.String(),
			agent.Oauth1UserIdentityRequestModelName.String(),
			agent.ValidateRequestModelName.String(),
		},
		&event.SubscriptionFilter{
			ObjectExpr: validateObjectExpr,
		},
		location,
		server.onEvent,
		config.Integration.Descriptor.RefType,
		"validate",
	)
	if err != nil {
		return nil, err
	}
	server.webhook, err = NewEventSubscriber(
		config,
		[]string{web.HookModelName.String()},
		&event.SubscriptionFilter{
			HeaderExpr: fmt.Sprintf(`self_managed:"%v"`, location == agent.ExportIntegrationLocationPrivate),
			ObjectExpr: fmt.Sprintf(`system:"%s"`, config.Integration.Descriptor.RefType),
		},
		location,
		server.onWebhook,
		config.Integration.Descriptor.RefType,
		"webhook",
	)
	if err != nil {
		return nil, fmt.Errorf("error starting webhook subscriber: %w", err)
	}
	server.mutation, err = NewEventSubscriber(
		config,
		[]string{agent.MutationModelName.String()},
		&event.SubscriptionFilter{
			HeaderExpr: fmt.Sprintf(`ref_type:"%s" AND location:"%s"`, config.Integration.Descriptor.RefType, location.String()),
		},
		location,
		server.onMutation,
		config.Integration.Descriptor.RefType,
		"mutation",
	)
	if err != nil {
		return nil, fmt.Errorf("error starting mutation subscriber: %w", err)
	}
	// commenting for now, this will need to be in registry instead of here
	// err = slackClient.SendMessage("🎉 *"+config.Integration.Descriptor.RefType+"* integration published", "sha", config.Integration.Descriptor.BuildCommitSHA)
	return server, nil
}
