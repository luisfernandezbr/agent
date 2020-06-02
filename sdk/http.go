package sdk

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// ErrTimedOut returns a timeout event when our deadline is exceeded
var ErrTimedOut = errors.New("timeout")

// HTTPRequest is a holder for options
type HTTPRequest struct {
	Request  *http.Request
	Deadline time.Time
}

// WithHTTPOption is an option for setting details on the request
type WithHTTPOption func(req *HTTPRequest) error

// HTTPClientManager is an interface for creating HTTP clients
type HTTPClientManager interface {
	// New is for creating a new HTTP client instance that can be reused
	New(url string, headers map[string]string) HTTPClient
}

// HTTPError is returned if the error is a non-200 status code
type HTTPError struct {
	StatusCode int
	Body       io.Reader
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("HTTP Error: %d", e.StatusCode)
}

// IsHTTPError returns true if an error is a HTTP error
func IsHTTPError(err error) (bool, int, io.Reader) {
	if e, ok := err.(*HTTPError); ok {
		return true, e.StatusCode, e.Body
	}
	return false, 0, nil
}

// HTTPResponse is a struct returned by the HTTPClient
type HTTPResponse struct {
	StatusCode int
	Headers    http.Header
}

// HTTPClient is an interface to a HTTP client
type HTTPClient interface {
	// Get will call a HTTP GET method and set the result (if JSON) to out
	Get(out interface{}, options ...WithHTTPOption) (*HTTPResponse, error)
	// Post will call a HTTP POST method passing the data and set the result (if JSON) to out
	Post(data io.Reader, out interface{}, options ...WithHTTPOption) (*HTTPResponse, error)
}

// WithHTTPHeader will add a specific header to an outgoing request
func WithHTTPHeader(key, value string) WithHTTPOption {
	return func(req *HTTPRequest) error {
		req.Request.Header.Set(key, value)
		return nil
	}
}

// WithContentType will set the Content-Type header
func WithContentType(value string) WithHTTPOption {
	return func(req *HTTPRequest) error {
		req.Request.Header.Set("Content-Type", value)
		return nil
	}
}

// WithAuthorization will set the Authorization header
func WithAuthorization(value string) WithHTTPOption {
	return func(req *HTTPRequest) error {
		req.Request.Header.Set("Authorization", value)
		return nil
	}
}

// WithGetQueryParameters will allow the query parameters to be overriden
func WithGetQueryParameters(variables url.Values) WithHTTPOption {
	return func(req *HTTPRequest) error {
		for k, v := range variables {
			req.Request.URL.Query().Set(k, v[0])
		}
		return nil
	}
}

// WithDeadline will set a deadline for getting a response
func WithDeadline(duration time.Duration) WithHTTPOption {
	return func(req *HTTPRequest) error {
		req.Deadline = time.Now().Add(duration)
		return nil
	}
}
