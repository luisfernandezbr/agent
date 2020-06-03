package sdk

import (
	pn "github.com/pinpt/go-common/v10/number"
	ps "github.com/pinpt/go-common/v10/strings"
)

// Config is the integration configuration
type Config struct {
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
	return Config{kv}
}
