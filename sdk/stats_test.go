package sdk

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStats(t *testing.T) {
	assert := assert.New(t)
	s := NewStats()
	s.Set("prs", 15)
	val, err := s.String()
	assert.NoError(err)
	assert.Equal("{\"prs\":15}", val)
}

func TestStatsIncrement(t *testing.T) {
	assert := assert.New(t)
	s := NewStats()
	s.Set("prs", 15)
	val, err := s.String()
	assert.NoError(err)
	assert.Equal("{\"prs\":15}", val)
	s.Increment("prs", 5)
	s.Increment("new", 1)
	val, err = s.String()
	assert.NoError(err)
	assert.Equal("{\"new\":1,\"prs\":20}", val)
}

func TestStatsConcurrency(t *testing.T) {
	assert := assert.New(t)
	f := func() {
		s := NewStats()
		go func() {
			for i := 0; i < 30; i++ {
				s.MarshalJSON()
			}
		}()
		go func() {
			for i := 0; i < 30; i++ {
				s.Set("prs", 15)
			}
		}()
	}
	assert.NotPanics(f)
}

func TestStatsWithPrefix(t *testing.T) {
	assert := assert.New(t)

	s := NewStats()
	testingStats := PrefixStats(s, "testing")
	fooStats := PrefixStats(s, "foo")

	s.Set("a", 15)
	testingStats.Set("b", 15)
	fooStats.Set("c", 15)

	expected := "{\"a\":15,\"foo.c\":15,\"testing.b\":15}"

	val, err := s.String()
	assert.NoError(err)
	assert.Equal(expected, val)
	val, err = testingStats.String()
	assert.NoError(err)
	assert.Equal(expected, val)
	val, err = fooStats.String()
	assert.NoError(err)
	assert.Equal(expected, val)
}

func TestStatsWithNestedPrefix(t *testing.T) {
	assert := assert.New(t)

	s := NewStats()
	testingStats := PrefixStats(s, "testing")
	doubleTestingStats := PrefixStats(testingStats, "testing")

	s.Set("a", 15)
	doubleTestingStats.Set("a", 15)

	expected := "{\"a\":15,\"testing.testing.a\":15}"

	val, err := s.String()
	assert.NoError(err)
	assert.Equal(expected, val)
}

func TestStatsStringify(t *testing.T) {
	assert := assert.New(t)

	s := NewStats()
	s.Set("a", 15)

	assert.Equal("{\"a\":15}", Stringify(s))
}
