package file

import (
	"io/ioutil"
	"os"
	"testing"

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
	assert.Error(err, "out argument must be a pointer but was string")
}
