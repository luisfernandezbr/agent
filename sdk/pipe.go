package sdk

import "github.com/pinpt/go-common/datamodel"

// Pipe for sending data back to pinpoint
type Pipe interface {
	// Write a model back to the output system
	Write(object datamodel.Model) error
	// Flush will tell the pipe to flush any pending data
	Flush() error
	// Close is called when the integration has completed and no more data will be sent
	Close() error
}
