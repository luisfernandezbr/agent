package sdk

import (
	"net/http"
	"time"
)

// WithGraphQLOption is an option for setting details on the request
type WithGraphQLOption func(req *http.Request) error

// GraphQLClientManager is an interface for creating graphql clients
type GraphQLClientManager interface {
	// New is for creating a new graphql client instance that can be reused
	New(url string, headers map[string]string) GraphQLClient
}

// RateLimitError is a specific error for detection of rate limit errors
type RateLimitError struct {
	RetryAfter time.Duration
}

func (e *RateLimitError) Error() string {
	return "rate limited"
}

// IsRateLimitError returns true if an error is a rate limit error and if so, the retry after duration
func IsRateLimitError(err error) (bool, time.Duration) {
	if re, ok := err.(*RateLimitError); ok {
		return true, re.RetryAfter
	}
	return false, 0
}

// GraphQLClient is an interface to a graphql client
type GraphQLClient interface {
	Query(query string, variables map[string]interface{}, out interface{}, options ...WithGraphQLOption) error
}

// WithGraphQLHeader will add a specific header to an outgoing request
func WithGraphQLHeader(key, value string) WithGraphQLOption {
	return func(req *http.Request) error {
		req.Header.Set(key, value)
		return nil
	}
}
