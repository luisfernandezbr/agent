package sdk

import (
	"encoding/json"
	"fmt"
	"strconv"

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
	RefID      string      `json:"ref_id"` // RefID is the id of the user in the source system
	OAuth2Auth *oauth2Auth `json:"oauth2_auth,omitempty"`
	OAuth1Auth *oauth1Auth `json:"oauth1_auth,omitempty"`
	BasicAuth  *basicAuth  `json:"basic_auth,omitempty"`
	APIKeyAuth *apikeyAuth `json:"apikey_auth,omitempty"`
}

// MutationData is the 'payload' field in agent.Mutation
type MutationData struct {
	RefID   string          `json:"ref_id"`  // RefID is the the ref_id of the model to update
	Model   string          `json:"model"`   // Model is the model name (eg. work.Issue)
	Action  MutationAction  `json:"action"`  // Action is either create, update, delete
	Payload json.RawMessage `json:"payload"` // Payload should be one of the Model Mutations defined below
	User    MutationUser    `json:"user"`    // User is a Mutation user on whom's behalf the mutation is being made
}

// Mutation is a control interface for a mutation
type Mutation interface {
	Control
	// Config is any customer specific configuration for this customer
	Config() Config
	// State is a customer specific state object for this integration and customer
	State() State
	// Pipe should be called to get the pipe for streaming data back to pinpoint
	Pipe() Pipe
	// ID returns the ref_id of the model to update
	ID() string
	// RefID returns the ref_id of the model to update
	RefID() string
	// Model is the name of the model of the payload
	Model() string
	// Action is the mutation action
	Action() MutationAction
	// Payload is the payload of the mutation which is one of the mutation types
	Payload() interface{}
	// User is the user that is requesting the mutation and any authorization details that might be required
	User() MutationUser
	// Logger the logger object to use in the integration
	Logger() Logger
}

// MutationResponse data
type MutationResponse struct {
	RefID      *string                // the ID of the entity in the source system
	EntityID   *string                // the entity ID in the pinpoint system for the entity that was mutated
	URL        *string                // optional url for the mutation entity in the source system
	Properties map[string]interface{} // optional properties to send in the result specific to the mutation type
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
			var payload AgileSprintCreateMutation
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
			var payload AgileSprintUpdateMutation
			err := json.Unmarshal(buf, &payload)
			return &payload, err
		}
	}
	return nil, nil
}

// Model Mutations

// SourcecodePullRequestUpdateMutation is an update mutation for a pull request
type SourcecodePullRequestUpdateMutation struct {
	Set struct {
		Title       *string                      `json:"title,omitempty"`       // Title is for updating the title to the pull request
		Description *string                      `json:"description,omitempty"` // Description is for updating the description of the pull request
		Status      *SourceCodePullRequestStatus `json:"status,omitempty"`      // Status is for changing the status of the pull request
	} `json:"set"`
}

// WorkIssueCreateMutation is a create mutation for a issue
type WorkIssueCreateMutation struct {
	Title         string     `json:"title"`                     // Title is for setting the title of the issue
	Description   string     `json:"description"`               // Description is for setting the description of the issue
	AssigneeRefID *string    `json:"assignee_ref_id,omitempty"` // AssigneeRefID is for setting the assignee of the issue to a ref_id
	Priority      *NameRefID `json:"priority,omitempty"`        // Priority is for setting the priority of the issue
	Type          *NameRefID `json:"type,omitempty"`            // Type is for setting the issue type of the issue
	ParentRefID   *string    `json:"parent_ref_id,omitempty"`   // ParentRefID is for setting the parent issue as a ref_id
	Epic          *NameRefID `json:"epic,omitempty"`            // Epic is for setting an epic for the issue
	Labels        []string   `json:"labels,omitempty"`          // Labels is for setting the labels for an issue

	// NOTE(robin): the above fields are for backwards compatibility, using MutationFields is the future 🚀
	ProjectRefID string               `json:"project_ref_id"` // ProjectID is the id to the issue project as a ref_id // DEPRECATED in favor of Project
	Project      NameRefID            `json:"project"`        // Project contains ref_id and name of the project
	Fields       []MutationFieldValue `json:"fields"`
}

// MutationFieldValue represents the value for the field, as defined in the project capabilities
type MutationFieldValue struct {
	RefID string                                       `json:"ref_id"`
	Type  WorkProjectCapabilityIssueMutationFieldsType `json:"type"`
	Value json.RawMessage                              `json:"value"`
}

