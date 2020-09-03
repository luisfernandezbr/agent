package sdk

import (
	"encoding/json"
	"sync"
)

// Stats is a write-only, concurrency safe map
type Stats interface {
	json.Marshaler
	Set(key string, val interface{})
	Increment(key string, n int64)
	String() (string, error)
}

// stats is write-only, concurrency safe map
type stats struct {
	kv map[string]interface{}
	mu sync.Mutex
}

// NewStats will return a properly initialized Stats
func NewStats() Stats {
	return &stats{
		kv: make(map[string]interface{}),
	}
}

// Set writes to the map
func (s *stats) Set(key string, val interface{}) {
	s.mu.Lock()
	s.kv[key] = val
	s.mu.Unlock()
}

// Set writes to the map
func (s *stats) Increment(key string, n int64) {
	s.mu.Lock()
	v, ok := s.kv[key]
	var nval int64
	if !ok {
		nval = n
	} else {
		if val, ok := v.(int); ok {
			nval = int64(val) + n
		} else if val, ok := v.(int64); ok {
			nval = val + n
		} else if val, ok := v.(int32); ok {
			nval = int64(val) + n
		} else {
			// TODO(robin): maybe throw an error?
		}
	}
	s.kv[key] = nval
	s.mu.Unlock()
}

// MarshalJSON returns the underlying map
func (s *stats) MarshalJSON() (buf []byte, err error) {
	s.mu.Lock()
	buf, err = json.Marshal(s.kv)
	s.mu.Unlock()
	return
}

func (s *stats) String() (string, error) {
	buf, err := s.MarshalJSON()
	return string(buf), err
}

// PrefixStats will return a stats that writes to s with keys prefixed with prefix
func PrefixStats(s Stats, prefix string) Stats {
	return &prefixedStats{
		s:      s,
		prefix: prefix,
	}
}

type prefixedStats struct {
	s      Stats
	prefix string
}

func (p *prefixedStats) withPrefix(key string) string {
	return p.prefix + "." + key
}

func (p *prefixedStats) Set(key string, val interface{}) {
	p.s.Set(p.withPrefix(key), val)
}

func (p *prefixedStats) Increment(key string, n int64) {
	p.s.Increment(p.withPrefix(key), n)
}

func (p *prefixedStats) MarshalJSON() ([]byte, error) {
	return p.s.MarshalJSON()
}

func (p *prefixedStats) String() (string, error) {
	return p.s.String()
}
