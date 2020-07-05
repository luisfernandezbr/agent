package dev

import (
	"fmt"
	gohttp "net/http"
	"net/url"
	"os"
	"path/filepath"

	"github.com/jhaynie/go-vcr/v2/recorder"
	"github.com/pinpt/agent.next/internal/graphql"
	"github.com/pinpt/agent.next/internal/http"
	"github.com/pinpt/agent.next/sdk"
	"github.com/pinpt/go-common/v10/api"
	"github.com/pinpt/go-common/v10/fileutil"
	"github.com/pinpt/go-common/v10/httpdefaults"
	"github.com/pinpt/go-common/v10/log"
)

type devManager struct {
	logger    log.Logger
	channel   string
	transport gohttp.RoundTripper
	recorder  *recorder.Recorder
}

var _ sdk.Manager = (*devManager)(nil)

// Close is called on shutdown to cleanup any resources
func (m *devManager) Close() error {
	if m.recorder != nil {
		if err := m.recorder.Stop(); err != nil {
			return err
		}
		m.recorder = nil
	}
	return nil
}

// GraphQLManager returns a graphql manager instance
func (m *devManager) GraphQLManager() sdk.GraphQLClientManager {
	return graphql.New(m.transport)
}

// HTTPManager returns a HTTP manager instance
func (m *devManager) HTTPManager() sdk.HTTPClientManager {
	return http.New(m.transport)
}

// CreateWebHook is used by the integration to create a webhook on behalf of the integration for a given customer and refid
func (m *devManager) CreateWebHook(customerID string, integrationInstanceID string, refType string, refID string) (string, error) {
	return "", fmt.Errorf("cannot create a webhook in dev mode")
}

// RefreshOAuth2Token will refresh the OAuth2 access token using the provided refreshToken and return a new access token
func (m *devManager) RefreshOAuth2Token(refType string, refreshToken string) (string, error) {
	theurl := api.BackendURL(api.AuthService, m.channel)
	theurl += fmt.Sprintf("oauth/%s/refresh/%s", refType, url.PathEscape(refreshToken))
	var res struct {
		AccessToken string `json:"access_token"`
	}
	client := http.New(m.transport).New(theurl, map[string]string{"Content-Type": "application/json"})
	_, err := client.Get(&res)
	if err != nil {
		return "", err
	}
	return res.AccessToken, nil
}

// New will create a new dev sdk.Manager
func New(logger log.Logger, channel string, recordDir, replayDir string) (m sdk.Manager, err error) {
	var transport gohttp.RoundTripper
	var r *recorder.Recorder
	name := "agent_" + channel
	if recordDir != "" {
		recordDir, _ := filepath.Abs(recordDir)
		os.RemoveAll(recordDir)
		os.MkdirAll(recordDir, 0700)
		fn := filepath.Join(recordDir, name)
		r, err = recorder.New(fn)
		if err != nil {
			return nil, err
		}
		transport = r
		r.SetTransport(httpdefaults.DefaultTransport())
	} else if replayDir != "" {
		replayDir, _ := filepath.Abs(replayDir)
		fn := filepath.Join(replayDir, name)
		if !fileutil.FileExists(fn) {
			return nil, fmt.Errorf("missing replay file at %s", fn)
		}
		r, err = recorder.New(fn)
		if err != nil {
			return nil, err
		}
		transport = r
		r.SetTransport(httpdefaults.DefaultTransport())
	} else {
		transport = httpdefaults.DefaultTransport()
	}
	return &devManager{logger, channel, transport, r}, nil
}
