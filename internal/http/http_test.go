package http

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/pinpt/agent/v4/sdk"
	"github.com/pinpt/go-common/v10/httpdefaults"
	pjson "github.com/pinpt/go-common/v10/json"
	"github.com/stretchr/testify/assert"
)

func TestHTTPGetRequest(t *testing.T) {
	assert := assert.New(t)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintln(w, `{"a":"b"}`)
	}))
	defer ts.Close()
	mgr := New(httpdefaults.DefaultTransport())
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
	mgr := New(httpdefaults.DefaultTransport())
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
	mgr := New(httpdefaults.DefaultTransport())
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
	mgr := New(httpdefaults.DefaultTransport())
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
		assert.Equal(http.MethodPost, r.Method)
		var buf bytes.Buffer
		io.Copy(&buf, r.Body)
		w.Write(buf.Bytes())
	}))
	defer ts.Close()
	mgr := New(httpdefaults.DefaultTransport())
	cl := mgr.New(ts.URL, nil)
	kv := make(map[string]interface{})
	resp, err := cl.Post(bytes.NewBuffer([]byte(`{"a":"b"}`)), &kv)
	assert.NoError(err)
	assert.NotNil(resp)
	assert.Equal(http.StatusOK, resp.StatusCode)
	assert.Equal("application/json", resp.Headers.Get("Content-Type"))
	assert.Equal("b", kv["a"])
}

func TestHTTPPutRequest(t *testing.T) {
	assert := assert.New(t)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		assert.Equal(http.MethodPut, r.Method)
		var buf bytes.Buffer
		io.Copy(&buf, r.Body)
		w.Write(buf.Bytes())
	}))
	defer ts.Close()
	mgr := New(httpdefaults.DefaultTransport())
	cl := mgr.New(ts.URL, nil)
	kv := make(map[string]interface{})
	resp, err := cl.Put(bytes.NewBuffer([]byte(`{"a":"b"}`)), &kv)
	assert.NoError(err)
	assert.NotNil(resp)
	assert.Equal(http.StatusOK, resp.StatusCode)
	assert.Equal("application/json", resp.Headers.Get("Content-Type"))
	assert.Equal("b", kv["a"])
}

func TestHTTPPutRequestNoContent(t *testing.T) {
	assert := assert.New(t)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		assert.Equal(http.MethodPut, r.Method)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer ts.Close()
	mgr := New(httpdefaults.DefaultTransport())
	cl := mgr.New(ts.URL, nil)
	kv := make(map[string]interface{})
	resp, err := cl.Put(bytes.NewBuffer([]byte(`{"a":"b"}`)), &kv)
	assert.NoError(err)
	assert.NotNil(resp)
	assert.Equal(http.StatusNoContent, resp.StatusCode)
}

func TestHTTPPatchRequest(t *testing.T) {
	assert := assert.New(t)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		assert.Equal(http.MethodPatch, r.Method)
		var buf bytes.Buffer
		io.Copy(&buf, r.Body)
		w.Write(buf.Bytes())
	}))
	defer ts.Close()
	mgr := New(httpdefaults.DefaultTransport())
	cl := mgr.New(ts.URL, nil)
	kv := make(map[string]interface{})
	resp, err := cl.Patch(bytes.NewBuffer([]byte(`{"a":"b"}`)), &kv)
	assert.NoError(err)
	assert.NotNil(resp)
	assert.Equal(http.StatusOK, resp.StatusCode)
	assert.Equal("application/json", resp.Headers.Get("Content-Type"))
	assert.Equal("b", kv["a"])
}

