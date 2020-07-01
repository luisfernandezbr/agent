package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/pinpt/agent.next/sdk"
	pjson "github.com/pinpt/go-common/v10/json"
)

// State is a simple file backed state store
type State struct {
	ctx    context.Context
	client *redis.Client
	prefix string
}

var _ sdk.State = (*State)(nil)
var _ io.Closer = (*State)(nil)

func (f *State) getKey(key string) string {
	return fmt.Sprintf("agent:%s:%s", f.prefix, key)
}

// Set a value by key in state. the value must be able to serialize to JSON
func (f *State) Set(key string, value interface{}) error {
	return f.client.Set(f.ctx, f.getKey(key), pjson.Stringify(value), 0).Err()
}

// SetWithExpires will set key and value and it will automatically expire from state after expiry
func (f *State) SetWithExpires(key string, value interface{}, expiry time.Duration) error {
	if expiry <= 0 {
		return fmt.Errorf("invalid expires duration, must be >0, was %d", expiry)
	}
	return f.client.Set(f.ctx, f.getKey(key), pjson.Stringify(value), expiry).Err()
}

// Get will return a value for a given key or nil if not found
func (f *State) Get(key string, val interface{}) (bool, error) {
	str, err := f.client.Get(f.ctx, f.getKey(key)).Result()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	err = json.Unmarshal([]byte(str), val)
	return true, err
}

// Exists return true if the key exists in state
func (f *State) Exists(key string) bool {
	val := f.client.Exists(f.ctx, f.getKey(key)).Val()
	return val > 0
}

// Delete will return data for key in state
func (f *State) Delete(key string) error {
	return f.client.Del(f.ctx, f.getKey(key)).Err()
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
func New(ctx context.Context, client *redis.Client, prefix string) (*State, error) {
	return &State{
		ctx:    ctx,
		client: client,
		prefix: prefix,
	}, nil
}
