package sdk

import (
	"time"

	"github.com/pinpt/go-common/v10/datetime"
	"github.com/pinpt/integration-sdk/agent"
	"github.com/pinpt/integration-sdk/work"
)

// NameRefID is a container for containing the RefID, Name or both
type NameRefID struct {
	RefID *string `json:"ref_id,omitempty"`
	Name  *string `json:"name,omitempty"`
}

// WorkIssueUpdate is an action for update a work.Issue
type WorkIssueUpdate struct {
	Set struct {
		Title            *string
		Description      *string
		Active           *bool
		StoryPoints      *float32
		Identifier       *string
		ProjectID        *string
		URL              *string
		DueDate          *time.Time
		Priority         *NameID
		Type             *NameID
		Status           *NameID
		AssigneeRefID    *string
		ParentID         *string
		Tags             *[]string
		EpicID           *string
		Resolution       *string
		PlannedStartDate *time.Time
		PlannedEndDate   *time.Time
		SprintIDs        *[]string
		Transitions      *[]NameRefID
	}
	Unset struct {
		StoryPoints      *bool
		DueDate          *bool
		ParentID         *bool
		EpicID           *bool
		PlannedStartDate *bool
		PlannedEndDate   *bool
	}
	Push struct {
		Tags       *[]string
		SprintIDs  *[]string
		ChangeLogs *[]WorkIssueChangeLog
	}
	Pull struct {
		Tags      *[]string
		SprintIDs *[]string
	}
}

// NewWorkIssueUpdate will create a new update object for work.Issue which can be sent to an sdk.Pipe using Write
func NewWorkIssueUpdate(customerID string, integrationInstanceID string, refID string, refType string, val WorkIssueUpdate) Model {
	data := &agent.UpdateData{
		ID:                    NewWorkIssueID(customerID, refID, refType),
		CustomerID:            customerID,
		RefID:                 refID,
		RefType:               refType,
		IntegrationInstanceID: StringPointer(integrationInstanceID),
		Model:                 work.IssueModelName.String(),
		Set:                   make(map[string]string),
		Unset:                 make([]string, 0),
		Push:                  make(map[string]string),
		Pull:                  make(map[string]string),
	}
	// setters
	if val.Set.Active != nil {
		data.Set[work.IssueModelActiveColumn] = Stringify(val.Set.Active)
	}
	if val.Set.Title != nil {
		data.Set["title"] = Stringify(val.Set.Title)
	}
	if val.Set.Description != nil {
		data.Set["description"] = Stringify(val.Set.Description)
	}
	if val.Set.StoryPoints != nil {
		data.Set["story_points"] = Stringify(val.Set.StoryPoints)
	}
	if val.Set.Identifier != nil {
		data.Set["identifier"] = Stringify(val.Set.Identifier)
	}
	if val.Set.ProjectID != nil {
		data.Set["project_id"] = Stringify(val.Set.ProjectID)
	}
	if val.Set.URL != nil {
		data.Set["url"] = Stringify(val.Set.URL)
	}
	if val.Set.DueDate != nil {
		data.Set["due_date"] = Stringify(NewDateWithTime(*val.Set.DueDate))
	}
	if val.Set.Priority != nil {
		data.Set["priority"] = Stringify(val.Set.Priority)
	}
	if val.Set.Type != nil {
		data.Set["type"] = Stringify(val.Set.Type)
	}
	if val.Set.Status != nil {
		data.Set["status"] = Stringify(val.Set.Status)
	}
	if val.Set.AssigneeRefID != nil {
		data.Set["assignee_ref_id"] = Stringify(val.Set.AssigneeRefID)
	}
	if val.Set.ParentID != nil {
		data.Set["parent_id"] = Stringify(val.Set.ParentID)
	}
	if val.Set.Tags != nil {
		data.Set["tags"] = Stringify(*val.Set.Tags)
	}
	if val.Set.EpicID != nil {
		data.Set["epic_id"] = Stringify(val.Set.EpicID)
	}
	if val.Set.Resolution != nil {
		data.Set["resolution"] = Stringify(val.Set.Resolution)
	}
	if val.Set.PlannedStartDate != nil {
		data.Set["planned_start_date"] = Stringify(NewDateWithTime(*val.Set.PlannedStartDate))
	}
	if val.Set.PlannedEndDate != nil {
		data.Set["planned_end_date"] = Stringify(NewDateWithTime(*val.Set.PlannedEndDate))
	}
	if val.Set.SprintIDs != nil {
		data.Set["sprint_ids"] = Stringify(*val.Set.SprintIDs)
	}
	if val.Set.Transitions != nil {
		data.Set["transitions"] = Stringify(*val.Set.Transitions)
	}
	// unsetters
	if val.Unset.StoryPoints != nil {
		data.Unset = append(data.Unset, "story_points")
	}
	if val.Unset.DueDate != nil {
		data.Unset = append(data.Unset, "due_date")
	}
	if val.Unset.ParentID != nil {
		data.Unset = append(data.Unset, "parent_id")
	}
	if val.Unset.EpicID != nil {
		data.Unset = append(data.Unset, "epic_id")
	}
	if val.Unset.PlannedStartDate != nil {
		data.Unset = append(data.Unset, "planned_start_date")
	}
	if val.Unset.PlannedEndDate != nil {
		data.Unset = append(data.Unset, "planned_end_date")
	}
	// pushers
	if val.Push.Tags != nil {
		data.Push["tags"] = Stringify(*val.Push.Tags)
	}
	if val.Push.SprintIDs != nil {
		data.Push["sprint_ids"] = Stringify(*val.Push.SprintIDs)
	}
	if val.Push.ChangeLogs != nil {
		data.Push["change_log"] = Stringify(*val.Push.ChangeLogs)
	}
	// pullers
	if val.Pull.Tags != nil {
		data.Pull["tags"] = Stringify(*val.Pull.Tags)
	}
	if val.Pull.SprintIDs != nil {
		data.Pull["sprint_ids"] = Stringify(*val.Pull.SprintIDs)
	}
	// always set the updated_date when updating
	data.Set["updated_date"] = Stringify(datetime.NewDateNow())

	// TODO: attachments, linked_issues

	// fmt.Println(StringifyPretty(data))

	return data
}

