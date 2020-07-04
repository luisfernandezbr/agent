package sdk

import "github.com/pinpt/integration-sdk/work"

// Type aliases for our exported datamodel types to create a stable version
// which Integrations depend on instead of directly depending on a specific
// version of the integration-sdk directly

// WorkIssue is a issue
type WorkIssue = work.Issue

// WorkIssuePartial is a issue partial
type WorkIssuePartial = work.IssuePartial

// WorkIssueComment is a issue comment
type WorkIssueComment = work.IssueComment

// WorkIssueCommentPartial is a issue comment partial
type WorkIssueCommentPartial = work.IssueCommentPartial

// WorkIssueStatus is a issue status
type WorkIssueStatus = work.IssueStatus

// WorkIssueStatusPartial is a issue status partial
type WorkIssueStatusPartial = work.IssueStatusPartial

// WorkIssuePriority is a issue priority
type WorkIssuePriority = work.IssuePriority

// WorkIssuePriorityPartial is a issue priority partial
type WorkIssuePriorityPartial = work.IssuePriorityPartial

// WorkIssueType is a issue type
type WorkIssueType = work.IssueType

// WorkIssueTypePartial is a issue type partial
type WorkIssueTypePartial = work.IssueTypePartial

// WorkProject is a project
type WorkProject = work.Project

// WorkProjectPartial is a project partial
type WorkProjectPartial = work.ProjectPartial

// WorkSprint is a sprint
type WorkSprint = work.Sprint

// WorkSprintPartial is a sprint partial
type WorkSprintPartial = work.SprintPartial

// WorkKanbanBoard is a kanban board
type WorkKanbanBoard = work.KanbanBoard

// WorkKanbanBoardPartial is a kanban board partial
type WorkKanbanBoardPartial = work.KanbanBoardPartial

// WorkUser is a user in the work system
type WorkUser = work.User

// WorkUserPartial is a user in the work system partial
type WorkUserPartial = work.UserPartial

// WorkIssueAttachments is the work issue attachments
type WorkIssueAttachments = work.IssueAttachments

// WorkIssueChangeLog is the issue changelog
type WorkIssueChangeLog = work.IssueChangeLog

// WorkIssueChangeLogCreatedDate is the issue change log created date
type WorkIssueChangeLogCreatedDate = work.IssueChangeLogCreatedDate

// WorkIssueChangeLogField is the issue change log field enum
type WorkIssueChangeLogField = work.IssueChangeLogField

// WorkIssueChangeLogFieldAssigneeRefID is the enumeration value for assignee_ref_id
const WorkIssueChangeLogFieldAssigneeRefID WorkIssueChangeLogField = work.IssueChangeLogFieldAssigneeRefID

// WorkIssueChangeLogFieldDueDate is the enumeration value for due_date
const WorkIssueChangeLogFieldDueDate WorkIssueChangeLogField = work.IssueChangeLogFieldDueDate

// WorkIssueChangeLogFieldEpicID is the enumeration value for epic_id
const WorkIssueChangeLogFieldEpicID WorkIssueChangeLogField = work.IssueChangeLogFieldEpicID

// WorkIssueChangeLogFieldIdentifier is the enumeration value for identifier
const WorkIssueChangeLogFieldIdentifier WorkIssueChangeLogField = work.IssueChangeLogFieldIdentifier

// WorkIssueChangeLogFieldParentID is the enumeration value for parent_id
const WorkIssueChangeLogFieldParentID WorkIssueChangeLogField = work.IssueChangeLogFieldParentID

// WorkIssueChangeLogFieldPriority is the enumeration value for priority
const WorkIssueChangeLogFieldPriority WorkIssueChangeLogField = work.IssueChangeLogFieldPriority

// WorkIssueChangeLogFieldProjectID is the enumeration value for project_id
const WorkIssueChangeLogFieldProjectID WorkIssueChangeLogField = work.IssueChangeLogFieldProjectID

// WorkIssueChangeLogFieldReporterRefID is the enumeration value for reporter_ref_id
const WorkIssueChangeLogFieldReporterRefID WorkIssueChangeLogField = work.IssueChangeLogFieldReporterRefID

// WorkIssueChangeLogFieldResolution is the enumeration value for resolution
const WorkIssueChangeLogFieldResolution WorkIssueChangeLogField = work.IssueChangeLogFieldResolution

// WorkIssueChangeLogFieldSprintIds is the enumeration value for sprint_ids
const WorkIssueChangeLogFieldSprintIds WorkIssueChangeLogField = work.IssueChangeLogFieldSprintIds

