package file

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/pinpt/agent.next/sdk"
	"github.com/pinpt/go-common/datamodel"
	"github.com/pinpt/go-common/log"
)

type filePipe struct {
	logger log.Logger
	dir    string
	closed bool
	mu     sync.Mutex
	files  map[string]*os.File
}

var _ sdk.Pipe = (*filePipe)(nil)

// Write a model back to the output system
func (p *filePipe) Write(object datamodel.Model) error {
	if p.closed {
		return fmt.Errorf("pipe closed")
	}
	model := object.GetModelName().String()
	p.mu.Lock()
	f := p.files[model]
	if f == nil {
		fp := filepath.Join(p.dir, model+".json")
		of, err := os.OpenFile(fp, os.O_APPEND|os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
		if err != nil {
			p.mu.Unlock()
			return err
		}
		f = of
		p.files[model] = f
	}
	p.mu.Unlock()
	f.WriteString(object.Stringify())
	f.WriteString("\n")
	return nil
}

// Close is called when the integration has completed and no more data will be sent
func (p *filePipe) Close() error {
	p.closed = true
	for model, of := range p.files {
		of.Close()
		delete(p.files, model)
	}
	return nil
}

// New will create a new console pipe
func New(logger log.Logger, dir string) sdk.Pipe {
	return &filePipe{
		logger: logger,
		dir:    dir,
		files:  make(map[string]*os.File),
	}
}
