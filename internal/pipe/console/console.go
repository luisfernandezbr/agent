package console

import (
	"fmt"

	"github.com/pinpt/agent.next/sdk"
	"github.com/pinpt/go-common/v10/datamodel"
	"github.com/pinpt/go-common/v10/log"
)

type consolePipe struct {
	logger log.Logger
	closed bool
}

var _ sdk.Pipe = (*consolePipe)(nil)

// Write a model back to the output system
func (p *consolePipe) Write(object datamodel.Model) error {
	if p.closed {
		return fmt.Errorf("pipe closed")
	}
	log.Debug(p.logger, object.Stringify(), "model", object.GetModelName())
	return nil
}

func (p *consolePipe) Flush() error {
	return nil
}

// Close is called when the integration has completed and no more data will be sent
func (p *consolePipe) Close() error {
	p.closed = true
	return nil
}

// New will create a new console pipe
func New(logger log.Logger) sdk.Pipe {
	return &consolePipe{logger, false}
}