// AsNameRefID will return m's value as a NameRefID if it's type is: WorkProjectCapabilityIssueMutationFieldsTypeWorkIssueType,
// WorkProjectCapabilityIssueMutationFieldsTypeWorkIssuePriority, or WorkProjectCapabilityIssueMutationFieldsTypeUser.
func (m MutationFieldValue) AsNameRefID() (*NameRefID, error) {
	if m.Type == WorkProjectCapabilityIssueMutationFieldsTypeWorkIssuePriority ||
		m.Type == WorkProjectCapabilityIssueMutationFieldsTypeWorkIssueType ||
		m.Type == WorkProjectCapabilityIssueMutationFieldsTypeEpic ||
		m.Type == WorkProjectCapabilityIssueMutationFieldsTypeStringArray ||
		m.Type == WorkProjectCapabilityIssueMutationFieldsTypeWorkSprint ||
		m.Type == WorkProjectCapabilityIssueMutationFieldsTypeUser {
		var nri NameRefID
		if err := json.Unmarshal(m.Value, &nri); err != nil {
			return nil, fmt.Errorf("error decoding mutation field %s into NameRefID: %w", m.RefID, err)
		}
		return &nri, nil
	}
	return nil, fmt.Errorf("type %s is not a NameRefID", m.Type.String())
}

// AsNumber will return m's value as an int if it's type is WorkProjectCapabilityIssueMutationFieldsTypeNumber.
func (m MutationFieldValue) AsNumber() (int, error) {
	if m.Type == WorkProjectCapabilityIssueMutationFieldsTypeNumber {
		return strconv.Atoi(string(m.Value))
	}
	return 0, fmt.Errorf("type %s is not a number", m.Type.String())
}

// AsString will return m's value as an int if it's type is WorkProjectCapabilityIssueMutationFieldsTypeString.
func (m MutationFieldValue) AsString() (string, error) {
	if m.Type == WorkProjectCapabilityIssueMutationFieldsTypeString ||
		m.Type == WorkProjectCapabilityIssueMutationFieldsTypeTextbox {
		var str string
		// you gotta unmarshal since it's a string
		if err := json.Unmarshal(m.Value, &str); err != nil {
			return "", fmt.Errorf("error decoding mutation field %s into string: %w", m.RefID, err)
		}
		return str, nil
	}
	return "", fmt.Errorf("type %s is not a string", m.Type.String())
}

// WorkIssueUpdateMutation is an update mutation for a issue
type WorkIssueUpdateMutation struct {
	Set struct {
		Title         *string    `json:"title"`                     // Title is for updating the title to the issue
		Transition    *NameRefID `json:"transition,omitempty"`      // Transition information (if used) for the issue
		Priority      *NameRefID `json:"priority,omitempty"`        // Priority is for changing the priority of the issue
		Resolution    *NameRefID `json:"resolution,omitempty"`      // Resolution is for changing the resolution of the issue
		Epic          *NameRefID `json:"epic,omitempty"`            // Epic is for updating the epic for the issue
		AssigneeRefID *string    `json:"assignee_ref_id,omitempty"` // AssigneeRefID is for changing the assignee of the issue to a ref_id
	} `json:"set"`
	Unset struct {
		Epic     bool `json:"epic"`     // Epic is for removing the epic from the issue (if set to true)
		Assignee bool `json:"assignee"` // Assignee is for removing the assignee from the issue (if set to true)
	} `json:"unset"`
}

const (
	// WorkIssueTransitionRequiresResolution tells the ui that resolution is required to make a transition
	WorkIssueTransitionRequiresResolution = "resolution"
)

// AgileSprintCreateMutation is an create mutation for a sprint
type AgileSprintCreateMutation struct {
	Name         string            `json:"name"`             // Name is the name of the sprint
	Goal         *string           `json:"goal,omitempty"`   // Goal is the optional goal for the sprint
	Status       AgileSprintStatus `json:"status,omitempty"` // Status is the status of the sprint
	StartDate    Date              `json:"start_date"`       // StartDate is the start date for the sprint
	EndDate      Date              `json:"end_date"`         // EndDate is the end date for the sprint
	IssueRefIDs  []string          `json:"issue_ref_ids"`    // IssueRefIDs is an array of issue ref_ids to add to the sprint
	ProjectRefID *string           `json:"project_ref_id"`   // ProjectRefID is the id to the issue project as a ref_id
	BoardRefIDs  []string          `json:"board_ref_ids"`    // BoardRefIDs are the ids of the boards to link the sprint to, required by some source systems
}

// AgileSprintUpdateMutation is an update mutation for a sprint
type AgileSprintUpdateMutation struct {
	Set struct {
		Name        *string            `json:"name,omitempty"`          // Name is the name of the sprint to update
		Goal        *string            `json:"goal,omitempty"`          // Goal is the optional goal for the sprint
		Status      *AgileSprintStatus `json:"status,omitempty"`        // Status is the status of the sprint
		StartDate   *Date              `json:"start_date,omitempty"`    // StartDate is the start date for the sprint
		EndDate     *Date              `json:"end_date,omitempty"`      // EndDate is the end date for the sprint
		IssueRefIDs []string           `json:"issue_ref_ids,omitempty"` // IssueRefIDs is an array of issue ref_ids to add to the sprint
	} `json:"set"`
	Unset struct {
		IssueRefIDs []string `json:"issue_ref_ids,omitempty"` // IssueRefIDs is an array of issue ref_ids to remove from the sprint
	} `json:"unset"`
}
