package sdk

import "github.com/pinpt/integration-sdk/sourcecode"

// Type aliases for our exported datamodel types to create a stable version
// which Integrations depend on instead of directly depending on a specific
// version of the integration-sdk directly

// SourceCodePullRequest is a pull request
type SourceCodePullRequest = sourcecode.PullRequest

// SourceCodePullRequestClosedDate is the pull request closed date
type SourceCodePullRequestClosedDate = sourcecode.PullRequestClosedDate

// SourceCodePullRequestCreatedDate is the pull request created date
type SourceCodePullRequestCreatedDate = sourcecode.PullRequestCreatedDate

// SourceCodePullRequestMergedDate is the pull request merged date
type SourceCodePullRequestMergedDate = sourcecode.PullRequestMergedDate

// SourceCodePullRequestUpdatedDate is the pull request updated date
type SourceCodePullRequestUpdatedDate = sourcecode.PullRequestUpdatedDate

// SourceCodePullRequestCommit is a pull request commit
type SourceCodePullRequestCommit = sourcecode.PullRequestCommit

// SourceCodePullRequestCommitCreatedDate is the created date for a pull request commit
type SourceCodePullRequestCommitCreatedDate = sourcecode.PullRequestCommitCreatedDate

// SourceCodePullRequestComment is a pull request comment
type SourceCodePullRequestComment = sourcecode.PullRequestComment

// SourceCodePullRequestCommentCreatedDate is a pull request comment created date
type SourceCodePullRequestCommentCreatedDate = sourcecode.PullRequestCommentCreatedDate

// SourceCodePullRequestCommentUpdatedDate is a pull request comment updated date
type SourceCodePullRequestCommentUpdatedDate = sourcecode.PullRequestCommentUpdatedDate

// SourceCodePullRequestReview is a pull request review
type SourceCodePullRequestReview = sourcecode.PullRequestReview

// SourceCodePullRequestReviewCreatedDate is a pull request review created date
type SourceCodePullRequestReviewCreatedDate = sourcecode.PullRequestReviewCreatedDate

// SourceCodeRepo is a code repository in the source code system
type SourceCodeRepo = sourcecode.Repo

// SourceCodeUser is a user in the source code system
type SourceCodeUser = sourcecode.User

// SourceCodeUserType is a user type in the source code system
type SourceCodeUserType = sourcecode.UserType

// SourceCodeUserTypeHuman is a HUMAN user type in the source code system for describing a user
const SourceCodeUserTypeHuman = sourcecode.UserTypeHuman

// SourceCodeUserTypeBot is a BOT user type in the source code system for describing an automated user account
const SourceCodeUserTypeBot = sourcecode.UserTypeBot

// SourceCodeUserTypeDeletedSpecialUser is a Special Deleted user type for describing a deleted user account
const SourceCodeUserTypeDeletedSpecialUser = sourcecode.UserTypeDeletedSpecialUser

// SourceCodePullRequestReviewState is a pull request review state enum
type SourceCodePullRequestReviewState = sourcecode.PullRequestReviewState

// SourceCodePullRequestReviewStateApproved is the enumeration value for approved
const SourceCodePullRequestReviewStateApproved SourceCodePullRequestReviewState = sourcecode.PullRequestReviewStateApproved

// SourceCodePullRequestReviewStateCommented is the enumeration value for commented
const SourceCodePullRequestReviewStateCommented SourceCodePullRequestReviewState = sourcecode.PullRequestReviewStateCommented

// SourceCodePullRequestReviewStateChangesRequested is the enumeration value for changes_requested
const SourceCodePullRequestReviewStateChangesRequested SourceCodePullRequestReviewState = sourcecode.PullRequestReviewStateChangesRequested

// SourceCodePullRequestReviewStatePending is the enumeration value for pending
const SourceCodePullRequestReviewStatePending SourceCodePullRequestReviewState = sourcecode.PullRequestReviewStatePending

