package sdk

// MutationAction is a mutation action type
type MutationAction string

const (
	// CreateAction is a create mutation action
	CreateAction MutationAction = "create"
	// UpdateAction is a update mutation action
	UpdateAction MutationAction = "update"
	// DeleteAction is a delete mutation action
	DeleteAction MutationAction = "delete"
)

// Mutation is a control interface for a mutation
type Mutation interface {
	Control
	// Config is any customer specific configuration for this customer
	Config() Config
	// State is a customer specific state object for this integration and customer
	State() State
	// CustomerID will return the customer id for the export
	CustomerID() string
	// IntegrationInstanceID will return the unique instance id for this integration for a customer
	IntegrationInstanceID() string
	// Pipe should be called to get the pipe for streaming data back to pinpoint
	Pipe() Pipe
	// ID is the primary key of the payload
	ID() string
	// Model is the name of the model of the payload
	Model() string
	// Action is the mutation action
	Action() MutationAction
	// Payload is the payload of the mutation which can be either a sdk.Model for create, sdk.PartialModel for update or nil for delete
	Payload() interface{}
}
