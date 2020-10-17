package console

import (
	"fmt"

	"github.com/pinpt/agent/v4/sdk"
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
	model := object.GetModelName()
	if intg, ok := object.(sdk.IntegrationModel); ok {
		if intg.GetIntegrationInstanceID() == nil || *(intg.GetIntegrationInstanceID()) == "" {
			return fmt.Errorf("object missing integration_instance_id: %s", model)
		}
		if intg.GetCustomerID() == "" {
			return fmt.Errorf("object missing customer_id: %s", model)
		}
		if intg.GetRefType() == "" {
			return fmt.Errorf("object missing ref_type: %s", model)
		}
	}
	log.Debug(p.logger, object.Stringify(), "model", model)
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
	log.Debug(logger, "using log pipe")
	return &consolePipe{logger, false}
}
