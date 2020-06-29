package http

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/pinpt/agent.next/sdk"
	"github.com/pinpt/go-common/v10/event"
	"github.com/pinpt/go-common/v10/httpdefaults"
)

var transport = httpdefaults.DefaultTransport()

type rewindReader struct {
	buf []byte
	r   bufio.Reader
}

var _ io.Reader = (*rewindReader)(nil)

func (r *rewindReader) Rewind() {
	r.r.Reset(bufio.NewReader(bytes.NewReader(r.buf)))
}

func (r *rewindReader) Read(p []byte) (int, error) {
	return r.r.Read(p)
}

type client struct {
	url     string
	headers map[string]string
	cl      *http.Client
}

var _ sdk.HTTPClient = (*client)(nil)

func (c *client) exec(req *sdk.HTTPRequest, out interface{}) (*sdk.HTTPResponse, error) {

	if req.Creds != nil {
		// set this here in case we need to refresh the token,
		// the Auth func will have the new token
		req.Request.Header.Set("Authorization", req.Creds.Auth())
	}
	resp, err := http.DefaultClient.Do(req.Request)
	if err != nil {
		return nil, err
	}
	res := &sdk.HTTPResponse{
		StatusCode: resp.StatusCode,
		Headers:    resp.Header,
	}
	// check to see if this was a rate limited response
	if resp.StatusCode == http.StatusTooManyRequests {
		val := resp.Header.Get("Retry-After")
		tv := 30 * time.Second // if we don't get any header back, pick a value
		if val != "" {
			v, _ := strconv.ParseInt(val, 10, 64)
			if v > 0 {
				tv = time.Second * time.Duration(v)
			}
		}
		return res, &sdk.RateLimitError{
			RetryAfter: tv,
		}
	}
	// if unauthorized and oauth, call refresh
	if resp.StatusCode == http.StatusUnauthorized && req.Creds != nil {
		if creds, ok := req.Creds.(*sdk.HTTPOAuthCreds); ok {
			// if the last time we refresh the token was less then a minute, then something is wrong
			// only refresh if the last time was a while ago, and then try again
			if time.Since(creds.LastRetry) > (1 * time.Minute) {
				if err := creds.Refresh(); err != nil {
					return nil, err
				}
				creds.LastRetry = time.Now()
				return c.exec(req, out)
			}
		}
	}

	if resp.StatusCode != http.StatusOK {
		var buf bytes.Buffer
		io.Copy(&buf, resp.Body)
		resp.Body.Close()
		return res, &sdk.HTTPError{
			StatusCode: resp.StatusCode,
			Body:       &buf,
		}
	}
	if strings.Contains(resp.Header.Get("Content-Type"), "json") {
		if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
			return res, err
		}
	}
	return res, nil
}

func (c *client) makeRequest(req *http.Request, deadline time.Time, options ...sdk.WithHTTPOption) (*sdk.HTTPRequest, error) {
	httpreq := &sdk.HTTPRequest{
		Request:  req,
		Deadline: deadline,
	}
	httpreq.Request.Header.Set("Accept", "application/json")
	httpreq.Request.Header.Set("Content-Type", "application/json")
	httpreq.Request.Header.Set("User-Agent", "pinpoint.com")
	for k, v := range c.headers {
		httpreq.Request.Header.Set(k, v)
	}
	for _, opt := range options {
		if opt != nil {
			if err := opt(httpreq); err != nil {
				return nil, err
			}
		}
	}
	return httpreq, nil
}

const backoffRange = 200

type requestMaker func() (*http.Request, error)

func isStatusRetryable(status int) bool {
	switch status {
	case http.StatusBadGateway, http.StatusGatewayTimeout, http.StatusServiceUnavailable:
		return true
	default:
		return false
	}
}

func (c *client) execWithRetry(maker requestMaker, out interface{}, options ...sdk.WithHTTPOption) (*sdk.HTTPResponse, error) {
	defaultDeadline := time.Now().Add(time.Minute) // default
	var i int
	for {
		req, err := maker()
		if err != nil {
			return nil, err
		}
		httpreq, err := c.makeRequest(req, defaultDeadline, options...)
		if err != nil {
			return nil, err
		}
		i++
		resp, err := c.exec(httpreq, out)
		if event.IsErrorRetryable(err) || (resp != nil && isStatusRetryable(resp.StatusCode)) {
			if time.Now().Before(httpreq.Deadline) {
				// do an expotential backoff
				time.Sleep(time.Millisecond * time.Duration(int64(i)*rand.Int63n(backoffRange)))
			}
			// check again
			if time.Now().Before(httpreq.Deadline) {
				continue
			}
			return nil, sdk.ErrTimedOut
		}
		return resp, err
	}
}

// Get will call a HTTP GET method and set the result (if JSON) to out
func (c *client) Get(out interface{}, options ...sdk.WithHTTPOption) (*sdk.HTTPResponse, error) {
	return c.execWithRetry(func() (*http.Request, error) {
		return http.NewRequest(http.MethodGet, c.url, nil)
	}, out, options...)
}

// Post will call a HTTP POST method passing the data and set the result (if JSON) to out
func (c *client) Post(data io.Reader, out interface{}, options ...sdk.WithHTTPOption) (*sdk.HTTPResponse, error) {
	var buf bytes.Buffer
	io.Copy(&buf, data)
	rw := &rewindReader{
		buf: buf.Bytes(),
	}
	return c.execWithRetry(func() (*http.Request, error) {
		rw.Rewind()
		return http.NewRequest(http.MethodPost, c.url, rw)
	}, out, options...)
}

type manager struct {
}

var _ sdk.HTTPClientManager = (*manager)(nil)

// New is for creating a new HTTP client instance that can be reused
func (m *manager) New(url string, headers map[string]string) sdk.HTTPClient {
	return &client{
		url:     url,
		headers: headers,
		cl:      &http.Client{Transport: transport},
	}
}

// New returns a new HTTPClientManager
func New() sdk.HTTPClientManager {
	return &manager{}
}
