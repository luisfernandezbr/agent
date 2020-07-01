package eventapi

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/pinpt/agent.next/internal/graphql"
	"github.com/pinpt/agent.next/internal/http"
	"github.com/pinpt/agent.next/sdk"
	"github.com/pinpt/go-common/v10/api"
	"github.com/pinpt/go-common/v10/log"
)

type eventAPIManager struct {
	logger      log.Logger
	channel     string
	secret      string
	apikey      string
	selfManaged bool
}

var _ sdk.Manager = (*eventAPIManager)(nil)

// GraphQLManager returns a graphql manager instance
func (m *eventAPIManager) GraphQLManager() sdk.GraphQLClientManager {
	return graphql.New()
}

// HTTPManager returns a HTTP manager instance
func (m *eventAPIManager) HTTPManager() sdk.HTTPClientManager {
	return http.New()
}

// CreateWebHook is used by the integration to create a webhook on behalf of the integration for a given customer and refid
func (m *eventAPIManager) CreateWebHook(customerID string, integrationInstanceID string, refType string, refID string) (string, error) {
	theurl := sdk.JoinURL(
		api.BackendURL(api.EventService, m.channel),
		"/hook",
	)
	client := http.New().New(theurl, map[string]string{"Content-Type": "application/json", "Accept": "application/json"})
	data := map[string]interface{}{
		"headers": map[string]string{
			"ref_id":                  refID,
			"integration_instance_id": integrationInstanceID,
			"self_managed":            fmt.Sprintf("%v", m.selfManaged),
			"customer_id":             customerID,
		},
		"system": refType,
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
		log.Debug(m.logger, "created webhook", "url", res.URL, "customer_id", customerID, "integration_id", integrationInstanceID, "ref_type", refType, "ref_id", refID)
		return res.URL, nil
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
	client := http.New().New(theurl, map[string]string{"Content-Type": "application/json"})
	_, err := client.Get(&res)
	log.Debug(m.logger, "refresh oauth2 token", "url", theurl, "err", err)
	if err != nil {
		return "", err
	}
	return res.AccessToken, nil
}

// Config is the required fields for a
type Config struct {
	Logger      log.Logger
	Channel     string
	Secret      string
	APIKey      string
	SelfManaged bool
}

// New will create a new event api sdk.Manager
func New(cfg Config) sdk.Manager {
	return &eventAPIManager{
		logger:      cfg.Logger,
		channel:     cfg.Channel,
		secret:      cfg.Secret,
		apikey:      cfg.APIKey,
		selfManaged: cfg.SelfManaged,
	}
}
