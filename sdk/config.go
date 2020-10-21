package sdk

import (
	"encoding/json"
	"fmt"
	"strings"

	pjson "github.com/pinpt/go-common/v10/json"
	pn "github.com/pinpt/go-common/v10/number"
	ps "github.com/pinpt/go-common/v10/strings"
	gi "github.com/sabhiram/go-gitignore"
)

type auth struct {
	URL     string `json:"url,omitempty"`
	Created int64  `json:"date_ts,omitempty"`
}

type basicAuth struct {
	auth
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}

type oauth1Auth struct {
	auth
	ConsumerKey string `json:"consumer_key"`
	Token       string `json:"oauth_token"`
	Secret      string `json:"oauth_token_secret"`
}

type oauth2Auth struct {
	auth
	AccessToken  string  `json:"access_token"`
	RefreshToken *string `json:"refresh_token"`
	Scopes       *string `json:"scopes"`
}

type apikeyAuth struct {
	auth
	APIKey string `json:"apikey"`
}

type matchListKV map[string]string

// ConfigAccountType the type of account, org or user
type ConfigAccountType string

const (
	//ConfigAccountTypeOrg org account type
	ConfigAccountTypeOrg ConfigAccountType = "ORG"
	// ConfigAccountTypeUser user account type
	ConfigAccountTypeUser ConfigAccountType = "USER"
)

// ConfigAccount single account
type ConfigAccount struct {
	ID     string            `json:"id"`
	Type   ConfigAccountType `json:"type"`
	Public bool              `json:"public"`

	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	AvatarURL   *string `json:"avatarUrl,omitempty"`
	TotalCount  *int64  `json:"totalCount,omitempty"`
	Selected    *bool   `json:"selected,omitempty"`
}

// ConfigAccounts contains accounts with projects or repos to be exported
type ConfigAccounts map[string]*ConfigAccount

type config struct {
	IntegrationType IntegrationType   `json:"integration_type"`
	Exclusions      *matchListKV      `json:"exclusions,omitempty"`
	Inclusions      *matchListKV      `json:"inclusions,omitempty"`
	OAuth1Auth      *oauth1Auth       `json:"oauth1_auth,omitempty"`
	OAuth2Auth      *oauth2Auth       `json:"oauth2_auth,omitempty"`
	BasicAuth       *basicAuth        `json:"basic_auth,omitempty"`
	APIKeyAuth      *apikeyAuth       `json:"apikey_auth,omitempty"`
	Accounts        *ConfigAccounts   `json:"accounts,omitempty"`
	Scope           *IntegrationScope `json:"scope,omitempty"`
}

type matchList struct {
	defaultValue bool
	parsers      map[string]gi.IgnoreParser
}

// Matches returns true if the name matches the list
func (l *matchList) Matches(entity, name string) bool {
	parser := l.parsers[entity]
	if parser == nil {
		return l.defaultValue
	}
	return parser.MatchesPath(name)
}

// IntegrationScope is the integration autoconfig scope type
type IntegrationScope string

const (
	// OrgScope is a org scope
	OrgScope IntegrationScope = "ORG"
	// UserScope is a user scope
	UserScope IntegrationScope = "USER"
)

// IntegrationType is the integration type
type IntegrationType string

const (
	// CloudIntegration is a cloud managed integration
	CloudIntegration IntegrationType = "CLOUD"
	// SelfManagedIntegration is a self-managed integration
	SelfManagedIntegration IntegrationType = "SELFMANAGED"
)

// Config is the integration configuration
type Config struct {
	IntegrationType IntegrationType   `json:"integration_type,omitempty"`
	OAuth1Auth      *oauth1Auth       `json:"oauth1_auth,omitempty"`
	OAuth2Auth      *oauth2Auth       `json:"oauth2_auth,omitempty"`
	BasicAuth       *basicAuth        `json:"basic_auth,omitempty"`
	APIKeyAuth      *apikeyAuth       `json:"apikey_auth,omitempty"`
	Inclusions      *matchList        `json:"inclusions,omitempty"`
	Exclusions      *matchList        `json:"exclusions,omitempty"`
	Accounts        *ConfigAccounts   `json:"accounts,omitempty"`
	Scope           *IntegrationScope `json:"scope,omitempty"`
	Logger          Logger
	kv              map[string]interface{}
}

// Exists will return true if the key exists
func (c Config) Exists(key string) bool {
	_, ok := c.kv[key]
	return ok
}

// Get will return a value if found
func (c Config) Get(key string) (bool, interface{}) {
	val, ok := c.kv[key]
	return ok, val
}

// GetString will return a string coerced value for key
func (c Config) GetString(key string) (bool, string) {
	val, ok := c.kv[key]
	if !ok || val == "" {
		return false, ""
	}
	return ok, ps.Value(val)
}

// GetInt will return a int coerced value for key
func (c Config) GetInt(key string) (bool, int64) {
	val, ok := c.kv[key]
	return ok, pn.ToInt64Any(val)
}

// GetBool will return a bool coerced value for key
func (c Config) GetBool(key string) (bool, bool) {
	val, ok := c.kv[key]
	return ok, pn.ToBoolAny(val)
}

