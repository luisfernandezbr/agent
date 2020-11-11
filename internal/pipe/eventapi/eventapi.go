package eventapi

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

	"github.com/pinpt/agent/v4/sdk"
	"github.com/pinpt/go-common/v10/datamodel"
	"github.com/pinpt/go-common/v10/datetime"
	"github.com/pinpt/go-common/v10/event"
	pjson "github.com/pinpt/go-common/v10/json"
	"github.com/pinpt/go-common/v10/log"
	pnum "github.com/pinpt/go-common/v10/number"
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

func (f *wrapperFile) WriteLine(buf []byte) (int, error) {
	return f.Write(append(buf, eol...))
}

func (f *wrapperFile) Close() error {
	if err := f.gz.Close(); err != nil {
		return err
	}
	return f.of.Close()
}

func newWrapperFile(dir, model string) (*wrapperFile, error) {
	fp := filepath.Join(dir, fmt.Sprintf("%s-%d.json.gz", model, datetime.EpochNow()))
	of, err := os.OpenFile(fp, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	gz, err := gzip.NewWriterLevel(of, gzip.BestCompression)
	if err != nil {
		return nil, err
	}
	return &wrapperFile{gz, of, time.Time{}, 0, 0}, nil
}

type eventAPIPipe struct {
	logger                log.Logger
	ctx                   context.Context
	cancel                context.CancelFunc
	dir                   string
	closed                bool
	fastlane              bool
	mu                    sync.Mutex
	files                 map[string]*wrapperFile
	customerID            string
	uuid                  string
	jobid                 string
	reftype               string
	channel               string
	apikey                string
	secret                string
	integrationInstanceID string
	wg                    sync.WaitGroup
	started               time.Time
	stats                 sdk.Stats
}

var _ sdk.Pipe = (*eventAPIPipe)(nil)

var eol = []byte("\n")

// Write a model back to the output system
func (p *eventAPIPipe) Write(object datamodel.Model) error {
	if p.closed {
		return fmt.Errorf("pipe closed")
	}
	model := object.GetModelName().String()
	if object == nil {
		return fmt.Errorf("wrote nil model to pipe: %s", model)
	}
	if p.stats != nil {
		p.stats.Increment(model, 1)
	}
	// if integration_instance_id, customer_id, or ref_type are missing, try to set them
	if intg, ok := object.(sdk.IntegrationModel); ok {
		if intg.GetIntegrationInstanceID() == nil || *(intg.GetIntegrationInstanceID()) == "" {
			intg.SetIntegrationInstanceID(p.integrationInstanceID)
		}
		if intg.GetCustomerID() == "" {
			if p.customerID == "" {
				return fmt.Errorf("object missing customer_id: %s", model)
			}
			intg.SetCustomerID(p.customerID)
		}
		if intg.GetRefType() == "" {
			if p.reftype == "" {
				return fmt.Errorf("object missing ref_type: %s", model)
			}
			intg.SetRefType(p.reftype)
		}
	}
	p.mu.Lock()
	f := p.files[model]
	if f == nil {
		var err error
		f, err = newWrapperFile(p.dir, model)
		if err != nil {
			return fmt.Errorf("error creating new wrapper file: %w", err)
		}
		p.files[model] = f
	}
	f.ts = time.Now() // keep track of last write
	buf := []byte(object.Stringify())
	f.count++
	f.bytes += int64(len(buf) + 1)
	if _, err := f.WriteLine(buf); err != nil {
		p.mu.Unlock()
		return fmt.Errorf("error writing to buffer file: %w", err)
	}
	p.mu.Unlock()
	return nil
}

// Flush will force any files pending to get sent to the server
func (p *eventAPIPipe) Flush() error {
	// on a flush we're going to send immediately
	p.mu.Lock()
	for model, of := range p.files {
		of.Close()
		delete(p.files, model)
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
	if err := p.Flush(); err != nil {
		log.Error(p.logger, "error flushing pipe", "err", err)
	}
	p.cancel()
	p.wg.Wait() // wait for our flush to finish
	p.files = nil
	log.Debug(p.logger, "pipe closed", "duration", time.Since(p.started))
	return nil
}

// if any of these limits are exceeded, we will transmit the data to the server
const (
	maxRecords  = 500              // max number of total records in the file before we want to transmit
	maxBytes    = 1024 * 1024 * 1  // max size (~1mb) in bytes in the file before we want to transmit
	maxDuration = 30 * time.Second // amount of time we have an idle file before we transmit
)

type sendRecord struct {
	model string
	file  *wrapperFile
}

func (p *eventAPIPipe) send(model string, f *wrapperFile) error {
	log.Debug(p.logger, "sending to event-api", "model", model, "size", pnum.ToBytesSize(f.bytes), "count", f.count, "last_event", time.Since(f.ts))
	buf, err := ioutil.ReadFile(f.of.Name())
	if err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}
	object := &agent.ExportData{
		CustomerID:            p.customerID,
		RefType:               p.reftype,
		RefID:                 p.uuid,
		JobID:                 p.jobid,
		IntegrationInstanceID: p.integrationInstanceID,
		Objects:               pjson.Stringify(map[string]string{model: base64.StdEncoding.EncodeToString(buf)}),
	}
	headers := map[string]string{
		"customer_id":             p.customerID,
		"uuid":                    p.uuid,
		"integration_instance_id": p.integrationInstanceID,
	}
	if p.jobid != "" {
		headers["jobid"] = p.jobid
	}
	if p.fastlane {
		headers["fastlane"] = "true"
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
	os.Remove(f.of.Name())
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
					if err := f.Close(); err != nil {
						log.Error(p.logger, "error closing pipe file", "err", err)
					}
					// delete, it gets created on demand
					delete(p.files, model)
					// ready to send
					ch <- sendRecord{model, f}
				}
			}
			p.mu.Unlock()
		case <-p.ctx.Done():
			log.Debug(p.logger, "pipe run ctx is done")
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
	Ctx                   context.Context
	Logger                log.Logger
	Dir                   string
	CustomerID            string
	UUID                  string
	JobID                 string
	IntegrationInstanceID string
	RefType               string
	Channel               string
	APIKey                string
	Secret                string
	Stats                 sdk.Stats
	Fastlane              bool
}

// New will create a new eventapi pipe
func New(config Config) sdk.Pipe {
	c := config.Ctx
	if c == nil {
		c = context.Background()
	}
	ctx, cancel := context.WithCancel(c)
	p := &eventAPIPipe{
		logger:                config.Logger,
		dir:                   config.Dir,
		files:                 make(map[string]*wrapperFile),
		channel:               config.Channel,
		customerID:            config.CustomerID,
		uuid:                  config.UUID,
		jobid:                 config.JobID,
		reftype:               config.RefType,
		apikey:                config.APIKey,
		secret:                config.Secret,
		integrationInstanceID: config.IntegrationInstanceID,
		ctx:                   ctx,
		cancel:                cancel,
		started:               time.Now(),
		stats:                 config.Stats,
		fastlane:              config.Fastlane,
	}
	go p.run()
	return p
}
