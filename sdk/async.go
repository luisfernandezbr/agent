package sdk

import "sync"

// Async simple async interface
type Async interface {
	Do(f func() error)
	Wait() error
}

// NewAsync instantiates a new Async object
func NewAsync(concurrency int) Async {
	a := &async{}
	a.funcs = make(chan func() error, concurrency)
	a.wg.Add(concurrency)
	for i := 0; i < concurrency; i++ {
		go func() {
			defer a.wg.Done()
			for f := range a.funcs {
				if err := f(); err != nil {
					a.mu.Lock()
					a.err = err
					a.mu.Unlock()
					return // stopping processing if we get an error
				}
			}
		}()
	}
	return a
}

type async struct {
	funcs chan func() error
	err   error
	wg    sync.WaitGroup
	mu    sync.Mutex
}

func (a *async) Do(f func() error) {
	a.funcs <- f
}

func (a *async) Wait() error {
	close(a.funcs)
	a.wg.Wait()
	return a.err
}
