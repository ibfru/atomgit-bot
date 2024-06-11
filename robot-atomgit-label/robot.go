package main

import (
	"fmt"
	"net/http"
	"time"

	_ "github.com/opensourceways/community-robot-lib/atomgitclient"
	sdk "github.com/opensourceways/go-atomgit/atomgit"
	"github.com/sirupsen/logrus"

	"github.com/opensourceways/community-robot-lib/config"
	framework "github.com/opensourceways/community-robot-lib/robot-atomgit-framework"
)

const botName = "label"

type iClient interface {
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

func (bot *robot) handleIssueComment(e *sdk.IssueCommentEvent, cfg config.Config, log *logrus.Entry) error {
	// TODO
	fmt.Println("-------------")

	return nil
}

type Message struct {
	CommitId string `json:"commit_id" required:"true"`
	Body     string `json:"body" required:"true"`
}

func (bot *robot) handleReviewComment(e *sdk.PullRequestReviewCommentEvent, cfg config.Config, log *logrus.Entry) error {
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
	// TODO
	fmt.Println("-------------")
	return nil
}

func (bot *robot) handlePullRequest(e *sdk.PullRequestEvent, cfg config.Config, log *logrus.Entry) error {
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
