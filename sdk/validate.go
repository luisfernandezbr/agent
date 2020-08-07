package sdk

// Validate is a control interface for validating a configuration before enroll
type Validate interface {
	Identifier
	Config() Config
}
