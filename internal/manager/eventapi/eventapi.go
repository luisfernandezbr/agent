package eventapi

import (
	"encoding/json"
	"errors"
	"fmt"
	gohttp "net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jhaynie/go-vcr/v2/recorder"
	"github.com/patrickmn/go-cache"
	"github.com/pinpt/agent.next/internal/graphql"
	"github.com/pinpt/agent.next/internal/http"
	"github.com/pinpt/agent.next/sdk"
	"github.com/pinpt/go-common/v10/api"
	"github.com/pinpt/go-common/v10/datetime"
	"github.com/pinpt/go-common/v10/fileutil"
	gql "github.com/pinpt/go-common/v10/graphql"
	"github.com/pinpt/go-common/v10/hash"
	"github.com/pinpt/go-common/v10/httpdefaults"
	"github.com/pinpt/go-common/v10/log"
	"github.com/pinpt/integration-sdk/agent"
	"github.com/pinpt/integration-sdk/sourcecode"
	"github.com/pinpt/integration-sdk/work"
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
	cache          *cache.Cache
}

var _ sdk.Manager = (*eventAPIManager)(nil)
var _ sdk.WebHookManager = (*eventAPIManager)(nil)
var _ sdk.AuthManager = (*eventAPIManager)(nil)
var _ sdk.UserManager = (*eventAPIManager)(nil)

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

// WebHookManager returns the WebHook manager instance
func (m *eventAPIManager) WebHookManager() sdk.WebHookManager {
	return m
}

// UserManager returns the User manager instance
func (m *eventAPIManager) UserManager() sdk.UserManager {
	return m
}

// AuthManager returns the Auth manager instance
func (m *eventAPIManager) AuthManager() sdk.AuthManager {
	return m
}

func (m *eventAPIManager) createGraphql(customerID string) gql.Client {
	url := api.BackendURL(api.GraphService, m.channel)
	client, err := gql.NewClient(customerID, "", m.secret, url)
	if err != nil {
		panic(err)
	}
	if m.apikey != "" {
		client.SetHeader("Authorization", m.apikey)
	}
	return client
}

func (m *eventAPIManager) webhookCacheKey(customerID string, integrationInstanceID string, refType string, refID string, scope sdk.WebHookScope) string {
	return hash.Values(customerID, integrationInstanceID, refType, refID, string(scope))
}

