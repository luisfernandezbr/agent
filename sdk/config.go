package sdk

import (
	"encoding/json"
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
}

type apikeyAuth struct {
	auth
	APIKey string `json:"apikey"`
}

type config struct {
	Exclusions *string     `json:"exclusions,omitempty"`
	Inclusions *string     `json:"inclusions,omitempty"`
	OAuth2Auth *oauth2Auth `json:"oauth2_auth,omitempty"`
	BasicAuth  *basicAuth  `json:"basic_auth,omitempty"`
	APIKeyAuth *apikeyAuth `json:"apikey_auth,omitempty"`
}

type matchList struct {
	parser gi.IgnoreParser
}

// Matches returns true if the name matches the exclusion list
func (l *matchList) Matches(name string) bool {
	return l.parser.MatchesPath(name)
}

// Config is the integration configuration
type Config struct {
	OAuth2Auth *oauth2Auth `json:"oauth2_auth,omitempty"`
	BasicAuth  *basicAuth  `json:"basic_auth,omitempty"`
	APIKeyAuth *apikeyAuth `json:"apikey_auth,omitempty"`
	Inclusions *matchList  `json:"-"`
	Exclusions *matchList  `json:"-"`

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
	return Config{kv: kv}
}

// Merge in new config
func (c *Config) Merge(kv map[string]interface{}) {
	for k, v := range kv {
		c.kv[k] = v
	}
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
		lines := strings.Split(*cfg.Exclusions, "\n")
		i, err := gi.CompileIgnoreLines(lines...)
		if err != nil {
			return err
		}
		c.Exclusions = &matchList{i}
	}
	if cfg.Inclusions != nil {
		lines := strings.Split(*cfg.Inclusions, "\n")
		i, err := gi.CompileIgnoreLines(lines...)
		if err != nil {
			return err
		}
		c.Inclusions = &matchList{i}
	}
	c.APIKeyAuth = cfg.APIKeyAuth
	c.BasicAuth = cfg.BasicAuth
	c.OAuth2Auth = cfg.OAuth2Auth
	return nil
}