// WorkIssueChangeLogFieldStatus is the enumeration value for status
const WorkIssueChangeLogFieldStatus WorkIssueChangeLogField = work.IssueChangeLogFieldStatus

// WorkIssueChangeLogFieldTags is the enumeration value for tags
const WorkIssueChangeLogFieldTags WorkIssueChangeLogField = work.IssueChangeLogFieldTags

// WorkIssueChangeLogFieldTitle is the enumeration value for title
const WorkIssueChangeLogFieldTitle WorkIssueChangeLogField = work.IssueChangeLogFieldTitle

// WorkIssueChangeLogFieldType is the enumeration value for type
const WorkIssueChangeLogFieldType WorkIssueChangeLogField = work.IssueChangeLogFieldType

// WorkIssueCreatedDate is the issue created date
type WorkIssueCreatedDate = work.IssueCreatedDate

// WorkIssueDueDate is the issue due date
type WorkIssueDueDate = work.IssueDueDate

// WorkIssueLinkedIssues is the issue linked issues
type WorkIssueLinkedIssues = work.IssueLinkedIssues

// WorkIssueLinkedIssuesLinkType is the linked isuse link type enum
type WorkIssueLinkedIssuesLinkType = work.IssueLinkedIssuesLinkType

// WorkIssueLinkedIssuesLinkTypeBlocks is the enumeration value for blocks
const WorkIssueLinkedIssuesLinkTypeBlocks WorkIssueLinkedIssuesLinkType = work.IssueLinkedIssuesLinkTypeBlocks

// WorkIssueLinkedIssuesLinkTypeClones is the enumeration value for clones
const WorkIssueLinkedIssuesLinkTypeClones WorkIssueLinkedIssuesLinkType = work.IssueLinkedIssuesLinkTypeClones

// WorkIssueLinkedIssuesLinkTypeDuplicates is the enumeration value for duplicates
const WorkIssueLinkedIssuesLinkTypeDuplicates WorkIssueLinkedIssuesLinkType = work.IssueLinkedIssuesLinkTypeDuplicates

// WorkIssueLinkedIssuesLinkTypeCauses is the enumeration value for causes
const WorkIssueLinkedIssuesLinkTypeCauses WorkIssueLinkedIssuesLinkType = work.IssueLinkedIssuesLinkTypeCauses

// WorkIssueLinkedIssuesLinkTypeRelates is the enumeration value for relates
const WorkIssueLinkedIssuesLinkTypeRelates WorkIssueLinkedIssuesLinkType = work.IssueLinkedIssuesLinkTypeRelates

// WorkIssuePlannedEndDate is the issue planned end date
type WorkIssuePlannedEndDate = work.IssuePlannedEndDate

// WorkIssuePlannedStartDate is the issue planned start date
type WorkIssuePlannedStartDate = work.IssuePlannedStartDate

// WorkIssueUpdatedDate is the issue updated date
type WorkIssueUpdatedDate = work.IssueUpdatedDate

// WorkIssueCommentCreatedDate is the issue comment created date
type WorkIssueCommentCreatedDate = work.IssueCommentCreatedDate

// WorkIssueCommentUpdatedDate is the issue comment updated date
type WorkIssueCommentUpdatedDate = work.IssueCommentUpdatedDate

// WorkSprintCompletedDate is the sprint completed date
type WorkSprintCompletedDate = work.SprintCompletedDate

// WorkSprintEndedDate is the sprint ended date
type WorkSprintEndedDate = work.SprintEndedDate

// WorkSprintStartedDate is the sprint started date
type WorkSprintStartedDate = work.SprintStartedDate

// WorkSprintStatus is the sprint status enum
type WorkSprintStatus = work.SprintStatus

// WorkSprintStatusActive is the enumeration value for active
const WorkSprintStatusActive WorkSprintStatus = work.SprintStatusActive

// WorkSprintStatusFuture is the enumeration value for future
const WorkSprintStatusFuture WorkSprintStatus = work.SprintStatusFuture

// WorkSprintStatusClosed is the enumeration value for closed
const WorkSprintStatusClosed WorkSprintStatus = work.SprintStatusClosed

// WorkKanbanBoardColumns is the kanban board columns
type WorkKanbanBoardColumns = work.KanbanBoardColumns

// WorkIssueTypeMappedType is the issue type mapped type enum
type WorkIssueTypeMappedType = work.IssueTypeMappedType

// WorkIssueTypeMappedTypeUnknown is the enumeration value for unknown
const WorkIssueTypeMappedTypeUnknown WorkIssueTypeMappedType = work.IssueTypeMappedTypeUnknown

