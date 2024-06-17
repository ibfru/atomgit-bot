package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"

	cache "github.com/opensourceways/atomgit-sig-file-cache/sdk"
	"github.com/opensourceways/community-robot-lib/atomgitclient"
	"github.com/opensourceways/community-robot-lib/config"
	framework "github.com/opensourceways/community-robot-lib/robot-atomgit-framework"
	"github.com/opensourceways/community-robot-lib/utils"
	"github.com/opensourceways/go-atomgit/atomgit"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/util/sets"
	"sigs.k8s.io/yaml"
)

const (
	botName        = "welcome"
	welcomeMessage = `
Hi ***%s***, welcome to the %s Community.
I'm the Bot here serving you. You can find the instructions on how to interact with me at **[Here](%s)**.
If you have any questions, please contact the SIG: [%s](https://gitee.com/openeuler/community/tree/master/sig/%s), and any of the maintainers: @%s`
	welcomeMessage2 = `
Hi ***%s***, welcome to the %s Community.
I'm the Bot here serving you. You can find the instructions on how to interact with me at **[Here](%s)**.
If you have any questions, please contact the SIG: [%s](https://gitee.com/openeuler/community/tree/master/sig/%s), and any of the maintainers: @%s, any of the committers: @%s`
	welcomeMessage3 = `
Hi ***%s***, welcome to the %s Community.
I'm the Bot here serving you. You can find the instructions on how to interact with me at **[Here](%s)**.
If you have any questions, please contact the SIG: [%s](https://gitee.com/openeuler/community/tree/master/sig/%s), and any of the maintainers.
`
)

type iClient interface {
	GetRepositoryLabels(pr *atomgitclient.PRIssue) ([]string, error)
	CreateRepoLabel(org, repo, label string) error

	GetPRLabels(pr *atomgitclient.PRIssue) ([]string, error)
	AddPRLabel(pr *atomgitclient.PRIssue, label string) error

	GetIssueLabels(is *atomgitclient.PRIssue) ([]string, error)
	AddIssueLabel(is *atomgitclient.PRIssue, label []string) error

	CreatePRComment(pr *atomgitclient.PRIssue, comment string) error
	CreateIssueComment(is *atomgitclient.PRIssue, comment string) error
	CreatePRCommentReply(pr *atomgitclient.PRIssue, comment, commentID string) error

	ListCollaborator(pr *atomgitclient.PRIssue) ([]*atomgit.User, error)

	GetPathContent(org, repo, path, branch string) (*atomgit.RepositoryContent, error)
	GetDirectoryTree(org, repo, branch string, recursive bool) ([]*atomgit.TreeEntry, error)
	GetPullRequestChanges(pr *atomgitclient.PRIssue) ([]*atomgit.CommitFile, error)

	AssignPR(pr *atomgitclient.PRIssue, logins []string) error
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
	f.RegisterIssueHandler(bot.handleIssue)
	f.RegisterPullRequestHandler(bot.handlePullRequest)
}

const (
	Issue = iota
	PullRequest
)

type param struct {
	prIssue *atomgitclient.PRIssue
	cnf     *botConfig
	log     *logrus.Entry
	flag    int
	author  string
}

func (bot *robot) handlePullRequest(e *atomgit.PullRequestEvent, pc config.Config, log *logrus.Entry) error {
	if e.GetAction() != atomgit.ActionStateCreated {
		return nil
	}

	org, repo := e.GetRepo().GetOrgAndRepo()
	cfg, err := bot.getConfig(pc, org, repo)
	if err != nil {
		return err
	}

	p := &param{
		flag:    PullRequest,
		prIssue: atomgitclient.BuildPRIssue(org, repo, e.GetPullRequest().GetNumber()),
		author:  e.GetPullRequest().GetUser().GetLogin(),
		cnf:     cfg,
		log:     log,
	}

	return bot.handle(p)
}

func (bot *robot) handleIssue(e *atomgit.IssuesEvent, pc config.Config, log *logrus.Entry) error {
	if e.GetAction() != atomgit.ActionStateCreated {
		return nil
	}

	org, repo := e.GetRepo().GetOrgAndRepo()
	cfg, err := bot.getConfig(pc, org, repo)
	if err != nil {
		return err
	}

	p := &param{
		flag:    Issue,
		prIssue: atomgitclient.BuildPRIssue(org, repo, e.GetIssue().GetNumber()),
		author:  e.GetIssue().GetUser().GetLogin(),
		cnf:     cfg,
		log:     log,
	}

	return bot.handle(p)
}

func (bot *robot) handle(p *param) error {
	mErr := utils.NewMultiErrors()
	if p.flag == PullRequest {
		resp, err := http.Get(fmt.Sprintf("https://ipb.osinfra.cn/pulls?author=%s", p.author))
		if err != nil {
			mErr.AddError(err)
		}
		defer func(Body io.ReadCloser) {
			_ = Body.Close()
		}(resp.Body)
		body, _ := io.ReadAll(resp.Body)
		type T struct {
			Total int `json:"total,omitempty"`
		}

		var t T
		err = json.Unmarshal(body, &t)
		if err != nil {
			mErr.AddError(err)
		}

		if t.Total == 0 {
			if err = bot.cli.AddPRLabel(p.prIssue, "newcomer"); err != nil {
				mErr.AddError(err)
			}
		}
	}

	sigName, comment, err := bot.genComment(p)
	if err != nil {
		return err
	}

	if p.flag == Issue {
		err = bot.cli.CreateIssueComment(p.prIssue, comment)
	} else {
		err = bot.cli.CreatePRComment(p.prIssue, comment)
	}
	if err != nil {
		mErr.AddError(err)
	}

	label := fmt.Sprintf("sig/%s", sigName)
	if n := 20; len(label) > n {
		label = label[:n]
	}

	if err = bot.createLabelIfNeed(p.prIssue, label); err != nil {
		p.log.Errorf("create repo label:%s, err:%s", label, err.Error())
	}

	if p.flag == Issue {
		err = bot.cli.AddIssueLabel(p.prIssue, []string{label})
	} else {
		err = bot.cli.AddPRLabel(p.prIssue, label)
	}
	if err != nil {
		mErr.AddError(err)
	}

	return mErr.Err()
}

