package main

import (
	"fmt"

	"github.com/opensourceways/community-robot-lib/utils"

	"github.com/opensourceways/community-robot-lib/atomgitclient"
	"github.com/opensourceways/go-atomgit/atomgit"
	"github.com/sirupsen/logrus"

	"github.com/opensourceways/community-robot-lib/config"
	framework "github.com/opensourceways/community-robot-lib/robot-atomgit-framework"
)

const botName = "label"

type iClient interface {
	GetRepositoryLabels(pr *atomgitclient.PRIssue) ([]string, error)
	CreateRepoLabel(org, repo, label string) error

	GetPRLabels(pr *atomgitclient.PRIssue) ([]string, error)
	AddPRLabel(pr *atomgitclient.PRIssue, label string) error
	RemovePRLabel(pr *atomgitclient.PRIssue, label string) error

	GetSinglePR(pr *atomgitclient.PRIssue) (*atomgit.PullRequest, error)
	GetPullRequests(pr *atomgitclient.PRIssue) ([]*atomgit.PullRequest, error)
	MergePR(pr *atomgitclient.PRIssue, commitMessage string, opt *atomgit.PullRequestOptions) error

	GetIssueLabels(is *atomgitclient.PRIssue) ([]string, error)
	AddIssueLabel(is *atomgitclient.PRIssue, label []string) error
	RemoveIssueLabel(is *atomgitclient.PRIssue, label string) error

	IsCollaborator(pr *atomgitclient.PRIssue, login string) (bool, error)

	GetDirectoryTree(org, repo, branch string, recursive bool) ([]*atomgit.TreeEntry, error)
	GetPullRequestChanges(pr *atomgitclient.PRIssue) ([]*atomgit.CommitFile, error)
	GetPathContent(org, repo, path, branch string) (*atomgit.RepositoryContent, error)

	ListOperationLogs(pr *atomgitclient.PRIssue) ([]*atomgit.Timeline, error)
	ListIssueComments(is *atomgitclient.PRIssue) ([]*atomgit.IssueComment, error)
	CreatePRComment(pr *atomgitclient.PRIssue, comment string) error
	CreateIssueComment(is *atomgitclient.PRIssue, comment string) error
	CreatePRCommentReply(pr *atomgitclient.PRIssue, comment, commentID string) error

	GetUserPermissionOfRepo(org, repo, user string) (*atomgit.RepositoryPermissionLevel, error)
}

func newRobot(cli iClient) *robot {
	return &robot{cli: cli}
}

type robot struct {
	cli iClient
}

func (bot *robot) NewConfig() config.Config {
	return &configuration{}
}

func (bot *robot) getConfig(cfg config.Config, org, repo string) (*botConfig, error) {
	c, ok := cfg.(*configuration)
	if !ok {
		return nil, fmt.Errorf("can't convert to configuration")
	}

	if bc := c.configFor(org, repo); bc != nil {
		return bc, nil
	}

	return nil, fmt.Errorf("no config for this repo:%s/%s", org, repo)
}

func (bot *robot) RegisterEventHandler(f framework.HandlerRegister) {
	f.RegisterIssueCommentHandler(bot.handleIssueComment)
	f.RegisterReviewCommentEventHandler(bot.handlePullRequestReviewComment)
	f.RegisterPullRequestHandler(bot.handlePullRequest)
}

// TODO atomgit 上 Issue comment event 不触发 webhook
func (bot *robot) handleIssueComment(e *atomgit.IssueCommentEvent, cfg config.Config, log *logrus.Entry) error {

	org, repo := e.GetRepo().GetOrgAndRepo()
	bc, err := bot.getConfig(cfg, org, repo)
	if err != nil {
		return err
	}

	toAdd, toRemove := getMatchedLabels(e.GetComment().GetBody())
	if len(toAdd) == 0 && len(toRemove) == 0 {
		log.Debug("invalid comment, skipping.")
		return nil
	}

	lh := &labelHelper{
		cli:         bot.cli,
		flag:        PullRequest,
		prIssue:     atomgitclient.BuildPRIssue(org, repo, e.GetIssue().GetNumber()),
		labels:      e.GetIssue().GetLabels(),
		commentator: e.GetComment().GetUser().GetLogin(),
		commitID:    e.GetComment().GetNodeID(), // TODO
		add:         toAdd,
		remove:      toRemove,
	}
	return bot.handleLabelsByComment(lh, bc, log)
}

func (bot *robot) handlePullRequestReviewComment(e *atomgit.PullRequestReviewCommentEvent, cfg config.Config, log *logrus.Entry) error {

	org, repo := e.GetRepo().GetOrgAndRepo()
	bc, err := bot.getConfig(cfg, org, repo)
	if err != nil {
		return err
	}

	toAdd, toRemove := getMatchedLabels(e.GetComment().GetBody())
	if len(toAdd) == 0 && len(toRemove) == 0 {
		log.Debug("invalid comment, skipping.")
		return nil
	}

	lh := &labelHelper{
		cli:         bot.cli,
		flag:        PullRequest,
		prIssue:     atomgitclient.BuildPRIssue(org, repo, e.GetPullRequest().GetNumber()),
		labels:      e.GetPullRequest().GetLabels(),
		commentator: e.GetComment().GetUser().GetLogin(),
		commitID:    e.GetComment().GetCommitID(),
		add:         toAdd,
		remove:      toRemove,
	}
	return bot.handleLabelsByComment(lh, bc, log)
}

// TODO atomgit 上 PR code update event 不触发 webhook
func (bot *robot) handlePullRequest(e *atomgit.PullRequestEvent, cfg config.Config, log *logrus.Entry) error {

	org, repo := e.GetRepo().GetOrgAndRepo()
	bc, err := bot.getConfig(cfg, org, repo)
	if err != nil {
		return err
	}

	lh := &labelHelper{
		cli:     bot.cli,
		flag:    PullRequest,
		prIssue: atomgitclient.BuildPRIssue(org, repo, e.GetPullRequest().GetNumber()),
		labels:  e.GetPullRequest().GetLabels(),
	}

	errs := utils.NewMultiErrors()
	if err = bot.clearLabelCaseByPRCodeUpdate(lh, bc); err != nil {
		errs.AddError(err)
	}

	if e.GetAction() == atomgit.ActionStateSynchronized {
		if err = bot.handleSquashLabel(lh, uint(e.GetPullRequest().GetCommits()), bc.SquashConfig); err != nil {
			errs.AddError(err)
		}
	}

	return errs.Err()
}

func (bot *robot) handleSquashLabel(lh *labelHelper, commits uint, cfg SquashConfig) error {
	if cfg.unableCheckingSquash() {
		return nil
	}

	labels := lh.getCurrentLabels()
	hasSquashLabel := labels.Has(cfg.SquashCommitLabel)
	exceeded := commits > cfg.CommitsThreshold

	if exceeded && !hasSquashLabel {
		return bot.cli.AddPRLabel(lh.prIssue, cfg.SquashCommitLabel)
	}

	if !exceeded && hasSquashLabel {
		return bot.cli.RemovePRLabel(lh.prIssue, cfg.SquashCommitLabel)
	}

	return nil
}
