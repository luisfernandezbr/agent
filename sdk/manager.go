package sdk

// Manager is the control interface that integrations can use to perform tasks on its behalf
type Manager interface {
	// GraphQLManager returns a graphql manager instance
	GraphQLManager() GraphQLClientManager
	// HTTPManager returns a HTTP manager instance
	HTTPManager() HTTPClientManager
	// CreateWebHook is used by the integration to create a webhook on behalf of the integration for a given customer, reftype and refid
	// the result will be a fully qualified URL to the webhook endpoint that should be registered with the integration
	CreateWebHook(customerID string, refType string, integrationInstanceID string, refID string) (string, error)
	// RefreshOAuth2Token will refresh the OAuth2 access token using the provided refreshToken and return a new access token
	RefreshOAuth2Token(refType string, refreshToken string) (string, error)
	// Close is called on shutdown to cleanup any resources
	Close() error
}
