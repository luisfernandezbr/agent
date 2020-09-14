package eventapi

import (
	"compress/gzip"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestWrapperFile(t *testing.T) {
	assert := assert.New(t)
	of, err := ioutil.TempFile("", "pipe")
	assert.NoError(err)
	defer os.Remove(of.Name())
	gz, err := gzip.NewWriterLevel(of, gzip.BestCompression)
	assert.NoError(err)
	f := &wrapperFile{gz, of, time.Time{}, 0, 0}
	wrote, err := f.Write([]byte("hello"))
	assert.NoError(err)
	assert.Equal(5, wrote)
	wrote, err = f.WriteLine([]byte("hello"))
	assert.NoError(err)
	assert.Equal(6, wrote)
	assert.NoError(f.Close())

	of2, err := os.Open(f.of.Name())
	assert.NoError(err)
	defer of2.Close()
	gz2, err := gzip.NewReader(of2)
	assert.NoError(err)
	buf, err := ioutil.ReadAll(gz2)
	assert.NoError(err)
	assert.EqualValues("hellohello\n", string(buf))
}
