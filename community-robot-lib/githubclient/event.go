package githubclient

import (
	sdk "github.com/google/go-github/v36/github"
	"k8s.io/apimachinery/pkg/util/sets"
)

type IssuePRInfo interface {
	GetOrgRepo() (string, string)

	// GetNumber will return PR or Issue Number
	GetNumber() int

	// GetLabels will return labels on PR or Issue
	GetLabels() sets.String

	// GetAuthor will return author of PR or Issue
	GetAuthor() string
}

func GenIssuePRInfo(e interface{}) IssuePRInfo {
	switch e := e.(type) {
	case *sdk.PullRequestEvent:
		return pullRequestEvent{e}

	case *sdk.IssueCommentEvent:
		return pullRequestCommentEvent{e}

	default:
		return nil
	}
}

type pullRequestEvent struct {
	e *sdk.PullRequestEvent
}

func (e pullRequestEvent) GetOrgRepo() (string, string) {
	return GetOrgRepo(e.e.GetRepo())
}

func (e pullRequestEvent) GetNumber() int {
	return e.e.GetNumber()
}

func (e pullRequestEvent) GetLabels() sets.String {
	pr := e.e.GetPullRequest()
	labels := sets.NewString()
	for _, item := range pr.Labels {
		labels.Insert(item.GetName())
	}

	return labels
}

func (e pullRequestEvent) GetAuthor() string {
	return e.e.GetPullRequest().GetUser().GetLogin()
}

type pullRequestCommentEvent struct {
	e *sdk.IssueCommentEvent
}

func (e pullRequestCommentEvent) GetOrgRepo() (string, string) {
	return GetOrgRepo(e.e.GetRepo())
}

func (e pullRequestCommentEvent) GetNumber() int {
	return e.e.GetIssue().GetNumber()
}

func (e pullRequestCommentEvent) GetLabels() sets.String {
	pr := e.e.GetIssue()
	labels := sets.NewString()
	for _, item := range pr.Labels {
		labels.Insert(item.GetName())
	}

	return labels
}

func (e pullRequestCommentEvent) GetAuthor() string {
	return e.e.GetIssue().GetUser().GetLogin()
}
