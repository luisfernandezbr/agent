package file

import (
	"compress/gzip"
	"context"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/pinpt/agent.next/sdk"
	"github.com/pinpt/go-common/v10/datamodel"
	"github.com/pinpt/go-common/v10/event"
	pjson "github.com/pinpt/go-common/v10/json"
	"github.com/pinpt/go-common/v10/log"
	"github.com/pinpt/integration-sdk/agent"
)

type wrapperFile struct {
	gz    *gzip.Writer
	of    *os.File
	ts    time.Time
	count int
	bytes int64
}

func (f *wrapperFile) Write(buf []byte) (int, error) {
	return f.gz.Write(buf)
}

func (f *wrapperFile) Close() error {
	f.gz.Close()
	return f.of.Close()
}

type eventAPIPipe struct {
	logger     log.Logger
	ctx        context.Context
	cancel     context.CancelFunc
	dir        string
	closed     bool
	mu         sync.Mutex
	files      map[string]*wrapperFile
	customerID string
	uuid       string
	jobid      string
	reftype    string
	channel    string
	apikey     string
	secret     string
	wg         sync.WaitGroup
}

var _ sdk.Pipe = (*eventAPIPipe)(nil)

var eol = []byte("\n")

// Write a model back to the output system
func (p *eventAPIPipe) Write(object datamodel.Model) error {
	if p.closed {
		return fmt.Errorf("pipe closed")
	}
	model := object.GetModelName().String()
	p.mu.Lock()
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
		f = &wrapperFile{gz, of, time.Time{}, 0, 0}
		p.files[model] = f
	}
	f.ts = time.Now() // keep track of last write
	buf := []byte(object.Stringify())
	f.count++
	f.bytes += int64(len(buf) + 1)
	f.Write(buf)
	f.Write(eol)
	p.mu.Unlock()
	return nil
}

// Flush will force any files pending to get sent to the server
func (p *eventAPIPipe) Flush() error {
	// on a flush we're going to send immediately
	p.mu.Lock()
	for model, of := range p.files {
		of.Close()
		if err := p.send(model, of); err != nil {
			log.Error(p.logger, "error sending data to event-api", "model", model, "err", err)
		}
	}
	p.mu.Unlock()
	return nil
}

// Close is called when the integration has completed and no more data will be sent
func (p *eventAPIPipe) Close() error {
	log.Debug(p.logger, "pipe closing")
	p.closed = true
	p.Flush()
	p.cancel()
	p.wg.Wait() // wait for our flush to finish
	p.files = nil
	log.Debug(p.logger, "pipe closed")
	return nil
}

// if any of these limits are exceeded, we will transmit the data to the server
const (
	maxRecords  = 500              // max number of total records in the file before we want to transmit
	maxBytes    = 1024 * 1024 * 5  // max size (~5mb) in bytes in the file before we want to transmit
	maxDuration = 30 * time.Second // amount of time we have an idle file before we transmit
)

type sendRecord struct {
	model string
	file  *wrapperFile
}

func (p *eventAPIPipe) send(model string, f *wrapperFile) error {
	log.Debug(p.logger, "sending to event-api", "model", model, "size", f.bytes, "count", f.count, "last_event", time.Since(f.ts))
	f.Close()
	buf, err := ioutil.ReadFile(f.of.Name())
	if err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}
	object := &agent.ExportData{
		CustomerID: p.customerID,
		RefType:    p.reftype,
		RefID:      p.uuid,
		JobID:      p.jobid,
		Objects:    pjson.Stringify(map[string]string{model: base64.StdEncoding.EncodeToString(buf)}),
	}
	headers := map[string]string{
		"customer_id": p.customerID,
		"uuid":        p.uuid,
	}
	if p.jobid != "" {
		headers["jobid"] = p.jobid
	}
	evt := event.PublishEvent{
		Logger:  p.logger,
		Object:  object,
		Headers: headers,
	}
	opts := make([]event.Option, 0)
	if p.secret != "" {
		opts = append(opts, event.WithHeaders(map[string]string{"x-api-key": p.secret}))
	}
	// publish our data to the event-api
	ts := time.Now()
	if err := event.Publish(p.ctx, evt, p.channel, p.apikey, opts...); err != nil {
		return err
	}
	log.Debug(p.logger, "sent to event-api", "model", model, "duration", time.Since(ts))
	return nil
}

// run will create a background goroutine for checking to see if we have any data that is ready
// to stream to the cloud and if so, will go ahead and send it on another background goroutine
func (p *eventAPIPipe) run() {
	ticker := time.NewTicker(5 * time.Second)
	ch := make(chan sendRecord, 3)
	// create a background sender
	var lock sync.Mutex
	p.wg.Add(1)
	go func() {
		defer p.wg.Done()
		for record := range ch {
			lock.Lock()
			if err := p.send(record.model, record.file); err != nil {
				log.Error(p.logger, "error sending data to event-api", "model", record.model, "err", err)
			}
			lock.Unlock()
		}
	}()
	for {
		select {
		case <-ticker.C:
			// cycle through and see if we have any models that are ready to transmit
			p.mu.Lock()
			for model, f := range p.files {
				if f.count >= maxRecords || f.bytes >= maxBytes || time.Since(f.ts) >= maxDuration {
					f.Close()
					// delete, it gets created on demand
					delete(p.files, model)
					// ready to send
					ch <- sendRecord{model, f}
				}
			}
			p.mu.Unlock()
		case <-p.ctx.Done():
			log.Debug(p.logger, "run ctx is done")
			ticker.Stop()
			// hold lock so we can safely close channel
			lock.Lock()
			close(ch)
			lock.Unlock()
			return
		}
	}
}

// Config is the configuration
type Config struct {
	Ctx        context.Context
	Logger     log.Logger
	Dir        string
	CustomerID string
	UUID       string
	JobID      string
	RefType    string
	Channel    string
	APIKey     string
	Secret     string
}

// New will create a new eventapi pipe
func New(config Config) sdk.Pipe {
	c := config.Ctx
	if c == nil {
		c = context.Background()
	}
	ctx, cancel := context.WithCancel(c)
	p := &eventAPIPipe{
		logger:     config.Logger,
		dir:        config.Dir,
		files:      make(map[string]*wrapperFile),
		channel:    config.Channel,
		customerID: config.CustomerID,
		uuid:       config.UUID,
		jobid:      config.JobID,
		reftype:    config.RefType,
		apikey:     config.APIKey,
		secret:     config.Secret,
		ctx:        ctx,
		cancel:     cancel,
	}
	go p.run()
	return p
}
