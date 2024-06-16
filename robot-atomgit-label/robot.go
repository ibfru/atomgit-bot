package main

import (
	"fmt"
	"net/http"
	"time"

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
	f.RegisterReviewCommentEventHandler(bot.handleReviewComment)
	f.RegisterPullRequestHandler(bot.handlePullRequest)
}

func (bot *robot) handleIssueComment(e *atomgit.IssueCommentEvent, cfg config.Config, log *logrus.Entry) error {
	// TODO atomgit 上 Issue comment event 不触发 webhook

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

	lh := &issueLabelHelper{
		number:    e.GetIssue().GetNumber(),
		labels:    e.GetIssue().GetLabels(),
		commenter: e.GetComment().GetUser().GetLogin(),
		commitID:  e.GetComment().GetCommitID(),
		repoLabelHelper: &repoLabelHelper{
			cli:    &bot.cli,
			org:    org,
			repo:   repo,
			add:    toAdd,
			remove: toRemove,
		},
	}

	return nil
}

type Message struct {
	CommitId string `json:"commit_id" required:"true"`
	Body     string `json:"body" required:"true"`
}

func (bot *robot) handleReviewComment(e *atomgit.PullRequestReviewCommentEvent, cfg config.Config, log *logrus.Entry) error {
	//m := Message{e.GetComment().GetCommitID(), "fasjkhghfvdaolkuhsivfdaljhsdvaljhsdvlakjhsdvasd"}
	//b, err := json.Marshal(m)
	//
	//req, err := http.NewRequest(http.MethodPost, "https://api.atomgit.com/repos/ibfuorg/fengchaopub/pulls/1/comments", bytes.NewBuffer(b))
	//if err != nil {
	//	fmt.Println("asdsadasdasdasd")
	//}
	//
	//req.Header.Set("Content-Type", "application/json")
	//req.Header.Set("Authorization", "Bearer atp_0no70zw700cag7v6kc0cm1vvi8lp3qsk")
	//res, err := do(req)
	//if err != nil {
	//	return err
	//}
	//fmt.Printf("%+v", res)

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

func (bot *robot) handlePullRequest(e *atomgit.PullRequestEvent, cfg config.Config, log *logrus.Entry) error {
	// TODO
	fmt.Println("-------------")

	return nil
}

func do(req *http.Request) (resp *http.Response, err error) {
	var tmp = http.DefaultClient
	if resp, err = tmp.Do(req); err == nil {
		return
	}

	maxRetries := 4
	backoff := 100 * time.Millisecond

	for retries := 0; retries < maxRetries; retries++ {
		time.Sleep(backoff)
		backoff *= 2

		if resp, err = tmp.Do(req); err == nil {
			break
		}
	}
	return
}
