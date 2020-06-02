package http

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/pinpt/agent.next/sdk"
	"github.com/stretchr/testify/assert"
)

func TestHTTPGetRequest(t *testing.T) {
	assert := assert.New(t)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `{"a":"b"}`)
	}))
	defer ts.Close()
	mgr := New()
	cl := mgr.New(ts.URL, nil)
	kv := make(map[string]interface{})
	resp, err := cl.Get(&kv)
	assert.NoError(err)
	assert.NotNil(resp)
	assert.Equal(http.StatusOK, resp.StatusCode)
	assert.Equal("application/json", resp.Headers.Get("Content-Type"))
	assert.Equal("b", kv["a"])
}

func TestHTTPGetRequestInitialHeader(t *testing.T) {
	assert := assert.New(t)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Bar", r.Header.Get("Foo"))
		fmt.Fprintln(w, `{"a":"b"}`)
	}))
	defer ts.Close()
	mgr := New()
	cl := mgr.New(ts.URL, map[string]string{"Foo": "Bar"})
	kv := make(map[string]interface{})
	resp, err := cl.Get(&kv)
	assert.NoError(err)
	assert.NotNil(resp)
	assert.Equal(http.StatusOK, resp.StatusCode)
	assert.Equal("application/json", resp.Headers.Get("Content-Type"))
	assert.Equal("Bar", resp.Headers.Get("Bar"))
	assert.Equal("b", kv["a"])
}

func TestHTTPGetRequestOverrideHeader(t *testing.T) {
	assert := assert.New(t)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Bar", r.Header.Get("Foo"))
		fmt.Fprintln(w, `{"a":"b"}`)
	}))
	defer ts.Close()
	mgr := New()
	cl := mgr.New(ts.URL, map[string]string{"Foo": "Bar"})
	kv := make(map[string]interface{})
	resp, err := cl.Get(&kv, sdk.WithHTTPHeader("Foo", "Foo"))
	assert.NoError(err)
	assert.NotNil(resp)
	assert.Equal(http.StatusOK, resp.StatusCode)
	assert.Equal("application/json", resp.Headers.Get("Content-Type"))
	assert.Equal("Foo", resp.Headers.Get("Bar"))
	assert.Equal("b", kv["a"])
}

func TestHTTPGetRequestError(t *testing.T) {
	assert := assert.New(t)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()
	mgr := New()
	cl := mgr.New(ts.URL, nil)
	kv := make(map[string]interface{})
	resp, err := cl.Get(&kv, nil)
	assert.Error(err)
	assert.NotNil(resp)
	assert.Equal(http.StatusNotFound, resp.StatusCode)
	assert.True(sdk.IsHTTPError(err))
}

func TestHTTPPostRequest(t *testing.T) {
	assert := assert.New(t)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		var buf bytes.Buffer
		io.Copy(&buf, r.Body)
		w.Write(buf.Bytes())
	}))
	defer ts.Close()
	mgr := New()
	cl := mgr.New(ts.URL, nil)
	kv := make(map[string]interface{})
	resp, err := cl.Post(bytes.NewBuffer([]byte(`{"a":"b"}`)), &kv)
	assert.NoError(err)
	assert.NotNil(resp)
	assert.Equal(http.StatusOK, resp.StatusCode)
	assert.Equal("application/json", resp.Headers.Get("Content-Type"))
	assert.Equal("b", kv["a"])
}

func TestHTTPGetRetry(t *testing.T) {
	assert := assert.New(t)
	var count int
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if count >= 5 {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"a":"b"}`))
			return
		}
		w.WriteHeader(http.StatusTooManyRequests)
		count++
	}))
	defer ts.Close()
	mgr := New()
	cl := mgr.New(ts.URL, nil)
	kv := make(map[string]interface{})
	resp, err := cl.Get(&kv, sdk.WithDeadline(time.Second*5))
	assert.NoError(err)
	assert.NotNil(resp)
	assert.Equal(http.StatusOK, resp.StatusCode)
	assert.Equal("application/json", resp.Headers.Get("Content-Type"))
	assert.Equal("b", kv["a"])
	assert.True(count >= 5)
}

func TestHTTPPostRetry(t *testing.T) {
	assert := assert.New(t)
	var count int
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if count > 5 {
			w.Header().Set("Content-Type", "application/json")
			var buf bytes.Buffer
			io.Copy(&buf, r.Body)
			w.WriteHeader(http.StatusOK)
			w.Write(buf.Bytes())
			return
		}
		w.WriteHeader(http.StatusTooManyRequests)
		count++
	}))
	defer ts.Close()
	mgr := New()
	cl := mgr.New(ts.URL, nil)
	kv := make(map[string]interface{})
	resp, err := cl.Post(bytes.NewBuffer([]byte(`{"a":"b"}`)), &kv, sdk.WithDeadline(time.Second*5))
	assert.NoError(err)
	assert.NotNil(resp)
	assert.Equal(http.StatusOK, resp.StatusCode)
	assert.Equal("application/json", resp.Headers.Get("Content-Type"))
	assert.Equal("b", kv["a"])
	assert.True(count >= 5)
}

func TestHTTPRetryTimeout(t *testing.T) {
	assert := assert.New(t)
	var count int
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		count++
	}))
	defer ts.Close()
	mgr := New()
	cl := mgr.New(ts.URL, nil)
	kv := make(map[string]interface{})
	resp, err := cl.Post(bytes.NewBuffer([]byte(`{"a":"b"}`)), &kv, sdk.WithDeadline(time.Second))
	assert.Error(err, sdk.ErrTimedOut)
	assert.Nil(resp)
	assert.True(count >= 5)
}
