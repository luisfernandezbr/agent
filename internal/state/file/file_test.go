package file

import (
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFile(t *testing.T) {
	assert := assert.New(t)
	tmpfn, _ := ioutil.TempFile("", "")
	defer os.Remove(tmpfn.Name())
	state, err := New(tmpfn.Name())
	assert.NoError(err)
	var val string
	ok, err := state.Get("b", &val)
	assert.NoError(err)
	assert.False(ok)
	assert.NoError(state.Set("b", "c"))
	ok, err = state.Get("b", &val)
	assert.NoError(err)
	assert.True(ok)
	assert.Equal("c", val)
	ok, err = state.Get("b", val)
	assert.False(ok)
	assert.EqualError(err, "json: Unmarshal(non-pointer string)")
	err = state.SetWithExpires("test", "foo", time.Microsecond)
	assert.NoError(err)
	time.Sleep(2 * time.Microsecond)
	assert.False(state.Exists("test"))
}