// SourceCodePullRequestReviewStateDismissed is the enumeration value for dismissed
const SourceCodePullRequestReviewStateDismissed SourceCodePullRequestReviewState = sourcecode.PullRequestReviewStateDismissed

// SourceCodePullRequestReviewStateRequested is the enumeration value for requested
const SourceCodePullRequestReviewStateRequested SourceCodePullRequestReviewState = sourcecode.PullRequestReviewStateRequested

// SourceCodePullRequestReviewStateRequestRemoved is the enumeration value for request_removed
const SourceCodePullRequestReviewStateRequestRemoved SourceCodePullRequestReviewState = sourcecode.PullRequestReviewStateRequestRemoved

// SourceCodePullRequestReviewStateAssigned is the enumeration value for assigned
const SourceCodePullRequestReviewStateAssigned SourceCodePullRequestReviewState = sourcecode.PullRequestReviewStateAssigned

// SourceCodePullRequestReviewStateUnassigned is the enumeration value for unassigned
const SourceCodePullRequestReviewStateUnassigned SourceCodePullRequestReviewState = sourcecode.PullRequestReviewStateUnassigned

// SourceCodePullRequestStatus is the pull request state enum
type SourceCodePullRequestStatus = sourcecode.PullRequestStatus

// SourceCodePullRequestStatusOpen is the enumeration value for open
const SourceCodePullRequestStatusOpen SourceCodePullRequestStatus = sourcecode.PullRequestStatusOpen

// SourceCodePullRequestStatusClosed is the enumeration value for closed
const SourceCodePullRequestStatusClosed SourceCodePullRequestStatus = sourcecode.PullRequestStatusClosed

// SourceCodePullRequestStatusMerged is the enumeration value for merged
const SourceCodePullRequestStatusMerged SourceCodePullRequestStatus = sourcecode.PullRequestStatusMerged

// SourceCodePullRequestStatusSuperseded is the enumeration value for superseded
const SourceCodePullRequestStatusSuperseded SourceCodePullRequestStatus = sourcecode.PullRequestStatusSuperseded

// SourceCodePullRequestStatusLocked is the enumeration value for locked
const SourceCodePullRequestStatusLocked SourceCodePullRequestStatus = sourcecode.PullRequestStatusLocked

// NewSourceCodeUserID returns a new sourcecode user id
func NewSourceCodeUserID(customerID string, refType string, refID string) string {
	return sourcecode.NewUserID(customerID, refType, refID)
}

// NewSourceCodeCommitID returns a new sourcecode commit id
func NewSourceCodeCommitID(customerID string, sha string, refType string, repoID string) string {
	return sourcecode.NewCommitID(customerID, sha, refType, repoID)
}

// NewSourceCodePullRequestCommentID returns a new sourcecode pull request comment id
func NewSourceCodePullRequestCommentID(customerID string, refID string, refType string, repoID string) string {
	return sourcecode.NewPullRequestCommentID(customerID, refID, refType, repoID)
}

// NewSourceCodePullRequestReviewID returns a new sourcecode pull request review id
func NewSourceCodePullRequestReviewID(customerID string, refID string, refType string, repoID string) string {
	return sourcecode.NewPullRequestReviewID(customerID, refID, refType, repoID)
}

// NewSourceCodeBranchID returns a new sourcecode branch id
func NewSourceCodeBranchID(customerID string, repoID string, refType string, branchName string, firstCommitID string) string {
	return sourcecode.NewBranchID(refType, repoID, customerID, branchName, firstCommitID)
}

// NewSourceCodePullRequestID returns a new sourcecode pullrequest id
func NewSourceCodePullRequestID(customerID string, refID string, refType string, repoID string) string {
	return sourcecode.NewPullRequestID(customerID, refID, refType, repoID)
}

// NewSourceCodeRepoID returns the new sourcecode repo id
func NewSourceCodeRepoID(customerID string, refID string, refType string) string {
	return sourcecode.NewRepoID(customerID, refType, refID)
}
