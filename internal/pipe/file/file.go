package file

import (
	"compress/gzip"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/pinpt/agent/sdk"
	"github.com/pinpt/go-common/v10/datamodel"
	"github.com/pinpt/go-common/v10/log"
)

type wrapperFile struct {
	gz *gzip.Writer
	of *os.File
}

func (f *wrapperFile) Write(buf []byte) (int, error) {
	return f.gz.Write(buf)
}

func (f *wrapperFile) Close() error {
	f.gz.Close()
	return f.of.Close()
}

type filePipe struct {
	logger log.Logger
	dir    string
	closed bool
	mu     sync.Mutex
	files  map[string]*wrapperFile
}

var _ sdk.Pipe = (*filePipe)(nil)

var eol = []byte("\n")

// Write a model back to the output system
func (p *filePipe) Write(object datamodel.Model) error {
	if p.closed {
		return fmt.Errorf("pipe closed")
	}
	model := object.GetModelName().String()
	p.mu.Lock()

	// if integration_instance_id, customer_id, or ref_type are missing, error and let the developer know
	if intg, ok := object.(sdk.IntegrationModel); ok {
		if intg.GetIntegrationInstanceID() == nil || *(intg.GetIntegrationInstanceID()) == "" {
			p.mu.Unlock()
			return fmt.Errorf("object missing integration_instance_id: %s", model)
		}
		if intg.GetCustomerID() == "" {
			p.mu.Unlock()
			return fmt.Errorf("object missing customer_id: %s", model)
		}
		if intg.GetRefType() == "" {
			p.mu.Unlock()
			return fmt.Errorf("object missing ref_type: %s", model)
		}
	}

	f := p.files[model]
	if f == nil {
		fp := filepath.Join(p.dir, model+".json.gz")
		of, err := os.OpenFile(fp, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
		if err != nil {
			p.mu.Unlock()
			return err
		}
		gz, err := gzip.NewWriterLevel(of, gzip.BestCompression)
		if err != nil {
			p.mu.Unlock()
			return err
		}
		f = &wrapperFile{gz, of}
		p.files[model] = f
	}
	f.Write([]byte(object.Stringify()))
	f.Write(eol)
	p.mu.Unlock()
	return nil
}

// Flush will tell the pipe to flush any data
func (p *filePipe) Flush() error {
	for _, f := range p.files {
		f.gz.Flush()
	}
	return nil
}

// Close is called when the integration has completed and no more data will be sent
func (p *filePipe) Close() error {
	p.closed = true
	for model, of := range p.files {
		of.Close()
		delete(p.files, model)
	}
	p.files = nil
	return nil
}

// New will create a new console pipe
func New(logger log.Logger, dir string) sdk.Pipe {
	log.Debug(logger, "using file pipe", "dir", dir)
	return &filePipe{
		logger: logger,
		dir:    dir,
		files:  make(map[string]*wrapperFile),
	}
}
