package sdk

import (
	"encoding/json"
	"fmt"
	"strings"

	pn "github.com/pinpt/go-common/v10/number"
	ps "github.com/pinpt/go-common/v10/strings"
	gi "github.com/sabhiram/go-gitignore"
)

type auth struct {
	URL string `json:"url,omitempty"`
}

type basicAuth struct {
	auth
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
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

type config struct {
	IntegrationType IntegrationType `json:"integrationType"`
	Exclusions      *matchListKV    `json:"exclusions,omitempty"`
	Inclusions      *matchListKV    `json:"inclusions,omitempty"`
	OAuth2Auth      *oauth2Auth     `json:"oauth2_auth,omitempty"`
	BasicAuth       *basicAuth      `json:"basic_auth,omitempty"`
	APIKeyAuth      *apikeyAuth     `json:"apikey_auth,omitempty"`
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
	IntegrationType IntegrationType `json:"integrationType"`
	OAuth2Auth      *oauth2Auth     `json:"oauth2_auth,omitempty"`
	BasicAuth       *basicAuth      `json:"basic_auth,omitempty"`
	APIKeyAuth      *apikeyAuth     `json:"apikey_auth,omitempty"`
	Inclusions      *matchList      `json:"-"`
	Exclusions      *matchList      `json:"-"`

	kv map[string]interface{}
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
	if cfg.OAuth2Auth != nil {
		c.OAuth2Auth = cfg.OAuth2Auth
	}
	if cfg.IntegrationType != "" {
		c.IntegrationType = cfg.IntegrationType
	}
	return nil
}