// NewConfig will return a new Config
func NewConfig(kv map[string]interface{}) Config {
	if kv == nil {
		kv = make(map[string]interface{})
	}
	c := Config{kv: kv}
	if exclusions, ok := kv["exclusions"].(string); ok {
		var kv matchListKV
		if err := json.Unmarshal([]byte(exclusions), &kv); err != nil {
			panic(fmt.Errorf("error parsing exclusion: %w", err))
		}
		ml, err := c.parseML(kv)
		if err != nil {
			panic(err)
		}
		c.Exclusions = ml
	}
	if inclusions, ok := kv["inclusions"].(string); ok {
		var kv matchListKV
		if err := json.Unmarshal([]byte(inclusions), &kv); err != nil {
			panic(fmt.Errorf("error parsing inclusion: %w", err))
		}
		ml, err := c.parseML(kv)
		if err != nil {
			panic(err)
		}
		c.Inclusions = ml
	}
	if strval, ok := kv["apikey_auth"].(string); ok {
		var auth apikeyAuth
		if err := json.Unmarshal([]byte(strval), &auth); err != nil {
			panic(fmt.Errorf("error parsing apikey_auth: %w", err))
		}
		c.APIKeyAuth = &auth
	}
	if strval, ok := kv["oauth1_auth"].(string); ok {
		var auth oauth1Auth
		if err := json.Unmarshal([]byte(strval), &auth); err != nil {
			panic(fmt.Errorf("error parsing oauth1_auth: %w", err))
		}
		c.OAuth1Auth = &auth
	}
	if strval, ok := kv["oauth2_auth"].(string); ok {
		var auth oauth2Auth
		if err := json.Unmarshal([]byte(strval), &auth); err != nil {
			panic(fmt.Errorf("error parsing oauth2_auth: %w", err))
		}
		c.OAuth2Auth = &auth
	}
	if strval, ok := kv["basic_auth"].(string); ok {
		var auth basicAuth
		if err := json.Unmarshal([]byte(strval), &auth); err != nil {
			panic(fmt.Errorf("error parsing basic_auth: %w", err))
		}
		c.BasicAuth = &auth
	}
	if strval, ok := kv["accounts"].(string); ok {
		var accounts ConfigAccounts
		if err := json.Unmarshal([]byte(strval), &accounts); err != nil {
			panic(fmt.Errorf("error parsing basic_auth: %w", err))
		}
		c.Accounts = &accounts
	}
	if strval, ok := kv["scope"].(string); ok {
		v := IntegrationScope(strval)
		c.Scope = &v
	}
	return c
}

// Merge in new config
func (c *Config) Merge(kv map[string]interface{}) {
	for k, v := range kv {
		c.kv[k] = v
	}
}

func (c *Config) parseML(val matchListKV) (*matchList, error) {
	ml := &matchList{
		defaultValue: false,
		parsers:      make(map[string]gi.IgnoreParser),
	}
	for entity, ex := range val {
		lines := strings.Split(ex, "\n")
		i, err := gi.CompileIgnoreLines(lines...)
		if err != nil {
			return nil, err
		}
		ml.parsers[entity] = i
	}
	return ml, nil
}

// Parse detail from a buffer into the config
func (c *Config) Parse(buf []byte) error {
	var cfg config
	if err := json.Unmarshal(buf, &c.kv); err != nil {
		return err
	}
	if err := json.Unmarshal(buf, &cfg); err != nil {
		return err
	}
	if cfg.Exclusions != nil {
		ml, err := c.parseML(*cfg.Exclusions)
		if err != nil {
			return err
		}
		c.Exclusions = ml
	}
	if cfg.Inclusions != nil {
		ml, err := c.parseML(*cfg.Inclusions)
		if err != nil {
			return err
		}
		c.Inclusions = ml
	}
	if cfg.APIKeyAuth != nil {
		c.APIKeyAuth = cfg.APIKeyAuth
	}
	if cfg.BasicAuth != nil {
		c.BasicAuth = cfg.BasicAuth
	}
	if cfg.OAuth1Auth != nil {
		c.OAuth1Auth = cfg.OAuth1Auth
	}
	if cfg.OAuth2Auth != nil {
		c.OAuth2Auth = cfg.OAuth2Auth
	}
	if cfg.IntegrationType != "" {
		c.IntegrationType = cfg.IntegrationType
	}
	if cfg.Accounts != nil {
		c.Accounts = cfg.Accounts
	}
	if cfg.Scope != nil {
		c.Scope = cfg.Scope
	}
	return nil
}

// From will serialize a JSON value at key into the interface provided
func (c *Config) From(key string, into interface{}) error {
	if str, ok := c.kv[key].(string); ok {
		return json.Unmarshal([]byte(str), into)
	}
	if str, ok := c.kv[key].(*string); ok {
		return json.Unmarshal([]byte(*str), into)
	}
	if kv, ok := c.kv[key].(map[string]interface{}); ok {
		return json.Unmarshal([]byte(pjson.Stringify(kv)), into)
	}
	found, ok := c.kv[key]
	if ok {
		return json.Unmarshal([]byte(pjson.Stringify(found)), into)
	}
	return fmt.Errorf("%s not found", key)
}
