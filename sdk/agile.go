package sdk

import "github.com/pinpt/integration-sdk/work/agile"

// Type aliases for our exported datamodel types to create a stable version
// which Integrations depend on instead of directly depending on a specific
// version of the integration-sdk directly

// AgileBoard is the agile board
type AgileBoard = agile.Board

// AgileBoardType is the board type
type AgileBoardType = agile.BoardType

// AgileBoardTypeScrum is the enumeration value for scrum
const AgileBoardTypeScrum = agile.BoardTypeScrum

// AgileBoardTypeKanban is the enumeration value for kanban
const AgileBoardTypeKanban = agile.BoardTypeKanban

// AgileBoardTypeOther is the enumeration value for other
const AgileBoardTypeOther = agile.BoardTypeOther

// AgileKanban is a kanban board
type AgileKanban = agile.Kanban

// AgileSprint is a sprint board
type AgileSprint = agile.Sprint

// AgileBoardColumns is the columns for an agile board
type AgileBoardColumns = agile.BoardColumns

// AgileKanbanColumns is the columns for a kanban board
type AgileKanbanColumns = agile.KanbanColumns

// AgileSprintClosedDate is the sprint closed date
type AgileSprintClosedDate = agile.SprintClosedDate

// AgileSprintEndDate is the sprint planned end date
type AgileSprintEndDate = agile.SprintEndDate

// AgileSprintStartDate is the sprint planned start date
type AgileSprintStartDate = agile.SprintStartDate

// AgileSprintColumns is the sprint columns
type AgileSprintColumns = agile.SprintColumns

// AgileSprintStatus is the sprint status
type AgileSprintStatus = agile.SprintStatus

// AgileSprintStatusClosed is the enumeration value for closed
const AgileSprintStatusClosed = agile.SprintStatusClosed

// AgileSprintStatusActive is the enumeration value for active
const AgileSprintStatusActive = agile.SprintStatusActive

// AgileSprintStatusFuture is the enumeration value for future
const AgileSprintStatusFuture = agile.SprintStatusFuture

// NewAgileBoardID will return the agile board id
func NewAgileBoardID(customerID string, refID string, refType string) string {
	return agile.NewBoardID(customerID, refID, refType)
}

// NewAgileKanbanID will return the agile kanban id
func NewAgileKanbanID(customerID string, refID string, refType string) string {
	return agile.NewKanbanID(customerID, refID, refType)
}

// NewAgileSprintID will return the agile sprint id
func NewAgileSprintID(customerID string, refID string, refType string) string {
	return agile.NewSprintID(customerID, refID, refType)
}
