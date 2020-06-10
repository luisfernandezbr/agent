package server

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/pinpt/go-common/v10/datamodel"
	"github.com/pinpt/go-common/v10/event"
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
func createDBChangeSubscriptionFilter(refType string) *event.SubscriptionFilter {
	modelexpr := []string{}
	for _, model := range models {
		modelexpr = append(modelexpr, fmt.Sprintf(`model:"%s"`, model))
	}
	return &event.SubscriptionFilter{
		HeaderExpr: "(" + strings.Join(modelexpr, " OR ") + ")",
		ObjectExpr: fmt.Sprintf(`ref_type:"%s"`, refType),
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
	ch *event.SubscriptionChannel
	cb SubscriberCallback
}

func (s *Subscriber) run() {
	for event := range s.ch.Channel() {
		s.cb(event)
	}
}

// Close will close the subscriber
func (s *Subscriber) Close() error {
	return s.ch.Close()
}

// SubscriberCallback is the callback for processing events
type SubscriberCallback func(event event.SubscriptionEvent) error

// NewDBChangeSubscriber will return a db change subscriber
func NewDBChangeSubscriber(config Config, callback SubscriberCallback) (*Subscriber, error) {
	headers := map[string]string{}
	httpheaders := map[string]string{}
	if config.Secret != "" {
		httpheaders["x-api-key"] = config.Secret
	}
	if config.UUID != "" {
		headers["uuid"] = config.UUID
	}
	ch, err := event.NewSubscription(config.Ctx, event.Subscription{
		Logger:            config.Logger,
		Topics:            []string{"ops.db.Change"},
		GroupID:           config.GroupID,
		HTTPHeaders:       httpheaders,
		APIKey:            config.APIKey,
		DisableAutoCommit: true,
		Channel:           config.Channel,
		DisablePing:       true,
		Headers:           headers,
		Filter:            createDBChangeSubscriptionFilter(config.Integration.Descriptor.RefType),
	})
	if err != nil {
		return nil, err
	}
	s := &Subscriber{
		ch: ch,
		cb: callback,
	}
	go s.run()
	return s, nil
}

// NewEventSubscriber will return an event subscriber
func NewEventSubscriber(config Config, topics []string, callback SubscriberCallback) (*Subscriber, error) {
	headers := map[string]string{}
	httpheaders := map[string]string{}
	if config.Secret != "" {
		httpheaders["x-api-key"] = config.Secret
	}
	if config.UUID != "" {
		headers["uuid"] = config.UUID
	}
	ch, err := event.NewSubscription(config.Ctx, event.Subscription{
		Logger:            config.Logger,
		Topics:            topics,
		GroupID:           config.GroupID,
		HTTPHeaders:       httpheaders,
		APIKey:            config.APIKey,
		DisableAutoCommit: true,
		Channel:           config.Channel,
		DisablePing:       true,
		Headers:           headers,
		Filter: &event.SubscriptionFilter{
			ObjectExpr: fmt.Sprintf(`ref_type:"%s"`, config.Integration.Descriptor.RefType),
		},
	})
	if err != nil {
		return nil, err
	}
	s := &Subscriber{
		ch: ch,
		cb: callback,
	}
	go s.run()
	return s, nil
}
