package sdk

import (
	"github.com/pinpt/go-common/log"
)

// Integration is the interface that integrations implement
type Integration interface {
	// RefType should return the integration ref_type (the short, unique identifier of the integration)
	RefType() string
	// Start is called when the integration is starting up
	Start(logger log.Logger, config Config, manager Manager) error
	// Export is called to tell the integration to run an export
	Export(export Export) error
	// WebHook is called when a webhook is received on behalf of the integration
	WebHook(webhook WebHook) error
	// Stop is called when the integration is shutting down for cleanup
	Stop() error
}