// Create is used by the integration to create a webhook on behalf of the integration for a given customer, reftype and refid
// the result will be a fully qualified URL to the webhook endpoint that should be registered with the integration
func (m *eventAPIManager) Create(customerID string, integrationInstanceID string, refType string, refID string, scope sdk.WebHookScope, params ...string) (string, error) {

	if !m.webhookEnabled {
		return "", nil
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
			"scope":                   string(scope),
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
	dbid := m.webhookCacheKey(customerID, integrationInstanceID, refType, refID, scope)
	if res.Success {
		qs := strings.Join(params, "&")
		if qs != "" {
			qs = "&" + qs
		}
		url := res.URL + "?integration_instance_id=" + integrationInstanceID + "&scope=" + string(scope) + qs
		log.Debug(m.logger, "created webhook", "url", url, "customer_id", customerID, "integration_instance_id", integrationInstanceID, "ref_type", refType, "ref_id", refID, "scope", scope)
		client := m.createGraphql(customerID)
		variables := make(gql.Variables)
		var err error
		switch scope {
		case sdk.WebHookScopeProject:
			projectID := work.NewProjectID(customerID, refID, refType)
			id := work.NewProjectWebhookID(customerID, projectID)
			work.ExecProjectWebhookDeleteMutation(client, id)
			object := work.ProjectWebhook{
				ID:                    id,
				CustomerID:            customerID,
				URL:                   &url,
				IntegrationInstanceID: &integrationInstanceID,
				ProjectID:             projectID,
				RefID:                 refID,
				RefType:               refType,
				Enabled:               true,
			}
			err = work.CreateProjectWebhook(client, object)
		case sdk.WebHookScopeRepo:
			repoID := sourcecode.NewRepoID(customerID, refType, refID)
			id := sourcecode.NewRepoWebhookID(customerID, repoID)
			sourcecode.ExecRepoWebhookDeleteMutation(client, id)
			object := sourcecode.RepoWebhook{
				ID:                    id,
				CustomerID:            customerID,
				URL:                   &url,
				IntegrationInstanceID: &integrationInstanceID,
				RepoID:                repoID,
				RefID:                 refID,
				RefType:               refType,
				Enabled:               true,
			}
			err = sourcecode.CreateRepoWebhook(client, object)
		case sdk.WebHookScopeOrg:
			instance, err := agent.FindIntegrationInstance(client, integrationInstanceID)
			if err != nil {
				return "", fmt.Errorf("error finding the integration instance: %w", err)
			}
			var found bool
			for _, webhook := range instance.Webhooks {
				if (webhook.RefID == nil && refID == "") || (webhook.RefID != nil && *webhook.RefID == refID) {
					webhook.URL = sdk.StringPointer(url)
					webhook.Enabled = true
					webhook.ErrorMessage = nil
					webhook.Errored = false
					found = true
				}
			}
			if !found {
				instance.Webhooks = append(instance.Webhooks, agent.IntegrationInstanceWebhooks{
					URL:     sdk.StringPointer(url),
					RefID:   sdk.StringPointer(refID),
					Enabled: true,
				})
			}
			variables[agent.IntegrationInstanceModelWebhooksColumn] = instance.Webhooks
			err = agent.ExecIntegrationInstanceSilentUpdateMutation(client, integrationInstanceID, variables, false)
		}
		if err != nil {
			m.cache.SetDefault(dbid, url)
		}
		return url, err
	}
	return "", fmt.Errorf("failed to create webhook url: %w", err)
}

// Delete will remove the webhook from the entity based on scope
func (m *eventAPIManager) Delete(customerID string, integrationInstanceID string, refType string, refID string, scope sdk.WebHookScope) error {
	dbid := m.webhookCacheKey(customerID, integrationInstanceID, refType, refID, scope)
	m.cache.Delete(dbid)
	client := m.createGraphql(customerID)
	switch scope {
	case sdk.WebHookScopeProject:
		projectID := work.NewProjectID(customerID, refID, refType)
		return work.ExecProjectWebhookDeleteMutation(client, work.NewProjectWebhookID(customerID, projectID))
	case sdk.WebHookScopeRepo:
		repoID := sourcecode.NewRepoID(customerID, refType, refID)
		return sourcecode.ExecRepoWebhookDeleteMutation(client, sourcecode.NewRepoWebhookID(customerID, repoID))
	case sdk.WebHookScopeOrg:
		instance, err := agent.FindIntegrationInstance(client, integrationInstanceID)
		if err != nil {
			return err
		}
		hooks := make([]agent.IntegrationInstanceWebhooks, 0)
		for _, webhook := range instance.Webhooks {
			if (webhook.RefID == nil && refID == "") || (webhook.RefID != nil && *webhook.RefID == refID) {
				continue
			}
			hooks = append(hooks, webhook)
		}
		variables := make(gql.Variables)
		variables[agent.IntegrationInstanceModelWebhooksColumn] = hooks
		return agent.ExecIntegrationInstanceSilentUpdateMutation(client, integrationInstanceID, variables, false)
	}
	return nil
}

// Exists returns true if the webhook is registered for the given entity based on ref_id and scope
func (m *eventAPIManager) Exists(customerID string, integrationInstanceID string, refType string, refID string, scope sdk.WebHookScope) bool {

	dbid := m.webhookCacheKey(customerID, integrationInstanceID, refType, refID, scope)
	if val, ok := m.cache.Get(dbid); ok && val != nil {
		return ok
	}
	client := m.createGraphql(customerID)
	switch scope {
	case sdk.WebHookScopeProject:
		projectID := work.NewProjectID(customerID, refID, refType)
		webhook, err := work.FindProjectWebhook(client, work.NewProjectWebhookID(customerID, projectID))
		if err == nil && webhook != nil && webhook.RefID == refID {
			m.cache.SetDefault(dbid, *webhook.URL)
			return true
		}
	case sdk.WebHookScopeRepo:
		repoID := sourcecode.NewRepoID(customerID, refType, refID)
		webhook, err := sourcecode.FindRepoWebhook(client, sourcecode.NewRepoWebhookID(customerID, repoID))
		if err == nil && webhook != nil && webhook.RefID == refID {
			m.cache.SetDefault(dbid, *webhook.URL)
			return true
		}
	case sdk.WebHookScopeOrg:
		instance, err := agent.FindIntegrationInstance(client, integrationInstanceID)
		if err == nil && instance != nil && instance.Webhooks != nil {
			for _, webhook := range instance.Webhooks {
				if (webhook.RefID == nil && refID == "") || (webhook.RefID != nil && *webhook.RefID == refID) {
					m.cache.SetDefault(dbid, *webhook.URL)
					return true
				}
			}
		}
	}
	return false
}

// HookURL will return the webhook url
func (m *eventAPIManager) HookURL(customerID string, integrationInstanceID string, refType string, refID string, scope sdk.WebHookScope) (string, error) {
	dbid := m.webhookCacheKey(customerID, integrationInstanceID, refType, refID, scope)
	val, ok := m.cache.Get(dbid)
	if !ok {
		return "", fmt.Errorf("webhook not found")
	}
	return val.(string), nil
}

// Errored will set the errored state on the webhook and the message will be the Error() value of the error
func (m *eventAPIManager) Errored(customerID string, integrationInstanceID string, refType string, refID string, scope sdk.WebHookScope, theerror error) {
	client := m.createGraphql(customerID)
	variables := make(gql.Variables)
	switch scope {
	case sdk.WebHookScopeProject:
		projectID := work.NewProjectID(customerID, refID, refType)
		id := work.NewProjectErrorID(customerID, projectID)
		work.ExecProjectErrorDeleteMutation(client, id)
		object := work.ProjectError{
			ID:                    id,
			CustomerID:            customerID,
			ProjectID:             projectID,
			IntegrationInstanceID: &integrationInstanceID,
			Errored:               true,
			ErrorMessage:          sdk.StringPointer(theerror.Error()),
			RefID:                 refID,
			RefType:               refType,
			UpdatedAt:             datetime.EpochNow(),
		}
		if err := work.CreateProjectError(client, object); err != nil {
			log.Error(m.logger, "error setting the instance project errored", "err", err, "integration_instance_id", integrationInstanceID, "customer_id", customerID, "project_id", projectID)
		}
	case sdk.WebHookScopeRepo:
		repoID := sourcecode.NewRepoID(customerID, refType, refID)
		id := sourcecode.NewRepoErrorID(customerID, repoID)
		sourcecode.ExecRepoErrorDeleteMutation(client, id)
		object := sourcecode.RepoError{
			ID:                    id,
			CustomerID:            customerID,
			RepoID:                repoID,
			IntegrationInstanceID: &integrationInstanceID,
			Errored:               true,
			ErrorMessage:          sdk.StringPointer(theerror.Error()),
			RefID:                 refID,
			RefType:               refType,
			UpdatedAt:             datetime.EpochNow(),
		}
		if err := sourcecode.CreateRepoError(client, object); err != nil {
			log.Error(m.logger, "error setting the instance repo errored", "err", err, "integration_instance_id", integrationInstanceID, "customer_id", customerID, "repo_id", repoID)
		}
	case sdk.WebHookScopeOrg:
		instance, err := agent.FindIntegrationInstance(client, integrationInstanceID)
		if err != nil {
			log.Error(m.logger, "error finding the integration instance", "err", err, "integration_instance_id", integrationInstanceID, "customer_id", customerID)
			return
		}
		for _, webhook := range instance.Webhooks {
			if (webhook.RefID == nil && refID == "") || (webhook.RefID != nil && *webhook.RefID == refID) {
				variables[agent.IntegrationInstanceModelWebhooksErroredColumn] = true
				variables[agent.IntegrationInstanceModelWebhooksErrorMessageColumn] = theerror.Error()
				if err := agent.ExecIntegrationInstanceSilentUpdateMutation(client, integrationInstanceID, variables, false); err != nil {
					log.Error(m.logger, "error setting the integration instance errored", "err", err, "integration_instance_id", integrationInstanceID, "customer_id", customerID)
				}
				return
			}
		}
		// if we don't have a webhook, we need to just update on the instance itself
		variables[agent.IntegrationInstanceModelErrorMessageColumn] = theerror.Error()
		variables[agent.IntegrationInstanceModelErroredColumn] = true
		if err := agent.ExecIntegrationInstanceSilentUpdateMutation(client, integrationInstanceID, variables, false); err != nil {
			log.Error(m.logger, "error setting the integration instance errored", "err", err, "integration_instance_id", integrationInstanceID, "customer_id", customerID)
		}
	}
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

type integrationUserResult struct {
	Custom struct {
		Agent struct {
			Users []sdk.User `json:"integrationUsers"`
		} `json:"agent"`
	} `json:"custom"`
}

// Users will return the integration users for a given integration
func (m *eventAPIManager) Users(control sdk.Control) ([]sdk.User, error) {
	key := control.CustomerID() + ":integration_user:" + control.RefType()
	val, ok := m.cache.Get(key)
	if ok && val != nil {
		users := make([]sdk.User, 0)
		err := json.Unmarshal([]byte(val.(string)), &users)
		if err == nil {
			return users, nil
		}
	}
	client := m.createGraphql(control.CustomerID())
	variables := make(gql.Variables)
	variables["refType"] = control.RefType()
	query := `
query($refType: String!) {
	custom {
		agent {
			integrationUsers(ref_type: $refType) {
				_id
				name
				emails
				ref_id
				oauth1_authorization {
					date_ts
					consumer_key
					oauth_token
					oauth_token_secret
				}
				oauth_authorization {
					date_ts
					token
					refresh_token
					scopes
				}
			}
		}
	}
}`
	var res integrationUserResult
	err := client.Query(query, variables, &res)
	if err == nil && len(res.Custom.Agent.Users) > 0 {
		m.cache.Set(key, sdk.Stringify(res.Custom.Agent.Users), time.Minute*5) // cache for a short period
	}
	return res.Custom.Agent.Users, err
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
		cache:          cache.New(time.Minute*5, time.Minute*6),
	}, nil
}