// WorkIssueTypeMappedTypeFeature is the enumeration value for feature
const WorkIssueTypeMappedTypeFeature WorkIssueTypeMappedType = work.IssueTypeMappedTypeFeature

// WorkIssueTypeMappedTypeBug is the enumeration value for bug
const WorkIssueTypeMappedTypeBug WorkIssueTypeMappedType = work.IssueTypeMappedTypeBug

// WorkIssueTypeMappedTypeEnhancement is the enumeration value for enhancement
const WorkIssueTypeMappedTypeEnhancement WorkIssueTypeMappedType = work.IssueTypeMappedTypeEnhancement

// WorkIssueTypeMappedTypeEpic is the enumeration value for epic
const WorkIssueTypeMappedTypeEpic WorkIssueTypeMappedType = work.IssueTypeMappedTypeEpic

// WorkIssueTypeMappedTypeStory is the enumeration value for story
const WorkIssueTypeMappedTypeStory WorkIssueTypeMappedType = work.IssueTypeMappedTypeStory

// WorkIssueTypeMappedTypeTask is the enumeration value for task
const WorkIssueTypeMappedTypeTask WorkIssueTypeMappedType = work.IssueTypeMappedTypeTask

// WorkIssueTypeMappedTypeSubtask is the enumeration value for subtask
const WorkIssueTypeMappedTypeSubtask WorkIssueTypeMappedType = work.IssueTypeMappedTypeSubtask

// WorkConfig is the work config model
type WorkConfig = work.Config

// WorkConfigPartial is the work config model partial
type WorkConfigPartial = work.ConfigPartial

// WorkConfigStatuses is the work config status type
type WorkConfigStatuses = work.ConfigStatuses

// WorkProjectVisibility is the visibility of the project
type WorkProjectVisibility = work.ProjectVisibility

// WorkProjectVisibilityPrivate is the enumeration value for private
const WorkProjectVisibilityPrivate = work.ProjectVisibilityPrivate

// WorkProjectVisibilityPublic is the enumeration value for public
const WorkProjectVisibilityPublic = work.ProjectVisibilityPublic

// WorkProjectAffiliation is the project affiliation
type WorkProjectAffiliation = work.ProjectAffiliation

// WorkProjectAffiliationOrganization is the enumeration value for organization
const WorkProjectAffiliationOrganization = work.ProjectAffiliationOrganization

// WorkProjectAffiliationUser is the enumeration value for user
const WorkProjectAffiliationUser = work.ProjectAffiliationUser

// WorkProjectAffiliationThirdparty is the enumeration value for thirdparty
const WorkProjectAffiliationThirdparty = work.ProjectAffiliationThirdparty

// NewWorkProjectID will return the work project id
func NewWorkProjectID(customerID string, refID string, refType string) string {
	return work.NewProjectID(customerID, refID, refType)
}

// NewWorkIssueID will return the work issue id
func NewWorkIssueID(customerID string, refID string, refType string) string {
	return work.NewIssueID(customerID, refID, refType)
}

// NewWorkIssueCommentID will return the work issue comment id
func NewWorkIssueCommentID(customerID string, refID string, refType string, projectID string) string {
	return work.NewIssueCommentID(customerID, refID, refType, projectID)
}

// NewWorkIssuePriorityID will return the work issue priority id
func NewWorkIssuePriorityID(customerID string, refType string, refID string) string {
	return work.NewIssuePriorityID(customerID, refType, refID)
}

// NewWorkIssueStatusID will return the work issue status id
func NewWorkIssueStatusID(customerID string, refType string, refID string) string {
	return work.NewIssueStatusID(customerID, refType, refID)
}

// NewWorkIssueTypeID will return the work issue type id
func NewWorkIssueTypeID(customerID string, refType string, refID string) string {
	return work.NewIssueTypeID(customerID, refType, refID)
}

// NewWorkKanbanBoardID will return the work kanban board id
func NewWorkKanbanBoardID(customerID string, refID string, refType string) string {
	return work.NewKanbanBoardID(customerID, refType, refID)
}

// NewWorkSprintID will return the work sprint id
func NewWorkSprintID(customerID string, refID string, refType string) string {
	return work.NewSprintID(customerID, refID, refType)
}

// NewWorkUserID will return the work user id
func NewWorkUserID(customerID string, refID string, refType string) string {
	return work.NewUserID(customerID, refID, refType)
}

// NewWorkConfigID will return the work config id
func NewWorkConfigID(customerID string, refType string, integrationInstanceID string) string {
	return work.NewConfigID(customerID, refType, integrationInstanceID)
}
