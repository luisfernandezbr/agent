package file

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"sync"

	"github.com/pinpt/agent.next/sdk"
	"github.com/pinpt/go-common/v10/fileutil"
	pjson "github.com/pinpt/go-common/v10/json"
)

// State is a simple file backed state store
type State struct {
	fn    string
	state map[string]interface{}
	mu    sync.RWMutex
}

var _ sdk.State = (*State)(nil)
var _ io.Closer = (*State)(nil)

func (f *State) getKey(refType string, key string) string {
	return fmt.Sprintf("%s_%s", refType, key)
}

// Set a value by key in state. the value must be able to serialize to JSON
func (f *State) Set(refType string, key string, value interface{}) error {
	f.mu.Lock()
	f.state[f.getKey(refType, key)] = value
	f.mu.Unlock()
	return nil
}

// Get will return a value for a given key or nil if not found
func (f *State) Get(refType string, key string, out interface{}) (bool, error) {
	f.mu.RLock()
	val := f.state[f.getKey(refType, key)]
	f.mu.RUnlock()
	if val == nil {
		return false, nil
	}
	valueof := reflect.ValueOf(out)
	if valueof.Kind() != reflect.Ptr {
		return false, fmt.Errorf("out argument must be a pointer but was %s", valueof.Kind())
	}
	valueof.Elem().Set(reflect.ValueOf(val))
	return true, nil
}

// Exists return true if the key exists in state
func (f *State) Exists(refType string, key string) bool {
	f.mu.RLock()
	_, exists := f.state[f.getKey(refType, key)]
	f.mu.RUnlock()
	return exists
}

// Delete will return data for key in state
func (f *State) Delete(refType string, key string) error {
	f.mu.Lock()
	delete(f.state, f.getKey(refType, key))
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
	kv := make(map[string]interface{})
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