func (bot *robot) genComment(p *param) (string, string, error) {
	sigName, err := bot.getSigOfRepo(p.prIssue.Org, p.prIssue.Repo)
	if err != nil {
		return "", "", err
	}

	if sigName == "" {
		return "", "", fmt.Errorf("cant get sig name of repo: %s/%s", p.prIssue.Org, p.prIssue.Repo)
	}

	if p.cnf.NoNeedToNotice {
		return sigName, fmt.Sprintf(
				welcomeMessage3, p.author, p.cnf.CommunityName, p.cnf.CommandLink, sigName, sigName),
			nil
	}

	// TODO use service
	maintainers, committers, err := bot.getMaintainers(sigName, p)
	if err != nil {
		return "", "", err
	}

	if p.cnf.NeedAssign && p.prIssue.Number != 0 {
		if err = bot.cli.AssignPR(p.prIssue, maintainers); err != nil {
			return "", "", err
		}
	}

	if len(committers) != 0 {
		return sigName, fmt.Sprintf(
			welcomeMessage2, p.author, p.cnf.CommunityName, p.cnf.CommandLink,
			sigName, sigName, strings.Join(maintainers, " , @"), strings.Join(committers, " , @"),
		), nil
	}

	return sigName, fmt.Sprintf(
		welcomeMessage, p.author, p.cnf.CommunityName, p.cnf.CommandLink,
		sigName, sigName, strings.Join(maintainers, " , @"),
	), nil
}

func (bot *robot) getMaintainers(sigName string, p *param) ([]string, []string, error) {
	if p.cnf.WelcomeSimpler {
		membersToContact, err := bot.findSpecialContact(p)
		if err == nil && len(*membersToContact) != 0 {
			return membersToContact.UnsortedList(), nil, nil
		}
	}

	users, err := bot.cli.ListCollaborator(p.prIssue)
	if err != nil {
		return nil, nil, err
	}

	uLen := len(users)
	r := make([]string, 0, uLen)
	var uTmp map[string]bool
	for i := 0; i < uLen; i++ {
		uTmp = users[i].Permissions
		if uTmp != nil && (uTmp["Admin"] || uTmp["Push"]) {
			r = append(r, *users[i].Login)
		}
	}

	// when OWNERS file exists, collaborators as maintainers, committers set empty

	// when sig-info.yaml file not exists, Collaborators as maintainers, committers set empty

	// when sig-info.yaml file exists, get maintainers and committers from service[sig-info-cache]

	return r, nil, nil
}

func (bot *robot) createLabelIfNeed(pris *atomgitclient.PRIssue, label string) error {
	repoLabels, err := bot.cli.GetRepositoryLabels(pris)
	if err != nil {
		return err
	}

	for _, v := range repoLabels {
		if v == label {
			return nil
		}
	}

	return bot.cli.CreateRepoLabel(pris.Org, pris.Repo, label)
}

func (bot *robot) findSpecialContact(p *param) (*sets.Set[string], error) {
	if p.prIssue.Number == 0 {
		return nil, nil
	}

	changes, err := bot.cli.GetPullRequestChanges(p.prIssue)
	if err != nil {
		p.log.Errorf("get pr changes failed: %v", err)
		return nil, err
	}

	content, err := bot.cli.GetPathContent(p.prIssue.Org, p.prIssue.Repo, p.cnf.FilePath, p.cnf.FileBranch)
	if err != nil {
		p.log.Errorf("get file %s/%s/%s failed, err: %v", p.prIssue.Org, p.prIssue.Repo, p.cnf.FilePath, err)
		return nil, err
	}

	c, err := base64.StdEncoding.DecodeString(*content.Content)
	if err != nil {
		p.log.Errorf("decode string err: %v", err)
		return nil, err
	}

	var r Relation

	err = yaml.Unmarshal(c, &r)
	if err != nil {
		p.log.Errorf("yaml unmarshal failed: %v", err)
		return nil, err
	}

	owners := sets.New[string]()
	var mo []Maintainer
	for _, cg := range changes {
		for _, f := range r.Relations {
			for _, ff := range f.Path {
				if strings.Contains(*cg.Filename, ff) {
					mo = append(mo, f.Owner...)
				}
				if strings.Contains(ff, "/*/") {
					reg := regexp.MustCompile(strings.Replace(ff, "/*/", "/[^\\s]+/", -1))
					if ok := reg.MatchString(*cg.Filename); ok {
						mo = append(mo, f.Owner...)
					}
				}
			}
		}
	}

	for _, m := range mo {
		owners.Insert(m.GiteeID)
	}

	return &owners, nil
}
