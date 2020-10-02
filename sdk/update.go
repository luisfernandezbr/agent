package sdk

import (
	"time"

	"github.com/pinpt/go-common/v10/datetime"
	"github.com/pinpt/integration-sdk/agent"
	"github.com/pinpt/integration-sdk/sourcecode"
	"github.com/pinpt/integration-sdk/work"
)

// NameRefID is a container for containing the RefID, Name or both
type NameRefID struct {
	RefID *string `json:"ref_id,omitempty"`
	Name  *string `json:"name,omitempty"`
}

// NameID is a container for containing the ID, Name or both
type NameID struct {
	ID   *string `json:"id,omitempty"`
	Name *string `json:"name,omitempty"`
}

// SourceCodePullRequestReviewRequestUpdate is an action for update a sourcecode.PullRequestReviewRequest
type SourceCodePullRequestReviewRequestUpdate struct {
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

// NewSourceCodePullRequestReviewRequestUpdate will create a new update object for sourcecode.PullRequestReviewRequest which can be sent to an sdk.Pipe using Write
func NewSourceCodePullRequestReviewRequestUpdate(customerID string, integrationInstanceID string, pullRequestReviewRequestID string, refType string, val SourceCodePullRequestReviewRequestUpdate) Model {
	data := &agent.UpdateData{
		ID:                    pullRequestReviewRequestID,
		CustomerID:            customerID,
		RefType:               refType,
		IntegrationInstanceID: StringPointer(integrationInstanceID),
		Model:                 sourcecode.PullRequestReviewRequestModelName.String(),
		Set:                   make(map[string]string),
		Unset:                 make([]string, 0),
		Push:                  make(map[string]string),
		Pull:                  make(map[string]string),
	}

	// setters
	if val.Set.Active != nil {
		data.Set[sourcecode.PullRequestReviewRequestModelActiveColumn] = Stringify(val.Set.Active)
	}
	// unsetters

	// pushers

	// pullers

	// always set the updated_date when updating
	// FIXME(robin): add this to review request
	// data.Set[sourcecode.PullRequestReviewRequestModelUpdatedDateColumn] = Stringify(datetime.NewDateNow())

	return data
}

// NewSourceCodePullRequestReviewRequestDeactivate will create a new update object that sets active to false for sourcecode.PullRequestReviewRequest which can be sent to an sdk.Pipe using Write
func NewSourceCodePullRequestReviewRequestDeactivate(customerID string, integrationInstanceID string, pullRequestReviewRequestID string, refType string) Model {
	var update SourceCodePullRequestReviewRequestUpdate
	update.Set.Active = BoolPointer(false)
	return NewSourceCodePullRequestReviewRequestUpdate(customerID, integrationInstanceID, pullRequestReviewRequestID, refType, update)
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
		Transitions      *[]WorkIssueTransitions
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
		Tags         *[]string
		SprintIDs    *[]string
		ChangeLogs   *[]WorkIssueChangeLog
		LinkedIssues *[]WorkIssueLinkedIssues
	}
	Pull struct {
		Tags         *[]string
		SprintIDs    *[]string
		LinkedIssues *[]WorkIssueLinkedIssues
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
		data.Set[work.IssueModelPriorityColumn] = Stringify(val.Set.Priority.Name)
		data.Set[work.IssueModelPriorityIDColumn] = Stringify(val.Set.Priority.ID)
	}
	if val.Set.Type != nil {
		data.Set[work.IssueModelTypeColumn] = Stringify(val.Set.Type.Name)
		data.Set[work.IssueModelTypeIDColumn] = Stringify(val.Set.Type.ID)
	}
	if val.Set.Status != nil {
		data.Set[work.IssueModelStatusColumn] = Stringify(val.Set.Status.Name)
		data.Set[work.IssueModelStatusIDColumn] = Stringify(val.Set.Status.ID)
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
		if *val.Set.Transitions != nil {
			data.Set["transitions"] = Stringify(*val.Set.Transitions)
		} else {
			data.Set["transitions"] = Stringify([]WorkIssueTransitions{})
		}
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
	if val.Push.LinkedIssues != nil {
		data.Push[work.IssueModelLinkedIssuesColumn] = Stringify(*val.Push.LinkedIssues)
	}
	// pullers
	if val.Pull.Tags != nil {
		data.Pull["tags"] = Stringify(*val.Pull.Tags)
	}
	if val.Pull.SprintIDs != nil {
		data.Pull["sprint_ids"] = Stringify(*val.Pull.SprintIDs)
	}
	if val.Pull.LinkedIssues != nil {
		data.Pull[work.IssueModelLinkedIssuesColumn] = Stringify(*val.Pull.LinkedIssues)
	}
	// always set the updated_date when updating
	data.Set["updated_date"] = Stringify(datetime.NewDateNow())

	// TODO: attachments, linked_issues

	// fmt.Println(StringifyPretty(data))

	return data
}

// NewWorkIssueDeactivate will create a new update object that sets active to false for work.Issue which can be sent to an sdk.Pipe using Write
func NewWorkIssueDeactivate(customerID string, integrationInstanceID string, refID string, refType string) Model {
	var update WorkIssueUpdate
	update.Set.Active = BoolPointer(false)
	return NewWorkIssueUpdate(customerID, integrationInstanceID, refID, refType, update)
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
func NewWorkIssueCommentUpdate(customerID string, integrationInstanceID string, refID string, refType string, val WorkIssueCommentUpdate) Model {
	data := &agent.UpdateData{
		ID:                    NewWorkIssueCommentID(customerID, refID, refType),
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

// NewWorkIssueCommentDeactivate will create a new update object that sets active to false for work.IssueComment which can be sent to an sdk.Pipe using Write
func NewWorkIssueCommentDeactivate(customerID string, integrationInstanceID string, refID string, refType string) Model {
	var update WorkIssueCommentUpdate
	update.Set.Active = BoolPointer(false)
	return NewWorkIssueCommentUpdate(customerID, integrationInstanceID, refID, refType, update)
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

// NewWorkProjectDeactivate will create a new update object that sets active to false for work.Project which can be sent to an sdk.Pipe using Write
func NewWorkProjectDeactivate(customerID string, integrationInstanceID string, refID string, refType string) Model {
	var update WorkProjectUpdate
	update.Set.Active = BoolPointer(false)
	return NewWorkProjectUpdate(customerID, integrationInstanceID, refID, refType, update)
}

// AgileSprintUpdate is an action for update a work.Sprint
type AgileSprintUpdate struct {
	Set struct {
		Active        *bool
		Name          *string
		Goal          *string
		StartedDate   *time.Time
		CompletedDate *time.Time
		EndedDate     *time.Time
		Status        *work.SprintStatus
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
	if val.Set.Status != nil {
		data.Set[work.SprintModelStatusColumn] = Stringify(val.Set.Status)
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
	if val.Set.CompletedDate != nil {
		data.Set[work.SprintModelCompletedDateColumn] = Stringify(NewDateWithTime(*val.Set.CompletedDate))
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

// NewAgileSprintDeactivate will create a new update object that sets active to false for Agile.Sprint which can be sent to an sdk.Pipe using Write
func NewAgileSprintDeactivate(customerID string, integrationInstanceID string, refID string, refType string) Model {
	var update AgileSprintUpdate
	update.Set.Active = BoolPointer(false)
	return NewAgileSprintUpdate(customerID, integrationInstanceID, refID, refType, update)
}

// AgileBoardUpdate is an action for update a work.Board
type AgileBoardUpdate struct {
	Set struct {
		Active *bool
		Name   *string
	}
	Unset struct {
	}
	Push struct {
	}
	Pull struct {
	}
}

// NewAgileBoardUpdate will create a new update object for work.Board which can be sent to an sdk.Pipe using Write
func NewAgileBoardUpdate(customerID string, integrationInstanceID string, refID string, refType string, val AgileBoardUpdate) Model {
	data := &agent.UpdateData{
		ID:                    NewAgileBoardID(customerID, refID, refType),
		CustomerID:            customerID,
		RefID:                 refID,
		RefType:               refType,
		IntegrationInstanceID: StringPointer(integrationInstanceID),
		Model:                 work.BoardModelName.String(),
		Set:                   make(map[string]string),
		Unset:                 make([]string, 0),
		Push:                  make(map[string]string),
		Pull:                  make(map[string]string),
	}

	// setters
	if val.Set.Active != nil {
		data.Set[work.BoardModelActiveColumn] = Stringify(val.Set.Active)
	}
	if val.Set.Name != nil {
		data.Set[work.BoardModelNameColumn] = Stringify(val.Set.Name)
	}
	// unsetters

	// pushers

	// pullers

	// always set the updated_date when updating
	data.Set[work.BoardModelUpdatedDateColumn] = Stringify(datetime.NewDateNow())

	return data
}

// NewAgileBoardDeactivate will create a new update object that sets active to false for Agile.Board which can be sent to an sdk.Pipe using Write
func NewAgileBoardDeactivate(customerID string, integrationInstanceID string, refID string, refType string) Model {
	var update AgileBoardUpdate
	update.Set.Active = BoolPointer(false)
	return NewAgileBoardUpdate(customerID, integrationInstanceID, refID, refType, update)
}
