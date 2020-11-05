package server

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/pinpt/agent/v4/sdk"
	"github.com/pinpt/go-common/v10/datamodel"
	"github.com/pinpt/go-common/v10/event"
	"github.com/pinpt/go-common/v10/log"
	"github.com/pinpt/go-common/v10/metrics"
	isdk "github.com/pinpt/integration-sdk"
	"github.com/pinpt/integration-sdk/agent"
)

// Action is the db change action type for the change set
type Action string

const (
	// Create is the action when a new record is created
	Create Action = "create"
	// Update is the action when an existing record is updated
	Update Action = "update"
	// Delete is the action when an existing record is deleted
	Delete Action = "delete"
)

// DbChangeEvent is an event which encapsulates a db change
type DbChangeEvent struct {
	Action Action          `json:"action"`
	Data   string          `json:"data"`
	Model  string          `json:"model"`
	Object datamodel.Model `json:"-"`
}

// these are the models we're going to listen for db change events on
var models = []string{
	agent.IntegrationInstanceModelName.String(),
}

// createDBChangeSubscriptionFilter will create the subscription filter for a specific refType
func createDBChangeSubscriptionFilter(refType string, location agent.ExportIntegrationLocation) *event.SubscriptionFilter {
	// TODO(robin): just put this in server.New, its already too specific
	modelexpr := []string{}
	for _, model := range models {
		modelexpr = append(modelexpr, fmt.Sprintf(`model:"%s"`, model))
	}
	return &event.SubscriptionFilter{
		HeaderExpr: "(" + strings.Join(modelexpr, " OR ") + `) AND origin:"graph"`,
		ObjectExpr: fmt.Sprintf("data.ref_type: \"%s\" AND data.location: \"%s\"", refType, location.String()),
	}
}

// createDBChangeEvent returns a db change event object from a db change data payload
func createDBChangeEvent(data string) (*DbChangeEvent, error) {
	var event DbChangeEvent
	if err := json.Unmarshal([]byte(data), &event); err != nil {
		return nil, fmt.Errorf("error unmarshaling db change event: %w", err)
	}
	instance := isdk.New(datamodel.ModelNameType(event.Model))
	if err := json.Unmarshal([]byte(event.Data), instance); err != nil {
		return nil, fmt.Errorf("error unmarshaling db change data payload: %w", err)
	}
	event.Object = instance
	return &event, nil
}

// Subscriber is a convenience wrapper around a subscription channel
type Subscriber struct {
	ch                  *event.SubscriptionChannel
	cb                  SubscriberCallback
	logger              log.Logger
	location            string
	refType             string
	metricServiceName   string
	metricOperationName string
}

func (s *Subscriber) run() {
	for event := range s.ch.Channel() {
		ts := time.Now()
		if err := s.cb(s.logger, event, s.refType, s.location); err != nil {
			metrics.RequestsTotal.WithLabelValues(s.metricServiceName, s.metricOperationName, "500").Inc()
			log.Error(s.logger, "error from callback", "err", err)
		} else {
			metrics.RequestsTotal.WithLabelValues(s.metricServiceName, s.metricOperationName, "200").Inc()
			metrics.RequestDurationMilliseconds.WithLabelValues(s.metricServiceName, s.metricOperationName).Observe(float64(time.Since(ts).Milliseconds()))
		}
		event.Commit()
	}
}

// Close will close the subscriber
func (s *Subscriber) Close() error {
	return s.ch.Close()
}

// SubscriberCallback is the callback for processing events
type SubscriberCallback func(logger sdk.Logger, event event.SubscriptionEvent, refType string, location string) error

// NewDBChangeSubscriber will return a db change subscriber
func NewDBChangeSubscriber(config Config, location agent.ExportIntegrationLocation, refType string, callback SubscriberCallback, metricServiceName, metricOperationName string) (*Subscriber, error) {
	return NewEventSubscriber(config, []string{"ops.db.Change"}, createDBChangeSubscriptionFilter(refType, location), location, callback, metricServiceName, metricOperationName)
}

// NewEventSubscriber will return an event subscriber
func NewEventSubscriber(config Config, topics []string, filters *event.SubscriptionFilter, location agent.ExportIntegrationLocation, callback SubscriberCallback, metricServiceName, metricOperationName string) (*Subscriber, error) {
	httpheaders := map[string]string{}
	if config.Secret != "" {
		httpheaders["x-api-key"] = config.Secret
	}
	log.Info(config.Logger, "creating NewEventSubscriber", "topics", topics, "headers", filters.HeaderExpr, "object", filters.ObjectExpr)
	ch, err := event.NewSubscription(config.Ctx, event.Subscription{
		// Logger:            config.Logger,
		Topics:            topics,
		GroupID:           config.GroupID,
		HTTPHeaders:       httpheaders,
		APIKey:            config.APIKey,
		DisableAutoCommit: true,
		Channel:           config.Channel,
		Filter:            filters,
	})
	if err != nil {
		return nil, err
	}
	s := &Subscriber{
		ch:                  ch,
		cb:                  callback,
		logger:              config.Logger,
		location:            location.String(),
		refType:             config.Integration.Descriptor.RefType,
		metricServiceName:   metricServiceName,
		metricOperationName: metricOperationName,
	}
	ch.WaitForReady()
	go s.run()
	return s, nil
}
