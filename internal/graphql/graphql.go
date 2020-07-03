package graphql

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/mailru/easyjson"
	"github.com/pinpt/agent.next/sdk"
)

type client struct {
	url     string
	headers map[string]string
}

var _ sdk.GraphQLClient = (*client)(nil)

func (g *client) Query(query string, variables map[string]interface{}, out interface{}, options ...sdk.WithGraphQLOption) error {
	payload := struct {
		Variables map[string]interface{} `json:"variables"`
		Query     string                 `json:"query"`
	}{variables, query}
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, g.url, bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	for k, v := range g.headers {
		req.Header.Set(k, v)
	}
	for _, opt := range options {
		if err := opt(req); err != nil {
			return err
		}
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// check to see if we've been rate limited with the Retry-After header
	val := resp.Header.Get("Retry-After")
	if val != "" {
		io.Copy(ioutil.Discard, resp.Body)
		secs, _ := strconv.ParseInt(val, 10, 32)
		return &sdk.RateLimitError{
			RetryAfter: time.Second * time.Duration(secs),
		}
	}

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return err
	}

	if resp.StatusCode == http.StatusOK {
		var datares struct {
			Data   json.RawMessage `json:"data"`
			Errors []struct {
				Message string `json:"message"`
			} `json:"errors"`
		}
		err = json.Unmarshal(body, &datares)
		if err != nil {
			return err
		}
		if len(datares.Errors) > 0 {
			if len(datares.Errors) == 1 && datares.Errors[0].Message != "" {
				return errors.New(datares.Errors[0].Message)
			}
			b, err := json.Marshal(datares.Errors)
			if err != nil {
				return err
			}
			return errors.New(string(b))
		}
		if i, ok := out.(easyjson.Unmarshaler); ok {
			return easyjson.Unmarshal(datares.Data, i)
		}
		return json.Unmarshal(datares.Data, out)
	}
	return fmt.Errorf("err: %s. status code: %s", string(body), resp.Status)
}

type manager struct {
}

var _ sdk.GraphQLClientManager = (*manager)(nil)

// New is for creating a new graphql client instance that can be reused
func (m *manager) New(url string, headers map[string]string) sdk.GraphQLClient {
	return &client{url, headers}
}

// New returns a new GraphQLClientManager
func New() sdk.GraphQLClientManager {
	return &manager{}
}
