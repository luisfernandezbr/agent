package dev

import (
	"github.com/pinpt/agent.next/internal/graphql"
	"github.com/pinpt/agent.next/sdk"
	"github.com/pinpt/go-common/log"
)

type eventAPIManager struct {
	logger log.Logger
}

var _ sdk.Manager = (*eventAPIManager)(nil)

// GraphQLManager returns a graphql manager instance
func (m *eventAPIManager) GraphQLManager() sdk.GraphQLClientManager {
	return graphql.New()
}

// CreateWebHook is used by the integration to create a webhook on behalf of the integration for a given customer and refid
func (m *eventAPIManager) CreateWebHook(customerID string, refType string, refID string) (string, error) {
	// FIXME: todo
	return "", nil
}

// New will create a new dev sdk.Manager
func New(logger log.Logger) sdk.Manager {
	return &eventAPIManager{logger}
}