// WorkIssueCommentUpdate is an action for update a work.IssueComment
type WorkIssueCommentUpdate struct {
	Set struct {
		Active *bool
	}
	Unset struct {
	}
	Push struct {
	}
	Pull struct {
	}
}

// NewWorkIssueCommentUpdate will create a new update object for work.IssueComment which can be sent to an sdk.Pipe using Write
func NewWorkIssueCommentUpdate(customerID string, integrationInstanceID string, refID string, refType string, projectRefID string, val WorkIssueCommentUpdate) Model {
	projectID := NewWorkProjectID(customerID, projectRefID, refType)
	data := &agent.UpdateData{
		ID:                    NewWorkIssueCommentID(customerID, refID, refType, projectID),
		CustomerID:            customerID,
		RefID:                 refID,
		RefType:               refType,
		IntegrationInstanceID: StringPointer(integrationInstanceID),
		Model:                 work.IssueCommentModelName.String(),
		Set:                   make(map[string]string),
		Unset:                 make([]string, 0),
		Push:                  make(map[string]string),
		Pull:                  make(map[string]string),
	}
	// setters
	if val.Set.Active != nil {
		data.Set[work.IssueCommentModelActiveColumn] = Stringify(val.Set.Active)
	}
	// unsetters

	// pushers

	// pullers

	// always set the updated_date when updating
	data.Set[work.IssueCommentModelUpdatedDateColumn] = Stringify(datetime.NewDateNow())

	return data
}

// WorkProjectUpdate is an action for update a work.Project
type WorkProjectUpdate struct {
	Set struct {
		Active      *bool
		Name        *string
		Description *string
	}
	Unset struct {
	}
	Push struct {
	}
	Pull struct {
	}
}

// NewWorkProjectUpdate will create a new update object for work.Project which can be sent to an sdk.Pipe using Write
func NewWorkProjectUpdate(customerID string, integrationInstanceID string, refID string, refType string, val WorkProjectUpdate) Model {
	data := &agent.UpdateData{
		ID:                    NewWorkProjectID(customerID, refID, refType),
		CustomerID:            customerID,
		RefID:                 refID,
		RefType:               refType,
		IntegrationInstanceID: StringPointer(integrationInstanceID),
		Model:                 work.ProjectModelName.String(),
		Set:                   make(map[string]string),
		Unset:                 make([]string, 0),
		Push:                  make(map[string]string),
		Pull:                  make(map[string]string),
	}
	// setters
	if val.Set.Active != nil {
		data.Set[work.ProjectModelActiveColumn] = Stringify(val.Set.Active)
	}
	if val.Set.Name != nil {
		data.Set[work.ProjectModelNameColumn] = Stringify(val.Set.Name)
	}
	if val.Set.Description != nil {
		data.Set[work.ProjectModelDescriptionColumn] = Stringify(val.Set.Description)
	}
	// unsetters

	// pushers

	// pullers

	// always set the updated_date when updating
	data.Set[work.ProjectModelUpdatedDateColumn] = Stringify(datetime.NewDateNow())

	return data
}

// AgileSprintUpdate is an action for update a work.Sprint
type AgileSprintUpdate struct {
	Set struct {
		Active      *bool
		Name        *string
		Goal        *string
		StartedDate *time.Time
		EndedDate   *time.Time
		Status      *work.SprintStatus
	}
	Unset struct {
		Goal *string
	}
	Push struct {
	}
	Pull struct {
	}
}

// NewAgileSprintUpdate will create a new update object for work.Sprint which can be sent to an sdk.Pipe using Write
func NewAgileSprintUpdate(customerID string, integrationInstanceID string, refID string, refType string, val AgileSprintUpdate) Model {
	data := &agent.UpdateData{
		ID:                    NewAgileSprintID(customerID, refID, refType),
		CustomerID:            customerID,
		RefID:                 refID,
		RefType:               refType,
		IntegrationInstanceID: StringPointer(integrationInstanceID),
		Model:                 work.SprintModelName.String(),
		Set:                   make(map[string]string),
		Unset:                 make([]string, 0),
		Push:                  make(map[string]string),
		Pull:                  make(map[string]string),
	}

	// setters
	if val.Set.Active != nil {
		data.Set[work.SprintModelActiveColumn] = Stringify(val.Set.Active)
	}
	if val.Set.Name != nil {
		data.Set[work.SprintModelNameColumn] = Stringify(val.Set.Name)
	}
	if val.Set.Goal != nil {
		data.Set[work.SprintModelGoalColumn] = Stringify(val.Set.Goal)
	}
	if val.Set.StartedDate != nil {
		data.Set[work.SprintModelStartedDateColumn] = Stringify(NewDateWithTime(*val.Set.StartedDate))
	}
	if val.Set.EndedDate != nil {
		data.Set[work.SprintModelEndedDateColumn] = Stringify(NewDateWithTime(*val.Set.EndedDate))
	}
	// unsetters
	if val.Unset.Goal != nil {
		data.Unset = append(data.Unset, work.SprintModelGoalColumn)
	}
	// pushers

	// pullers

	// always set the updated_date when updating
	data.Set[work.SprintModelUpdatedDateColumn] = Stringify(datetime.NewDateNow())

	return data
}
