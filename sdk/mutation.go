package sdk

import (
	"encoding/json"
	"time"

	"github.com/pinpt/integration-sdk/sourcecode"
	"github.com/pinpt/integration-sdk/work"
)

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

// MutationUser is the user that is requesting the mutation
type MutationUser struct {
	// ID is the ref_id to the source system
	ID         string      `json:"id"`
	OAuth2Auth *oauth2Auth `json:"oauth2_auth,omitempty"`
	BasicAuth  *basicAuth  `json:"basic_auth,omitempty"`
	APIKeyAuth *apikeyAuth `json:"apikey_auth,omitempty"`
}

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
	// Payload is the payload of the mutation which is one of the mutation types
	Payload() interface{}
	// User is the user that is requesting the mutation and any authorization details that might be required
	User() MutationUser
}

// CreateMutationPayloadFromData will create a mutation payload object from a data payload
func CreateMutationPayloadFromData(model string, action MutationAction, buf []byte) (interface{}, error) {
	switch action {
	case CreateAction:
		switch model {
		case work.IssueModelName.String():
			var payload WorkIssueCreateMutation
			err := json.Unmarshal(buf, &payload)
			return &payload, err
		case work.SprintModelName.String():
			var payload WorkSprintCreateMutation
			err := json.Unmarshal(buf, &payload)
			return &payload, err
		}
	case UpdateAction:
		switch model {
		case sourcecode.PullRequestModelName.String():
			var payload SourcecodePullRequestUpdateMutation
			err := json.Unmarshal(buf, &payload)
			return &payload, err
		case work.IssueModelName.String():
			var payload WorkIssueUpdateMutation
			err := json.Unmarshal(buf, &payload)
			return &payload, err
		case work.SprintModelName.String():
			var payload WorkSprintUpdateMutation
			err := json.Unmarshal(buf, &payload)
			return &payload, err
		}
	}
	return nil, nil
}

// NameID is a container for containing the ID, Name or both
type NameID struct {
	ID   *string `json:"id,omitempty"`
	Name *string `json:"name,omitempty"`
}

// SourcecodePullRequestUpdateMutation is an update mutation for a pull request
type SourcecodePullRequestUpdateMutation struct {
	Title       *string                      `json:"title,omitempty"`       // Title is for updating the title to the pull request
	Description *string                      `json:"description,omitempty"` // Description is for updating the description of the pull request
	Status      *SourceCodePullRequestStatus `json:"status,omitempty"`      // Status is for changing the status of the pull request
}

// WorkIssueCreateMutation is a create mutation for a issue
type WorkIssueCreateMutation struct {
	Title         string   `json:"title"`               // Title is for setting the title of the issue
	Description   string   `json:"description"`         // Description is for setting the description of the issue
	AssigneeRefID *string  `json:"assignee,omitempty"`  // AssigneeRefID is for setting the assignee of the issue to a ref_id
	Priority      *NameID  `json:"priority,omitempty"`  // Priority is for setting the priority of the issue
	Type          *NameID  `json:"type,omitempty"`      // Type is for setting the issue type of the issue
	ProjectRefID  string   `json:"project_id"`          // ProjectID is the id to the issue project as a ref_id
	Epic          *NameID  `json:"epic,omitempty"`      // Epic is for setting an epic for the issue
	ParentRefID   *string  `json:"parent_id,omitempty"` // ParentRefID is for setting the parent issue as a ref_id
	Labels        []string `json:"labels,omitempty"`    // Labels is for setting the labels for an issue
}

// WorkIssueUpdateMutation is an update mutation for a issue
type WorkIssueUpdateMutation struct {
	Set struct {
		Title         *string `json:"title"`                // Title is for updating the title to the issue
		Transition    *NameID `json:"transition,omitempty"` // Transition information (if used) for the issue
		Status        *NameID `json:"status,omitempty"`     // Status is for changing the status of the issue
		Priority      *NameID `json:"priority,omitempty"`   // Priority is for changing the priority of the issue
		Resolution    *NameID `json:"resolution,omitempty"` // Resolution is for changing the resolution of the issue
		Epic          *NameID `json:"epic,omitempty"`       // Epic is for updating the epic for the issue
		AssigneeRefID *string `json:"assignee,omitempty"`   // AssigneeRefID is for changing the assignee of the issue to a ref_id
	} `json:"set"`
	Unset struct {
		Epic     bool `json:"epic"`     // Epic is for removing the epic from the issue (if set to true)
		Assignee bool `json:"assignee"` // Assignee is for removing the assignee from the issue (if set to true)
	} `json:"unset"`
}

// WorkSprintCreateMutation is an create mutation for a sprint
type WorkSprintCreateMutation struct {
	Name         string           `json:"name"`             // Name is the name of the sprint
	Goal         *string          `json:"goal,omitempty"`   // Goal is the optional goal for the sprint
	Status       WorkSprintStatus `json:"status,omitempty"` // Status is the status of the sprint
	StartDate    time.Time        `json:"start_date"`       // StartDate is the start date for the sprint
	EndDate      time.Time        `json:"end_date"`         // EndDate is the end date for the sprint
	IssueRefIDs  []string         `json:"issue_ref_ids"`    // IssueRefIDs is an array of issue ref_ids to add to the sprint
	ProjectRefID string           `json:"project_id"`       // ProjectID is the id to the issue project as a ref_id
}

// WorkSprintUpdateMutation is an update mutation for a sprint
type WorkSprintUpdateMutation struct {
	Set struct {
		Name        *string           `json:"name,omitempty"`          // Name is the name of the sprint to update
		Goal        *string           `json:"goal,omitempty"`          // Goal is the optional goal for the sprint
		Status      *WorkSprintStatus `json:"status,omitempty"`        // Status is the status of the sprint
		StartDate   *time.Time        `json:"start_date,omitempty"`    // StartDate is the start date for the sprint
		EndDate     *time.Time        `json:"end_date,omitempty"`      // EndDate is the end date for the sprint
		IssueRefIDs []string          `json:"issue_ref_ids,omitempty"` // IssueRefIDs is an array of issue ref_ids to add to the sprint
	} `json:"set"`
	Unset struct {
		IssueRefIDs []string `json:"issue_ref_ids,omitempty"` // IssueRefIDs is an array of issue ref_ids to remove from the sprint
	} `json:"unset"`
}
