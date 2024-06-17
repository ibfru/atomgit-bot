package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/opensourceways/community-robot-lib/atomgitclient"

	"github.com/opensourceways/community-robot-lib/config"
	framework "github.com/opensourceways/community-robot-lib/robot-atomgit-framework"
	"github.com/opensourceways/community-robot-lib/utils"
	"github.com/opensourceways/go-atomgit/atomgit"
	"github.com/sirupsen/logrus"
)

const (
	botName        = "cla"
	maxLengthOfSHA = 8
)

var checkCLARe = regexp.MustCompile(`(?mi)^/check-cla\s*$`)

type iClient interface {
	AddPRLabel(pr *atomgitclient.PRIssue, label string) error
	RemovePRLabel(pr *atomgitclient.PRIssue, label string) error

	CreatePRComment(pr *atomgitclient.PRIssue, comment string) error
	DeletePRComment(org, repo, commentId string) error

	GetPRCommits(pr *atomgitclient.PRIssue) ([]*atomgit.RepositoryCommit, error)
	GetPRComments(pr *atomgitclient.PRIssue) ([]*atomgit.PullRequestComment, error)
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
	f.RegisterPullRequestHandler(bot.handlePullRequest)
	f.RegisterReviewCommentEventHandler(bot.handlePullRequestReviewComment)
}

func (bot *robot) handlePullRequest(e *atomgit.PullRequestEvent, c config.Config, log *logrus.Entry) error {
	if e.GetPullRequest().GetState() != "opened" && e.GetPullRequest().GetState() != "open" {
		return nil
	}

	// action in [opened, source_branch_changed]
	//if v := atomgit.GetPullRequestAction(e); v != atomgit.PRActionOpened && v != atomgit.PRActionChangedSourceBranch {
	//	return nil
	//}

	org, repo := e.GetRepo().GetOrgAndRepo()
	cfg, err := bot.getConfig(c, org, repo)
	if err != nil {
		return err
	}

	//time.Sleep(3 * time.Second)

	return bot.handle(org, repo, e.GetPullRequest(), cfg, false, log)
}

func (bot *robot) handlePullRequestReviewComment(e *atomgit.PullRequestReviewCommentEvent, c config.Config, log *logrus.Entry) error {

	// Only consider "/check-cla" comments.
	if !checkCLARe.MatchString(e.GetComment().GetBody()) {
		return nil
	}

	org, repo := e.GetRepo().GetOrgAndRepo()
	cfg, err := bot.getConfig(c, org, repo)
	if err != nil {
		return err
	}

	return bot.handle(org, repo, e.GetPullRequest(), cfg, true, log)
}

func (bot *robot) handle(
	org, repo string,
	pr *atomgit.PullRequest,
	cfg *botConfig,
	notifyAuthorIfSigned bool,
	log *logrus.Entry,
) error {
	p := atomgitclient.BuildPRIssue(org, repo, pr.GetNumber())

	unsigned, err := bot.getPRCommitsAbout(p, cfg)
	if err != nil {
		return err
	}

	labels := sets.NewString()
	for _, lb := range pr.GetLabels() {
		labels.Insert(*lb.Name)
	}
	hasCLAYes := labels.Has(cfg.CLALabelYes)
	hasCLANo := labels.Has(cfg.CLALabelNo)

	deleteSignGuide(p, bot.cli, cfg)

	if len(unsigned) == 0 {
		if hasCLANo {
			if err := bot.cli.RemovePRLabel(p, cfg.CLALabelNo); err != nil {
				log.WithError(err).Warningf("Could not remove %s label.", cfg.CLALabelNo)
			}
		}

		if !hasCLAYes {
			if err := bot.cli.AddPRLabel(p, cfg.CLALabelYes); err != nil {
				log.WithError(err).Warningf("Could not add %s label.", cfg.CLALabelYes)
			}

			if notifyAuthorIfSigned {
				return bot.cli.CreatePRComment(p, alreadySigned(pr.GetUser().GetLogin()))
			}
		}

		return nil
	}

	if hasCLAYes {
		if err := bot.cli.RemovePRLabel(p, cfg.CLALabelYes); err != nil {
			log.WithError(err).Warningf("Could not remove %s label.", cfg.CLALabelYes)
		}
	}

	if !hasCLANo {
		if err := bot.cli.AddPRLabel(p, cfg.CLALabelNo); err != nil {
			log.WithError(err).Warningf("Could not add %s label.", cfg.CLALabelNo)
		}
	}

	return bot.cli.CreatePRComment(p, signGuide(cfg.SignURL, generateUnSignComment(unsigned), cfg.FAQURL, cfg))
}

