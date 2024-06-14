package framework

import (
	_ "github.com/opensourceways/community-robot-lib/atomgitclient"
	"github.com/opensourceways/community-robot-lib/config"
	"github.com/opensourceways/go-atomgit/atomgit"
	"github.com/sirupsen/logrus"
)

// AccessHandler defines the function contract for a github.IssuesEvent handler.
type AccessHandler func(e *atomgit.AccessEvent, cfg config.Config, log *logrus.Entry, payload []byte) error

// IssueHandler defines the function contract for a github.IssuesEvent handler.
type IssueHandler func(e *atomgit.IssuesEvent, cfg config.Config, log *logrus.Entry) error

// IssueCommentHandler defines the function contract for a github.IssueCommentEvent handler.
type IssueCommentHandler func(e *atomgit.IssueCommentEvent, cfg config.Config, log *logrus.Entry) error

// PullRequestHandler defines the function contract for a github.PullRequestEvent handler.
type PullRequestHandler func(e *atomgit.PullRequestEvent, cfg config.Config, log *logrus.Entry) error

// PushEventHandler defines the function contract for a github.PushEvent handler.
type PushEventHandler func(e *atomgit.PushEvent, cfg config.Config, log *logrus.Entry) error

// ReviewEventHandler defines the function contract for a github.PullRequestReviewEvent handler.
type ReviewEventHandler func(e *atomgit.PullRequestReviewEvent, cfg config.Config, log *logrus.Entry) error

// ReviewCommentEventHandler defines the function contract for a github.PullRequestReviewCommentEvent handler.
type ReviewCommentEventHandler func(e *atomgit.PullRequestReviewCommentEvent, cfg config.Config, log *logrus.Entry) error

type handlers struct {
	accessHandlers            AccessHandler
	issueHandlers             IssueHandler
	pullRequestHandler        PullRequestHandler
	pushEventHandler          PushEventHandler
	issueCommentHandler       IssueCommentHandler
	reviewEventHandler        ReviewEventHandler
	reviewCommentEventHandler ReviewCommentEventHandler
}

// RegisterAccessHandler registers a plugin's github.IssueEvent handler.
func (h *handlers) RegisterAccessHandler(fn AccessHandler) {
	h.accessHandlers = fn
}

// RegisterIssueHandler registers a plugin's github.IssueEvent handler.
func (h *handlers) RegisterIssueHandler(fn IssueHandler) {
	h.issueHandlers = fn
}

// RegisterPullRequestHandler registers a plugin's github.PullRequestEvent handler.
func (h *handlers) RegisterPullRequestHandler(fn PullRequestHandler) {
	h.pullRequestHandler = fn
}

// RegisterPushEventHandler registers a plugin's github.PushEvent handler.
func (h *handlers) RegisterPushEventHandler(fn PushEventHandler) {
	h.pushEventHandler = fn
}

// RegisterIssueCommentHandler registers a plugin's github.IssueCommentEvent handler.
func (h *handlers) RegisterIssueCommentHandler(fn IssueCommentHandler) {
	h.issueCommentHandler = fn
}

// RegisterReviewEventHandler registers a plugin's github.ReviewEvent handler.
func (h *handlers) RegisterReviewEventHandler(fn ReviewEventHandler) {
	h.reviewEventHandler = fn
}

// RegisterReviewCommentEventHandler registers a plugin's github.ReviewCommentEvent handler.
func (h *handlers) RegisterReviewCommentEventHandler(fn ReviewCommentEventHandler) {
	h.reviewCommentEventHandler = fn
}
