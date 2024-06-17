package main

import (
	"fmt"

	"github.com/opensourceways/go-atomgit/atomgit"

	cache "github.com/opensourceways/atomgit-sig-file-cache/sdk"
	"github.com/opensourceways/community-robot-lib/atomgitclient"
	"github.com/opensourceways/community-robot-lib/config"
	framework "github.com/opensourceways/community-robot-lib/robot-atomgit-framework"
	"github.com/opensourceways/community-robot-lib/utils"
	"github.com/sirupsen/logrus"
)

const botName = "review"

type iClient interface {
	GetRepositoryLabels(pr *atomgitclient.PRIssue) ([]string, error)
	CreateRepoLabel(org, repo, label string) error

	GetPRLabels(pr *atomgitclient.PRIssue) ([]string, error)
	AddPRLabel(pr *atomgitclient.PRIssue, label string) error
	RemovePRLabel(pr *atomgitclient.PRIssue, label string) error

	GetUserPermissionOfRepo(org, repo, user string) (*atomgit.RepositoryPermissionLevel, error)
	GetPathContent(org, repo, path, branch string) (*atomgit.RepositoryContent, error)
	GetPullRequestChanges(pr *atomgitclient.PRIssue) ([]*atomgit.CommitFile, error)

	GetPullRequests(pr *atomgitclient.PRIssue) ([]*atomgit.PullRequest, error)
	MergePR(pr *atomgitclient.PRIssue, commitMessage string, opt *atomgit.PullRequestOptions) error
	UpdatePR(pr *atomgitclient.PRIssue, request *atomgit.PullRequest) (*atomgit.PullRequest, error)

	GetPRComments(pr *atomgitclient.PRIssue) ([]*atomgit.PullRequestComment, error)
	CreatePRComment(pr *atomgitclient.PRIssue, comment string) error

	ListOperationLogs(pr *atomgitclient.PRIssue) ([]*atomgit.Timeline, error)
}

func newRobot(cli iClient, cacheCli *cache.SDK) *robot {
	return &robot{cli: cli, cacheCli: cacheCli}
}

type robot struct {
	cli      iClient
	cacheCli *cache.SDK
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
	f.RegisterReviewCommentEventHandler(bot.handlePullRequestReviewComment)
	f.RegisterPullRequestHandler(bot.handlePullRequest)
}

func (bot *robot) handlePREvent(e *atomgit.PullRequestEvent, pc config.Config, log *logrus.Entry) error {
	org, repo := e.GetOrgRepo()
	cfg, err := bot.getConfig(pc, org, repo)
	if err != nil {
		return err
	}

	merr := utils.NewMultiErrors()
	if err := bot.clearLabel(e); err != nil {
		merr.AddError(err)
	}

	if err := bot.doRetest(e); err != nil {
		merr.AddError(err)
	}

	if err := bot.checkReviewer(e, cfg); err != nil {
		merr.AddError(err)
	}

	if err := bot.handleLabelUpdate(e, cfg, log); err != nil {
		merr.AddError(err)
	}

	return merr.Err()
}

func (bot *robot) handleNoteEvent(e *atomgit.NoteEvent, pc config.Config, log *logrus.Entry) error {
	org, repo := e.GetOrgRepo()
	cfg, err := bot.getConfig(pc, org, repo)
	if err != nil {
		return err
	}

	merr := utils.NewMultiErrors()
	if err := bot.handleLGTM(e, cfg, log); err != nil {
		merr.AddError(err)
	}

	if err = bot.handleApprove(e, cfg, log); err != nil {
		merr.AddError(err)
	}

	if err = bot.handleCheckPR(e, cfg, log); err != nil {
		merr.AddError(err)
	}

	if err = bot.removeInvalidCLA(e, cfg, log); err != nil {
		merr.AddError(err)
	}

	if err = bot.handleRebase(e, cfg, log); err != nil {
		merr.AddError(err)
	}

	if err = bot.handleFlattened(e, cfg, log); err != nil {
		merr.AddError(err)
	}

	if err = bot.removeRebase(e, cfg, log); err != nil {
		merr.AddError(err)
	}

	if err = bot.removeFlattened(e, cfg, log); err != nil {
		merr.AddError(err)
	}

	if err = bot.handleACK(e, cfg, log); err != nil {
		merr.AddError(err)
	}

	return merr.Err()
}
