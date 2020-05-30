package file

import (
	"compress/gzip"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/pinpt/agent.next/sdk"
	"github.com/pinpt/go-common/datamodel"
	"github.com/pinpt/go-common/event"
	"github.com/pinpt/go-common/log"
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
	p.closed = true
	p.Flush()
	p.cancel()
	p.wg.Wait() // wait for our flush to finish
	p.files = nil
	return nil
}

// if any of these limits are exceeded, we will transmit the data to the server
const (
	maxRecords  = 250              // max number of total records in the file before we want to transmit
	maxBytes    = 1024 * 1024 * 3  // max size (~3mb) in bytes in the file before we want to transmit
	maxDuration = 30 * time.Second // amount of time we have an idle file before we transmit
)

type sendRecord struct {
	model string
	file  *wrapperFile
}

func (p *eventAPIPipe) send(model string, f *wrapperFile) error {
	log.Debug(p.logger, "sending to event-api", "model", model, "size", f.bytes, "count", f.count)
	object := &agent.ExportResponse{} // FIXME - need a new batch type object
	evt := event.PublishEvent{
		Logger: p.logger,
		Object: object,
		Headers: map[string]string{
			"customer_id": p.customerID,
			"uuid":        p.uuid,
			"jobid":       p.jobid,
		},
	}
	var opts event.Option
	if p.apikey != "" {
		opts = event.WithHeaders(map[string]string{"x-api-key": p.apikey})
	}
	// publish our data to the event-api
	ts := time.Now()
	if err := event.Publish(p.ctx, evt, p.channel, p.apikey, opts); err != nil {
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
	go func() {
		defer p.wg.Done()
		p.wg.Add(1)
		for record := range ch {
			if err := p.send(record.model, record.file); err != nil {
				log.Error(p.logger, "error sending data to event-api", "model", record.model, "err", err)
			}
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
			ticker.Stop()
			return
		}
	}
}

// New will create a new eventapi pipe
func New(logger log.Logger, dir string, customerID string, uuid string, jobid string, channel string, apikey string, secret string) sdk.Pipe {
	ctx, cancel := context.WithCancel(context.Background())
	p := &eventAPIPipe{
		logger:     logger,
		dir:        dir,
		files:      make(map[string]*wrapperFile),
		channel:    channel,
		customerID: customerID,
		uuid:       uuid,
		jobid:      jobid,
		apikey:     apikey,
		secret:     secret,
		ctx:        ctx,
		cancel:     cancel,
	}
	go p.run()
	return p
}
