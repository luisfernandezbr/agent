package file

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/pinpt/agent/v4/sdk"
	"github.com/pinpt/go-common/v10/fileutil"
	pjson "github.com/pinpt/go-common/v10/json"
)

type entry struct {
	Value   string
	Expires time.Time
}

// State is a simple file backed state store
type State struct {
	fn    string
	state map[string]*entry
	mu    sync.RWMutex
}

var _ sdk.State = (*State)(nil)
var _ io.Closer = (*State)(nil)

func (f *State) getKey(key string) string {
	return key
}

// Set a value by key in state. the value must be able to serialize to JSON
func (f *State) Set(key string, value interface{}) error {
	f.mu.Lock()
	f.state[f.getKey(key)] = &entry{pjson.Stringify(value), time.Time{}}
	f.mu.Unlock()
	return nil
}

// SetWithExpires will set key and value and it will automatically expire from state after expiry
func (f *State) SetWithExpires(key string, value interface{}, expiry time.Duration) error {
	if expiry <= 0 {
		return fmt.Errorf("invalid expires duration, must be >0, was %d", expiry)
	}
	f.mu.Lock()
	f.state[f.getKey(key)] = &entry{pjson.Stringify(value), time.Now().Add(expiry)}
	f.mu.Unlock()
	return nil
}

// Get will return a value for a given key or nil if not found
func (f *State) Get(key string, out interface{}) (bool, error) {
	statekey := f.getKey(key)
	f.mu.RLock()
	val, found := f.state[statekey]
	f.mu.RUnlock()
	if !found || val == nil || val.Value == "" {
		return false, nil
	}
	if val.Expires.Unix() > 0 && time.Now().After(val.Expires) {
		f.mu.Lock()
		delete(f.state, statekey)
		f.mu.Unlock()
		return false, nil
	}
	err := json.Unmarshal([]byte(val.Value), out)
	return err == nil, err
}

// Exists return true if the key exists in state
func (f *State) Exists(key string) bool {
	statekey := f.getKey(key)
	f.mu.RLock()
	val, exists := f.state[statekey]
	f.mu.RUnlock()
	if exists && val.Expires.Unix() > 0 && time.Now().After(val.Expires) {
		f.mu.Lock()
		delete(f.state, statekey)
		f.mu.Unlock()
		return false
	}
	return exists
}

// Delete will return data for key in state
func (f *State) Delete(key string) error {
	f.mu.Lock()
	delete(f.state, f.getKey(key))
	f.mu.Unlock()
	return nil
}

// Flush any pending data to storage
func (f *State) Flush() error {
	f.mu.Lock()
	err := ioutil.WriteFile(f.fn, []byte(pjson.Stringify(f.state)), 0600)
	f.mu.Unlock()
	return err
}

// Close the state and sync data to the state file
func (f *State) Close() error {
	return f.Flush()
}

// New will create a new state store backed by a file
func New(fn string) (*State, error) {
	kv := make(map[string]*entry)
	var of *os.File
	var err error
	if fileutil.FileExists(fn) {
		of, err = os.Open(fn)
	} else {
		if err := os.MkdirAll(filepath.Dir(fn), 0755); err != nil {
			return nil, err
		}
		of, err = os.Create(fn)
	}
	if err != nil {
		return nil, err
	}
	defer of.Close()
	if err := json.NewDecoder(of).Decode(&kv); err != nil && err != io.EOF {
		return nil, err
	}

	return &State{
		fn:    fn,
		state: kv,
	}, nil
}
