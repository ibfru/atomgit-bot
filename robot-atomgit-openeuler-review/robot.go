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

	GetSinglePR(pr *atomgitclient.PRIssue) (*atomgit.PullRequest, error)
	GetPullRequests(pr *atomgitclient.PRIssue) ([]*atomgit.PullRequest, error)
	MergePR(pr *atomgitclient.PRIssue, commitMessage string, opt *atomgit.PullRequestOptions) error
	UpdatePR(pr *atomgitclient.PRIssue, request *atomgit.PullRequest) (*atomgit.PullRequest, error)
	AssignPR(pr *atomgitclient.PRIssue, logins []string) error

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

type parameter struct {
	prArg                                                 *atomgitclient.PRIssue
	realPR                                                *atomgit.PullRequest
	bcf                                                   *botConfig
	log                                                   *logrus.Entry
	commentContent, commentator, commitID, action, author string
}

type flowFunc func(arg *parameter) error

func handleFlow(p *parameter, flow []flowFunc) error {
	merr := utils.NewMultiErrors()
	var err error
	for i, j := 0, len(flow); i < j; i++ {
		if err = flow[j](p); err != nil {
			merr.AddError(err)
		}
	}
	return merr.Err()
}

func (bot *robot) handlePullRequest(e *atomgit.PullRequestEvent, pc config.Config, log *logrus.Entry) error {
	org, repo := e.GetRepo().GetOrgAndRepo()
	cfg, err := bot.getConfig(pc, org, repo)
	if err != nil {
		return err
	}

	p := &parameter{
		prArg:  atomgitclient.BuildPRIssue(org, repo, e.GetPullRequest().GetNumber()),
		realPR: e.GetPullRequest(),
		bcf:    cfg,
		log:    log,
		action: e.GetAction(),
		author: e.GetPullRequest().GetUser().GetLogin(),
	}

	return handleFlow(
		p,
		[]flowFunc{
			bot.clearLabel,
			bot.doRetest,
			bot.checkReviewer,
			bot.handleLabelUpdate,
		},
	)
}

func (bot *robot) handlePullRequestReviewComment(e *atomgit.PullRequestReviewCommentEvent, pc config.Config, log *logrus.Entry) error {
	if e.GetAction() != "opened" {
		return nil
	}

	org, repo := e.GetRepo().GetOrgAndRepo()
	cfg, err := bot.getConfig(pc, org, repo)
	if err != nil {
		return err
	}
	p := &parameter{
		prArg:          atomgitclient.BuildPRIssue(org, repo, e.GetPullRequest().GetNumber()),
		realPR:         e.GetPullRequest(),
		bcf:            cfg,
		log:            log,
		commentContent: e.GetComment().GetBody(),
		commentator:    e.GetComment().GetUser().GetLogin(),
		commitID:       e.GetComment().GetCommitID(),
		action:         e.GetAction(),
		author:         e.GetPullRequest().GetUser().GetLogin(),
	}

	return handleFlow(
		p,
		[]flowFunc{
			bot.handleLGTM,
			bot.handleApprove,
			bot.handleCheckPR,
			bot.removeInvalidCLA,
			bot.handleRebase,
			bot.handleFlattened,
			bot.removeRebase,
			bot.removeFlattened,
			bot.handleACK,
		},
	)
}
