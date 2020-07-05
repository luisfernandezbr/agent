package eventapi

import (
	"errors"
	"fmt"
	gohttp "net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/jhaynie/go-vcr/v2/recorder"
	"github.com/pinpt/agent.next/internal/graphql"
	"github.com/pinpt/agent.next/internal/http"
	"github.com/pinpt/agent.next/sdk"
	"github.com/pinpt/go-common/v10/api"
	"github.com/pinpt/go-common/v10/fileutil"
	"github.com/pinpt/go-common/v10/httpdefaults"
	"github.com/pinpt/go-common/v10/log"
)

// ErrWebHookDisabled is returned if webhook is disabled
var ErrWebHookDisabled = errors.New("webhook: disabled")

type eventAPIManager struct {
	logger         log.Logger
	channel        string
	secret         string
	apikey         string
	selfManaged    bool
	webhookEnabled bool
	transport      gohttp.RoundTripper
	recorder       *recorder.Recorder
}

var _ sdk.Manager = (*eventAPIManager)(nil)

// Close is called on shutdown to cleanup any resources
func (m *eventAPIManager) Close() error {
	if m.recorder != nil {
		if err := m.recorder.Stop(); err != nil {
			return err
		}
		m.recorder = nil
	}
	return nil
}

// GraphQLManager returns a graphql manager instance
func (m *eventAPIManager) GraphQLManager() sdk.GraphQLClientManager {
	return graphql.New(m.transport)
}

// HTTPManager returns a HTTP manager instance
func (m *eventAPIManager) HTTPManager() sdk.HTTPClientManager {
	return http.New(m.transport)
}

// CreateWebHook is used by the integration to create a webhook on behalf of the integration for a given customer and refid
func (m *eventAPIManager) CreateWebHook(customerID, refType, integrationInstanceID, refID string) (string, error) {
	if !m.webhookEnabled {
		return "", ErrWebHookDisabled
	}
	theurl := sdk.JoinURL(
		api.BackendURL(api.EventService, m.channel),
		"/hook",
	)
	client := http.New(m.transport).New(theurl, map[string]string{"Content-Type": "application/json", "Accept": "application/json"})
	data := map[string]interface{}{
		"headers": map[string]string{
			"ref_id":                  refID,
			"integration_instance_id": integrationInstanceID,
			"self_managed":            fmt.Sprintf("%v", m.selfManaged),
			"customer_id":             customerID,
		},
		"customer_id": customerID,
		"system":      refType,
	}
	var res struct {
		Success bool   `json:"success"`
		URL     string `json:"url"`
	}
	opts := make([]sdk.WithHTTPOption, 0)
	if m.channel == "dev" {
		opts = append(opts, sdk.WithHTTPHeader("x-api-key", "fa0s8f09a8sd09f8iasdlkfjalsfm,.m,xf"))
	} else if m.secret != "" {
		opts = append(opts, sdk.WithHTTPHeader("x-api-key", m.secret))
	} else {
		opts = append(opts, sdk.WithAuthorization(m.apikey))
	}
	_, err := client.Post(strings.NewReader(sdk.Stringify(data)), &res, opts...)
	if err != nil {
		return "", fmt.Errorf("error creating webhook url. %w", err)
	}
	if res.Success {
		url := res.URL
		url += "?integration_instance_id=" + integrationInstanceID
		log.Debug(m.logger, "created webhook", "url", url, "customer_id", customerID, "integration_instance_id", integrationInstanceID, "ref_type", refType, "ref_id", refID)
		return url, nil
	}
	return "", fmt.Errorf("failed to create webhook url")
}

// RefreshOAuth2Token will refresh the OAuth2 access token using the provided refreshToken and return a new access token
func (m *eventAPIManager) RefreshOAuth2Token(refType string, refreshToken string) (string, error) {
	if refType == "" {
		return "", fmt.Errorf("error refreshing oauth2 token, missing refType")
	}
	if refreshToken == "" {
		return "", fmt.Errorf("error refreshing oauth2 token, missing refreshToken")
	}
	theurl := sdk.JoinURL(
		api.BackendURL(api.AuthService, m.channel),
		fmt.Sprintf("oauth/%s/refresh/%s", refType, url.PathEscape(refreshToken)),
	)
	var res struct {
		AccessToken string `json:"access_token"`
	}
	client := http.New(m.transport).New(theurl, map[string]string{"Content-Type": "application/json"})
	_, err := client.Get(&res)
	log.Debug(m.logger, "refresh oauth2 token", "url", theurl, "err", err)
	if err != nil {
		return "", err
	}
	if res.AccessToken == "" {
		return "", errors.New("new token not returned, refresh_token might be bad")
	}
	return res.AccessToken, nil
}

// Config is the required fields for a
type Config struct {
	Logger         log.Logger
	Channel        string
	Secret         string
	APIKey         string
	SelfManaged    bool
	WebhookEnabled bool
	RecordDir      string
	ReplayDir      string
}

// New will create a new event api sdk.Manager
func New(cfg Config) (m sdk.Manager, err error) {
	var transport gohttp.RoundTripper
	var r *recorder.Recorder
	name := "agent_" + cfg.Channel + ".yml"
	if cfg.RecordDir != "" {
		recordDir, _ := filepath.Abs(cfg.RecordDir)
		os.RemoveAll(recordDir)
		os.MkdirAll(recordDir, 0700)
		fn := filepath.Join(recordDir, name)
		r, err = recorder.New(fn)
		if err != nil {
			return nil, err
		}
		r.SetTransport(httpdefaults.DefaultTransport())
		transport = r
		log.Info(cfg.Logger, "will record HTTP interactions to "+fn)
	} else if cfg.ReplayDir != "" {
		replayDir, _ := filepath.Abs(cfg.ReplayDir)
		fn := filepath.Join(replayDir, name)
		if !fileutil.FileExists(fn) {
			return nil, fmt.Errorf("missing replay file at %s", fn)
		}
		r, err = recorder.New(fn)
		if err != nil {
			return nil, err
		}
		r.SetTransport(httpdefaults.DefaultTransport())
		transport = r
		log.Info(cfg.Logger, "will replay HTTP interactions from "+fn)
	} else {
		transport = httpdefaults.DefaultTransport()
	}
	return &eventAPIManager{
		logger:         cfg.Logger,
		channel:        cfg.Channel,
		secret:         cfg.Secret,
		apikey:         cfg.APIKey,
		selfManaged:    cfg.SelfManaged,
		webhookEnabled: cfg.WebhookEnabled,
		transport:      transport,
		recorder:       r,
	}, nil
}
