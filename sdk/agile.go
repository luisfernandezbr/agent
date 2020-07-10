package sdk

import "github.com/pinpt/integration-sdk/work"

// Type aliases for our exported datamodel types to create a stable version
// which Integrations depend on instead of directly depending on a specific
// version of the integration-sdk directly

// AgileBoard is the agile board
type AgileBoard = work.Board

// AgileBoardType is the board type
type AgileBoardType = work.BoardType

// AgileBoardTypeScrum is the enumeration value for scrum
const AgileBoardTypeScrum = work.BoardTypeScrum

// AgileBoardTypeKanban is the enumeration value for kanban
const AgileBoardTypeKanban = work.BoardTypeKanban

// AgileBoardTypeOther is the enumeration value for other
const AgileBoardTypeOther = work.BoardTypeOther

// AgileKanban is a kanban board
type AgileKanban = work.Kanban

// AgileSprint is a sprint board
type AgileSprint = work.Sprint

// AgileBoardColumns is the columns for an agile board
type AgileBoardColumns = work.BoardColumns

// AgileKanbanColumns is the columns for a kanban board
type AgileKanbanColumns = work.KanbanColumns

// AgileSprintClosedDate is the sprint closed date
type AgileSprintClosedDate = work.SprintCompletedDate

// AgileSprintEndDate is the sprint planned end date
type AgileSprintEndDate = work.SprintEndedDate

// AgileSprintStartDate is the sprint planned start date
type AgileSprintStartDate = work.SprintStartedDate

// AgileSprintColumns is the sprint columns
type AgileSprintColumns = work.SprintColumns

// AgileSprintStatus is the sprint status
type AgileSprintStatus = work.SprintStatus

// AgileSprintStatusClosed is the enumeration value for closed
const AgileSprintStatusClosed = work.SprintStatusClosed

// AgileSprintStatusActive is the enumeration value for active
const AgileSprintStatusActive = work.SprintStatusActive

// AgileSprintStatusFuture is the enumeration value for future
const AgileSprintStatusFuture = work.SprintStatusFuture

// NewAgileBoardID will return the agile board id
func NewAgileBoardID(customerID string, refID string, refType string) string {
	return work.NewBoardID(customerID, refID, refType)
}

// NewAgileKanbanID will return the agile kanban id
func NewAgileKanbanID(customerID string, refID string, refType string) string {
	return work.NewKanbanID(customerID, refID, refType)
}

// NewAgileSprintID will return the agile sprint id
func NewAgileSprintID(customerID string, refID string, refType string) string {
	return work.NewSprintID(customerID, refID, refType)
}
