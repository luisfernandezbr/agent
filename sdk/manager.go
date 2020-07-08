package sdk

// Manager is the control interface that integrations can use to perform tasks on its behalf
type Manager interface {
	// GraphQLManager returns a graphql manager instance
	GraphQLManager() GraphQLClientManager
	// HTTPManager returns a HTTP manager instance
	HTTPManager() HTTPClientManager
	// WebHookManager returns the WebHook manager instance
	WebHookManager() WebHookManager
	// AuthManager returns the Auth manager instance
	AuthManager() AuthManager
	// Close is called on shutdown to cleanup any resources
	Close() error
}