func TestHTTPDeleteRequest(t *testing.T) {
	assert := assert.New(t)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(http.MethodDelete, r.Method)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(pjson.Stringify(map[string]string{"a": "b"})))
	}))
	defer ts.Close()
	mgr := New(httpdefaults.DefaultTransport())
	cl := mgr.New(ts.URL, nil)
	kv := make(map[string]interface{})
	resp, err := cl.Delete(&kv)
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
		w.Header().Set("Retry-After", "1")
		w.WriteHeader(http.StatusTooManyRequests)
		count++
	}))
	defer ts.Close()
	mgr := New(httpdefaults.DefaultTransport())
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
			w.Header().Set("Retry-After", "6")
			w.WriteHeader(http.StatusOK)
			w.Write(buf.Bytes())
			return
		}
		w.Header().Set("Retry-After", "1")
		w.WriteHeader(http.StatusTooManyRequests)
		count++
	}))
	defer ts.Close()
	mgr := New(httpdefaults.DefaultTransport())
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
		w.Header().Set("Retry-After", "2")
		w.WriteHeader(http.StatusTooManyRequests)
		count++
	}))
	defer ts.Close()
	mgr := New(httpdefaults.DefaultTransport())
	cl := mgr.New(ts.URL, nil)
	kv := make(map[string]interface{})
	resp, err := cl.Post(bytes.NewBuffer([]byte(`{"a":"b"}`)), &kv, sdk.WithDeadline(time.Second))
	assert.Error(err, sdk.ErrTimedOut)
	assert.Nil(resp)
	assert.True(count > 0)
}

func TestHTTPGetWithEndpoint(t *testing.T) {

	// testing endpoints. For example:
	//     base url: https://www.googleapis.com/calendar/v3
	//     endpoint: /users/me/calendarList
	// complete url: https://www.googleapis.com/calendar/v3/users/me/calendarList

	assert := assert.New(t)
	mux := http.DefaultServeMux

	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"url":"` + r.URL.Path + `"}`))
	}
	mux.HandleFunc("/", handler)
	mux.HandleFunc("/foo", handler)
	mux.HandleFunc("/bar", handler)
	mux.HandleFunc("/hello/world", handler)

	ts := httptest.NewServer(mux)

	defer ts.Close()
	mgr := New(httpdefaults.DefaultTransport())

	cl := mgr.New(ts.URL, nil)
	kv := make(map[string]interface{})
	resp, err := cl.Get(&kv, sdk.WithEndpoint("foo"))
	assert.NoError(err)
	assert.NotNil(resp)
	assert.Equal(http.StatusOK, resp.StatusCode)
	assert.Equal("/foo", kv["url"])

	resp, err = cl.Get(&kv, sdk.WithEndpoint("bar"))
	assert.NoError(err)
	assert.NotNil(resp)
	assert.Equal(http.StatusOK, resp.StatusCode)
	assert.Equal("/bar", kv["url"])

	resp, err = cl.Get(&kv, sdk.WithEndpoint("hello/world"))
	assert.NoError(err)
	assert.NotNil(resp)
	assert.Equal(http.StatusOK, resp.StatusCode)
	assert.Equal("/hello/world", kv["url"])

	resp, err = cl.Get(&kv)
	assert.NoError(err)
	assert.NotNil(resp)
	assert.Equal(http.StatusOK, resp.StatusCode)
	assert.Equal("/", kv["url"])
}

func TestHTTPBasicAuth(t *testing.T) {
	assert := assert.New(t)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"auth": "` + r.Header.Get("Authorization") + `"}`))
	}))
	defer ts.Close()
	username := "pinpoint"
	password := "rocks!"
	mgr := New(httpdefaults.DefaultTransport())
	cl := mgr.New(ts.URL, nil)
	var out struct {
		Auth string `json:"auth"`
	}
	_, err := cl.Get(&out, sdk.WithBasicAuth(username, password))
	assert.NoError(err)
	assert.Equal("Basic "+base64.StdEncoding.EncodeToString([]byte(username+":"+password)), out.Auth)
}

type fakeOAuthManager struct {
}

var _ sdk.Manager = (*fakeOAuthManager)(nil)

