package sdk

import "time"

// State is a state object to allow the integration to serialize state for a given customer
type State interface {
	// Set a value by key in state. the value must be able to serialize to JSON
	Set(key string, value interface{}) error
	// SetWithExpires will set key and value and it will automatically expire from state after expiry
	SetWithExpires(key string, value interface{}, expiry time.Duration) error
	// Get will return a value for a given key and set the value to the address of out
	Get(key string, out interface{}) (bool, error)
	// Exists return true if the key exists in state
	Exists(key string) bool
	// Delete will return data for key in state
	Delete(key string) error
	// Flush any pending data to storage
	Flush() error
}
