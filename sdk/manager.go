package sdk

// Manager is the control interface that integrations can use to perform tasks on its behalf
type Manager interface {
	// GraphQLManager returns a graphql manager instance
	GraphQLManager() GraphQLClientManager
	// CreateWebHook is used by the integration to create a webhook on behalf of the integration for a given customer and refid
	CreateWebHook(customerID string, refID string) error
}
