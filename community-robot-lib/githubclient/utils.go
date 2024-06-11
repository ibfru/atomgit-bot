package githubclient

import "github.com/google/go-github/v36/github"

const (
	ActionOpened  = "opened"
	ActionCreated = "created"
	ActionReopen  = "reopened"
	ActionClosed  = "closed"

	PRActionOpened              = "opened"
	PRActionChangedSourceBranch = "synchronize"
)

// GetOrgRepo return the owner and name of the repository
func GetOrgRepo(repo *github.Repository) (string, string) {
	return repo.GetOwner().GetLogin(), repo.GetName()
}

// IsIssueOpened tells whether the issue is opened
func IsIssueOpened(action string) bool {
	return action == ActionOpened
}

// IsPROpened tells whether the PR is opened
func IsPROpened(action string) bool {
	return action == ActionOpened
}

// IsPRSourceBranchChanged tells whether the PR's source branch is changed
func IsPRSourceBranchChanged(action string) bool {
	return action == PRActionChangedSourceBranch
}

// IsCommentCreated tells whether the comment is created now.
func IsCommentCreated(e *github.IssueCommentEvent) bool {
	return e.GetAction() == ActionCreated
}

// IsCommentOnPullRequest tells whether the comment is on pull request
func IsCommentOnPullRequest(e *github.IssueCommentEvent) bool {
	return e.GetIssue().IsPullRequest()
}
