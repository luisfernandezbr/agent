package sdk

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// ErrTimedOut returns a timeout event when our deadline is exceeded
var ErrTimedOut = errors.New("timeout")

// HTTPOptions is a holder for options
type HTTPOptions struct {
	Request     *http.Request
	Response    *HTTPResponse // only set in the response case or nil in the request case
	Deadline    time.Time
	ShouldRetry bool
	RetryAfter  time.Duration
}

// WithHTTPOption is an option for setting details on the request
type WithHTTPOption func(opt *HTTPOptions) error

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
	// Put will call a HTTP PUT method passing the data and set the result (if JSON) to out
	Put(data io.Reader, out interface{}, options ...WithHTTPOption) (*HTTPResponse, error)
	// Patch will call a HTTP PATCH method passing the data and set the result (if JSON) to out
	Patch(data io.Reader, out interface{}, options ...WithHTTPOption) (*HTTPResponse, error)
	// Delete will call a HTTP DELETE method and set the result (if JSON) to out
	Delete(out interface{}, options ...WithHTTPOption) (*HTTPResponse, error)
}

// WithHTTPHeader will add a specific header to an outgoing request
func WithHTTPHeader(key, value string) WithHTTPOption {
	return func(opt *HTTPOptions) error {
		if opt.Response == nil {
			opt.Request.Header.Set(key, value)
		}
		return nil
	}
}

// WithEndpoint will add to the url path
func WithEndpoint(value string) WithHTTPOption {
	return func(opt *HTTPOptions) error {
		if opt.Response == nil {
			opt.Request.URL.Path = JoinURL(opt.Request.URL.Path, value)
			opt.Request.URL, _ = url.Parse(opt.Request.URL.String())
		}
		return nil
	}
}

// WithContentType will set the Content-Type header
func WithContentType(value string) WithHTTPOption {
	return func(opt *HTTPOptions) error {
		if opt.Response == nil {
			opt.Request.Header.Set("Content-Type", value)
		}
		return nil
	}
}

// WithAuthorization will set the Authorization header
func WithAuthorization(value string) WithHTTPOption {
	return func(opt *HTTPOptions) error {
		if opt.Response == nil {
			opt.Request.Header.Set("Authorization", value)
		}
		return nil
	}
}

// WithGetQueryParameters will allow the query parameters to be overriden
func WithGetQueryParameters(variables url.Values) WithHTTPOption {
	return func(opt *HTTPOptions) error {
		if opt.Response == nil {
			q := opt.Request.URL.Query()
			for k, v := range variables {
				q[k] = v
			}
			opt.Request.URL.RawQuery = q.Encode()
		}
		return nil
	}
}

// WithDeadline will set a deadline for getting a response
func WithDeadline(duration time.Duration) WithHTTPOption {
	return func(opt *HTTPOptions) error {
		if opt.Response == nil {
			opt.Deadline = time.Now().Add(duration)
		}
		return nil
	}
}

// WithBasicAuth will add the Basic authentication header to the outgoing request
func WithBasicAuth(username string, password string) WithHTTPOption {
	return WithAuthorization("Basic " + base64.StdEncoding.EncodeToString([]byte(username+":"+password)))
}

// WithOAuth2Refresh will set the oauth2 information and support automatic token refresh
func WithOAuth2Refresh(manager Manager, refType string, accessToken string, refreshToken string) WithHTTPOption {
	var lastRetry time.Time
	token := accessToken // capture this in the closure since we can change it on refresh
	return func(opt *HTTPOptions) error {
		if opt.Response == nil {
			opt.Request.Header.Set("Authorization", "Bearer "+token)
			return nil
		}
		if opt.Response.StatusCode == http.StatusUnauthorized && refreshToken != "" {
			var err error
			token, err = manager.AuthManager().RefreshOAuth2Token(refType, refreshToken)
			if err != nil {
				return err
			}
			// if the last time we refresh the token was less then a minute, then something is wrong
			// only refresh if the last time was a while ago, and then try again
			if time.Since(lastRetry) > (1 * time.Minute) {
				opt.ShouldRetry = true
				lastRetry = time.Now()
			}
		}
		return nil
	}
}
