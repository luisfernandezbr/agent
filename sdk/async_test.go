package sdk

import (
	"errors"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAsync(t *testing.T) {
	assert := assert.New(t)
	words := "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Fusce vestibulum metus id interdum dapibus. Phasellus imperdiet ac tellus et porttitor"
	slice := strings.Split(words, " ")
	a := NewAsync(10)
	var newslice []string
	var mu sync.Mutex
	for _, _word := range slice {
		// IMPORTANT!
		// copy the values that need to be used inside the function to new vars
		word := _word
		a.Do(func() error {
			mu.Lock()
			newslice = append(newslice, word)
			mu.Unlock()
			return nil
		})
	}
	err := a.Wait()
	assert.NoError(err)
	// the order of the words are different, but they contain the same amount of words
	assert.Equal(len(slice), len(newslice))
	assert.NotEqual(slice, newslice)
}

func TestAsyncError(t *testing.T) {

	assert := assert.New(t)
	words := "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Fusce vestibulum metus id interdum dapibus. Phasellus imperdiet ac tellus et porttitor"
	slice := strings.Split(words, " ")
	a := NewAsync(10)
	var newslice []string
	var mu sync.Mutex
	var index int
	for _, _word := range slice {
		word := _word
		a.Do(func() error {
			mu.Lock()
			index++
			defer mu.Unlock()
			if index == 5 {
				return errors.New("dummy error")
			}
			newslice = append(newslice, word)
			return nil
		})
	}
	err := a.Wait()
	assert.Error(err)
	assert.EqualError(err, "dummy error")
	assert.NotEqual(len(slice), len(newslice))
	assert.NotEqual(slice, newslice)

}
