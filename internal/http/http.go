package http

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/mailru/easyjson"
	"github.com/pinpt/agent.next/sdk"
	"github.com/pinpt/go-common/v10/event"
)

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
	url       string
	headers   map[string]string
	cl        *http.Client
	transport http.RoundTripper
}

var _ sdk.HTTPClient = (*client)(nil)

func (c *client) exec(opt *sdk.HTTPOptions, out interface{}, options ...sdk.WithHTTPOption) (*sdk.HTTPResponse, error) {
	c.cl.Transport = opt.Transport // reset it each time in case it changed
	resp, err := c.cl.Do(opt.Request)
	if err != nil {
		return nil, err
	}
	res := &sdk.HTTPResponse{
		StatusCode: resp.StatusCode,
		Headers:    resp.Header,
	}
	opt.Response = res
	for _, o := range options {
		if o != nil {
			if err := o(opt); err != nil {
				return nil, err
			}
		}
	}
	opt.Response = nil
	// no content means there's no body
	if resp.StatusCode == http.StatusNoContent {
		return res, nil
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
		opt.ShouldRetry = true
		opt.RetryAfter = tv
		return nil, nil
	}
	// read the body
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, resp.Body); err != nil {
		return nil, fmt.Errorf("error copying response body: %w", err)
	}
	resp.Body.Close()
	res.Body = buf.Bytes()
	if resp.StatusCode > 299 {
		return res, &sdk.HTTPError{
			StatusCode: resp.StatusCode,
			Body:       &buf,
		}
	}
	if out == nil {
		return res, nil
	}
	if strings.Contains(resp.Header.Get("Content-Type"), "json") {
		if i, ok := out.(easyjson.Unmarshaler); ok {
			err := easyjson.Unmarshal(buf.Bytes(), i)
			return res, err
		}
		if err := json.NewDecoder(&buf).Decode(out); err != nil {
			return res, err
		}
	}
	return res, nil
}

func (c *client) makeRequest(req *http.Request, deadline time.Time, options ...sdk.WithHTTPOption) (*sdk.HTTPOptions, error) {
	transport := c.transport
	if transport == nil {
		transport = http.DefaultTransport
	}
	opts := &sdk.HTTPOptions{
		Request:   req,
		Deadline:  deadline,
		Transport: transport,
	}
	opts.Request.Header.Set("Accept", "application/json")
	opts.Request.Header.Set("Content-Type", "application/json")
	opts.Request.Header.Set("User-Agent", "pinpoint.com")
	for k, v := range c.headers {
		opts.Request.Header.Set(k, v)
	}
	for _, opt := range options {
		if opt != nil {
			if err := opt(opts); err != nil {
				return nil, err
			}
		}
	}
	return opts, nil
}

const backoffRange = 200

type requestMaker func() (*http.Request, error)

func isStatusRetryable(status int) bool {
	switch status {
	case http.StatusBadGateway, http.StatusGatewayTimeout, http.StatusServiceUnavailable, http.StatusTooManyRequests:
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
		resp, err := c.exec(httpreq, out, options...)
		if httpreq.ShouldRetry || event.IsErrorRetryable(err) || (resp != nil && isStatusRetryable(resp.StatusCode)) {
			if time.Now().Before(httpreq.Deadline) {
				if httpreq.RetryAfter > 0 {
					// retry after our header tells us
					time.Sleep(httpreq.RetryAfter)
				} else {
					// do an expotential backoff
					time.Sleep(time.Millisecond * time.Duration(int64(i)*rand.Int63n(backoffRange)))
				}
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

// Put will call a HTTP PUT method passing the data and set the result (if JSON) to out
func (c *client) Put(data io.Reader, out interface{}, options ...sdk.WithHTTPOption) (*sdk.HTTPResponse, error) {
	var buf bytes.Buffer
	io.Copy(&buf, data)
	rw := &rewindReader{
		buf: buf.Bytes(),
	}
	return c.execWithRetry(func() (*http.Request, error) {
		rw.Rewind()
		return http.NewRequest(http.MethodPut, c.url, rw)
	}, out, options...)
}

// Patch will call a HTTP PATCH method passing the data and set the result (if JSON) to out
func (c *client) Patch(data io.Reader, out interface{}, options ...sdk.WithHTTPOption) (*sdk.HTTPResponse, error) {
	var buf bytes.Buffer
	io.Copy(&buf, data)
	rw := &rewindReader{
		buf: buf.Bytes(),
	}
	return c.execWithRetry(func() (*http.Request, error) {
		rw.Rewind()
		return http.NewRequest(http.MethodPatch, c.url, rw)
	}, out, options...)
}

// Post will call a HTTP DELETE method and set the result (if JSON) to out
func (c *client) Delete(out interface{}, options ...sdk.WithHTTPOption) (*sdk.HTTPResponse, error) {
	return c.execWithRetry(func() (*http.Request, error) {
		return http.NewRequest(http.MethodDelete, c.url, nil)
	}, out, options...)
}

type manager struct {
	transport http.RoundTripper
}

var _ sdk.HTTPClientManager = (*manager)(nil)

// New is for creating a new HTTP client instance that can be reused
func (m *manager) New(url string, headers map[string]string) sdk.HTTPClient {
	return &client{
		url:       url,
		headers:   headers,
		cl:        http.DefaultClient,
		transport: m.transport,
	}
}

// New returns a new HTTPClientManager
func New(transport http.RoundTripper) sdk.HTTPClientManager {
	return &manager{transport}
}
