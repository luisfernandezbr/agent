package sdk

// State is a state object to allow the integration to serialize state for a given customer
type State interface {
	// Set a value by key in state. the value must be able to serialize to JSON
	Set(refType string, key string, value interface{}) error
	// Get will return a value for a given key and set the value to the address of out
	Get(refType string, key string, out interface{}) (bool, error)
	// Exists return true if the key exists in state
	Exists(refType string, key string) bool
	// Delete will return data for key in state
	Delete(refType string, key string) error
	// Flush any pending data to storage
	Flush() error
}