func (bot *robot) getPRCommitsAbout(
	p *atomgitclient.PRIssue,
	cfg *botConfig,
) ([]*atomgit.RepositoryCommit, error) {
	// add retry logic
	var prCommits []*atomgit.RepositoryCommit
	retryTimes := 0
	for {
		commits, err := bot.cli.GetPRCommits(p)
		if err != nil {
			retryTimes += 1

			// take a sleep before next api call
			time.Sleep(1000 * time.Millisecond)
			continue
		}

		if err == nil || retryTimes >= 2 {
			prCommits = commits
			break
		}
	}

	if len(prCommits) == 0 {
		return nil, fmt.Errorf("commits is empty, cla cannot be checked")
	}

	result := map[string]bool{}
	unsigned := make([]*atomgit.RepositoryCommit, 0, len(prCommits))
	for i := range prCommits {
		c := prCommits[i]
		email := strings.Trim(getAuthorOfCommit(c, cfg), " ")

		if !utils.IsValidEmail(email) {
			unsigned = append(unsigned, c)
			continue
		}

		if v, ok := result[email]; ok {
			if !v {
				unsigned = append(unsigned, c)
			}
			continue
		}

		b, err := isSigned(email, cfg.CheckURL)
		if err != nil {
			return nil, err
		}

		result[email] = b
		if !b {
			unsigned = append(unsigned, c)
		}
	}

	return unsigned, nil
}

func getAuthorOfCommit(c *atomgit.RepositoryCommit, cfg *botConfig) string {
	if c == nil {
		return ""
	}

	if cfg.CheckByCommitter {
		v := c.Commit.GetCommitter()

		if !cfg.LitePRCommitter.isLitePR(v.GetEmail(), v.GetName()) {
			return v.GetEmail()
		}
	}

	return c.Commit.GetAuthor().GetEmail()
}

func isSigned(email, url string) (bool, error) {
	endpoint := fmt.Sprintf("%s?email=%s", url, email)

	resp, err := http.Get(endpoint)
	if err != nil {
		return false, err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	rb, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, err
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return false, fmt.Errorf("response has status %q and body %q", resp.Status, string(rb))
	}

	type signingInfo struct {
		Signed bool `json:"signed"`
	}
	var v struct {
		Data signingInfo `json:"data"`
	}

	if err := json.Unmarshal(rb, &v); err != nil {
		return false, fmt.Errorf("unmarshal failed: %s", err.Error())
	}

	return v.Data.Signed, nil
}

func deleteSignGuide(p *atomgitclient.PRIssue, c iClient, cfg *botConfig) {
	v, err := c.GetPRComments(p)
	if err != nil {
		return
	}

	prefix := signGuideTitle(cfg)
	prefixOld := "Thanks for your pull request. Before we can look at your pull request, you'll need to sign a Contributor License Agreement (CLA)."
	fn := func(s string) bool {
		return strings.HasPrefix(s, prefix) || strings.HasPrefix(s, prefixOld)
	}

	for i := range v {
		if item := v[i]; fn(*item.Body) {
			_ = c.DeletePRComment(p.Org, p.Repo, *item.ID)
		}
	}
}

func signGuideTitle(cfg *botConfig) string {
	return fmt.Sprintf("[![cla-sign](%s \"cla-sign\")](%s)",
		"https://atomgit.com/openeuler/infrastructure/raw/master/docs/cla-guide/comment-img/cla-sign.png", cfg.SignURL)
}

func signGuide(signURL, cInfo, faq string, cfg *botConfig) string {
	s := `%s

%s

[![cla-faq](%s "cla-faq")](%s).`

	return fmt.Sprintf(s, signGuideTitle(cfg), cInfo, "https://gitee.com/openeuler/infrastructure/raw/master/docs/cla-guide/comment-img/faq-guide.png", faq)
}

func alreadySigned(user string) string {
	s := `***@%s***, thanks for your pull request. All authors of the commits have signed the CLA. :wave: `
	return fmt.Sprintf(s, user)
}

func generateUnSignComment(commits []*atomgit.RepositoryCommit) string {
	n := len(commits)
	if n == 0 {
		return ""
	}

	ten := 10
	if n > ten {
		commits = commits[:ten]
	}

	cs := make([]string, 0, len(commits))
	for _, c := range commits {
		msg := ""
		if c.Commit != nil {
			msg = *c.Commit.Message
		}

		sha := *c.SHA
		if len(sha) > maxLengthOfSHA {
			sha = sha[:maxLengthOfSHA]
		}

		cs = append(cs, fmt.Sprintf("**%s** | %s", sha, msg))
	}

	if n <= ten {
		return strings.Join(cs, "\n")
	}

	return fmt.Sprintf("Total %d commits are not signed, and 10 of them are bellow.\n%s",
		n, strings.Join(cs, "\n"),
	)
}
