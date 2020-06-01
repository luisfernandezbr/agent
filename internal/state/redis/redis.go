package redis

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/go-redis/redis"
	"github.com/pinpt/agent.next/sdk"
	pjson "github.com/pinpt/go-common/json"
)

// State is a simple file backed state store
type State struct {
	client     *redis.Client
	customerID string
}

var _ sdk.State = (*State)(nil)
var _ io.Closer = (*State)(nil)

func (f *State) getKey(refType string, key string) string {
	return fmt.Sprintf("%s_%s_%s", f.customerID, refType, key)
}

// Set a value by key in state. the value must be able to serialize to JSON
func (f *State) Set(refType string, key string, value interface{}) error {
	return f.client.Set(f.getKey(refType, key), pjson.Stringify(value), 0).Err()
}

// Get will return a value for a given key or nil if not found
func (f *State) Get(refType string, key string) (interface{}, error) {
	res := f.client.Get(f.getKey(refType, key))
	err := res.Err()
	if err == redis.Nil {
		return nil, err
	}
	if err != nil {
		return nil, err
	}
	str := res.Val()
	var val interface{}
	err = json.Unmarshal([]byte(str), val)
	return val, err
}

// Exists return true if the key exists in state
func (f *State) Exists(refType string, key string) bool {
	val := f.client.Exists(f.getKey(refType, key)).Val()
	return val > 0
}

// Delete will return data for key in state
func (f *State) Delete(refType string, key string) error {
	return f.client.Del(f.getKey(refType, key)).Err()
}

// Flush any pending data to storage
func (f *State) Flush() error {
	return nil
}

// Close the state and sync data to the state file
func (f *State) Close() error {
	return nil
}

// New will create a new state store backed by Redis
func New(client *redis.Client, customerID string) (*State, error) {
	return &State{
		client:     client,
		customerID: customerID,
	}, nil
}