func (f *fakeOAuthManager) Close() error                             { return nil }
func (f *fakeOAuthManager) GraphQLManager() sdk.GraphQLClientManager { return nil }
func (f *fakeOAuthManager) HTTPManager() sdk.HTTPClientManager       { return nil }
func (f *fakeOAuthManager) WebHookManager() sdk.WebHookManager       { return nil }
func (f *fakeOAuthManager) AuthManager() sdk.AuthManager             { return f }
func (f *fakeOAuthManager) UserManager() sdk.UserManager             { return nil }
func (f *fakeOAuthManager) CreateWebHook(customerID string, refType string, integrationInstanceID string, refID string) (string, error) {
	return "", nil
}
func (f *fakeOAuthManager) RefreshOAuth2Token(refType string, refreshToken string) (string, error) {
	return "NEW_TOKEN " + refreshToken, nil
}
func (f *fakeOAuthManager) PrivateKey(identifier sdk.Identifier) (*rsa.PrivateKey, error) {
	return rsa.GenerateKey(rand.Reader, 1024)
}

func TestHTTPOAuth(t *testing.T) {
	assert := assert.New(t)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"auth": "` + r.Header.Get("Authorization") + `"}`))
	}))
	defer ts.Close()
	mgr := New(httpdefaults.DefaultTransport())
	cl := mgr.New(ts.URL, nil)
	var out struct {
		Auth string `json:"auth"`
	}
	_, err := cl.Get(&out, sdk.WithOAuth2Refresh(&fakeOAuthManager{}, "foo", "12345TOKEN67890", "12345REFRESH_TOKEN67890"))
	assert.NoError(err)
	assert.Equal("Bearer 12345TOKEN67890", out.Auth)
}

func TestHTTPOAuthRefresh(t *testing.T) {
	assert := assert.New(t)
	shouldfail := true
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if shouldfail {
			shouldfail = false
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"auth": "` + r.Header.Get("Authorization") + `"}`))
	}))
	defer ts.Close()
	token := "12345TOKEN67890"
	refreshToken := "12345REFRESH_TOKEN67890"
	mgr := New(httpdefaults.DefaultTransport())
	cl := mgr.New(ts.URL, nil)
	var out struct {
		Auth string `json:"auth"`
	}
	_, err := cl.Get(&out, sdk.WithOAuth2Refresh(&fakeOAuthManager{}, "foo", token, refreshToken))
	assert.NoError(err)
	assert.Equal("Bearer NEW_TOKEN "+refreshToken, out.Auth)
}

func TestHTTPOAuthTooMany(t *testing.T) {
	assert := assert.New(t)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer ts.Close()
	token := "12345TOKEN67890"
	refreshToken := "12345REFRESH_TOKEN67890"
	mgr := New(httpdefaults.DefaultTransport())
	cl := mgr.New(ts.URL, nil)
	var out struct {
		Auth string `json:"auth"`
	}
	resp, err := cl.Get(&out, sdk.WithOAuth2Refresh(&fakeOAuthManager{}, "foo", token, refreshToken))
	assert.Error(err)
	assert.Equal(http.StatusUnauthorized, resp.StatusCode)
}

func TestHTTPOAuth1(t *testing.T) {
	assert := assert.New(t)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		val := r.Header.Get("Authorization")
		if strings.Contains(val, "OAuth oauth_consumer_key=") {
			w.WriteHeader(http.StatusOK)
			return
		}
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer ts.Close()
	mgr := New(httpdefaults.DefaultTransport())
	cl := mgr.New(ts.URL, nil)
	var out struct {
		Auth string `json:"auth"`
	}
	resp, err := cl.Get(&out, sdk.WithOAuth1(&fakeOAuthManager{}, sdk.NewSimpleIdentifier("1234", "1", "reftype"), "consumerkey", "consumersecret", "token", "secret"))
	assert.NoError(err)
	assert.Equal(http.StatusOK, resp.StatusCode)
}
