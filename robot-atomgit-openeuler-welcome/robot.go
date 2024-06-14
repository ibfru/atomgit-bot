package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"

	"github.com/opensourceways/community-robot-lib/atomgitclient"

	"github.com/opensourceways/community-robot-lib/config"
	framework "github.com/opensourceways/community-robot-lib/robot-atomgit-framework"
	"github.com/opensourceways/community-robot-lib/utils"
	"github.com/opensourceways/go-atomgit/atomgit"
	"github.com/opensourceways/repo-file-cache/models"
	cache "github.com/opensourceways/repo-file-cache/sdk"
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

func (bot *robot) handlePREvent(e *atomgit.PullRequestEvent, pc config.Config, log *logrus.Entry) error {
	if e.GetAction() != atomgit.ActionStateCreated {
		return nil
	}

	org, repo := e.GetRepo().GetOrgAndRepo()
	cfg, err := bot.getConfig(pc, org, repo)
	if err != nil {
		return err
	}

	number := e.GetPRNumber()

	return bot.handle(
		org, repo, e.GetPRAuthor(), cfg, log,

		func(c string) error {
			return bot.cli.CreatePRComment(org, repo, number, c)
		},

		func(label string) error {
			return bot.cli.AddPRLabel(org, repo, number, label)
		},
		number,
	)
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

	author := e.GetIssue().GetUser().GetLogin()
	number := e.GetIssue().GetNumber()

	return bot.handle(
		org, repo, author, cfg, log,

		func(c string) error {
			return bot.cli.CreateIssueComment(atomgitclient.BuildPRIssue(org, repo, number), c)
		},

		func(label string) error {
			return bot.cli.AddIssueLabel(atomgitclient.BuildPRIssue(org, repo, number), []string{label})
		},
		0,
	)
}

func (bot *robot) handle(
	org, repo, author string,
	cfg *botConfig, log *logrus.Entry,
	addMsg, addLabel func(string) error,
	number int,
) error {
	mErr := utils.NewMultiErrors()
	if number > 0 {
		resp, err := http.Get(fmt.Sprintf("https://ipb.osinfra.cn/pulls?author=%s", author))
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
			if err = bot.cli.AddPRLabel(atomgitclient.BuildPRIssue(org, repo, number), "newcomer"); err != nil {
				mErr.AddError(err)
			}
		}
	}

	sigName, comment, err := bot.genComment(org, repo, author, cfg, log, number)
	if err != nil {
		return err
	}

	if err := addMsg(comment); err != nil {
		mErr.AddError(err)
	}

	label := fmt.Sprintf("sig/%s", sigName)
	if n := 20; len(label) > n {
		label = label[:n]
	}

	if err := bot.createLabelIfNeed(org, repo, label); err != nil {
		log.Errorf("create repo label:%s, err:%s", label, err.Error())
	}

	if err := addLabel(label); err != nil {
		mErr.AddError(err)
	}

	return mErr.Err()
}

func (bot *robot) genComment(org, repo, author string, cfg *botConfig, log *logrus.Entry, number int) (string, string, error) {
	sigName, err := bot.getSigOfRepo(org, repo, cfg)
	if err != nil {
		return "", "", err
	}

	if sigName == "" {
		return "", "", fmt.Errorf("cant get sig name of repo: %s/%s", org, repo)
	}

	if cfg.NoNeedToNotice {
		return sigName, fmt.Sprintf(
				welcomeMessage3, author, cfg.CommunityName, cfg.CommandLink, sigName, sigName),
			nil
	}

	// TODO use mongodb
	maintainers, committers, err := bot.getMaintainers(org, repo, sigName, number, cfg, log)
	if err != nil {
		return "", "", err
	}

	if cfg.NeedAssign && number != 0 {
		if err = bot.cli.AssignPR(atomgitclient.BuildPRIssue(org, repo, number), maintainers); err != nil {
			return "", "", err
		}
	}

	if len(committers) != 0 {
		return sigName, fmt.Sprintf(
			welcomeMessage2, author, cfg.CommunityName, cfg.CommandLink,
			sigName, sigName, strings.Join(maintainers, " , @"), strings.Join(committers, " , @"),
		), nil
	}

	return sigName, fmt.Sprintf(
		welcomeMessage, author, cfg.CommunityName, cfg.CommandLink,
		sigName, sigName, strings.Join(maintainers, " , @"),
	), nil
}

func (bot *robot) getMaintainers(org, repo, sigName string, number int, cfg *botConfig, log *logrus.Entry) ([]string, []string, error) {
	if cfg.WelcomeSimpler {
		membersToContact, err := bot.findSpecialContact(org, repo, number, cfg, log)
		if err == nil && len(membersToContact) != 0 {
			return membersToContact.UnsortedList(), nil, nil
		}
	}

	v, err := bot.cli.ListCollaborators(org, repo)
	if err != nil {
		return nil, nil, err
	}

	r := make([]string, 0, len(v))
	for i := range v {
		p := v[i].Permissions
		if p != nil && (p.Admin || p.Push) {
			r = append(r, v[i].Login)
		}
	}

	f, err := bot.getFiles("ibforuorg", "test_org", "master", "OWNERS")
	if len(f.Files) != 0 {
		return r, nil, err
	}

	s, err := bot.getFiles("ibforuorg", "test_org", "master", "sig-info.yaml")
	if len(s.Files) == 0 {
		return r, nil, err
	}

	for _, v := range s.Files {
		p := v.Path.FullPath()
		if !strings.Contains(p, sigName) {
			continue
		}
		maintainers, committers := decodeSigInfoFile(v.Content)
		return maintainers.UnsortedList(), committers.UnsortedList(), nil
	}

	return r, nil, nil
}

func (bot *robot) createLabelIfNeed(org, repo, label string) error {
	repoLabels, err := bot.cli.GetRepoLabels(org, repo)
	if err != nil {
		return err
	}

	for _, v := range repoLabels {
		if v.Name == label {
			return nil
		}
	}

	return bot.cli.CreateRepoLabel(org, repo, label, "")
}

func (bot *robot) getFiles(org, repo, branch, fileName string) (models.FilesInfo, error) {
	files, err := bot.cacheCli.GetFiles(
		models.Branch{
			Platform: "gitee",
			Org:      org,
			Repo:     repo,
			Branch:   branch,
		},
		fileName, false,
	)
	if err != nil {
		return models.FilesInfo{}, err
	}

	if len(files.Files) == 0 {
		return models.FilesInfo{}, nil
	}

	return files, nil
}

func (bot *robot) findSpecialContact(org, repo string, number int, cfg *botConfig, log *logrus.Entry) (sets.String, error) {
	if number == 0 {
		return nil, nil
	}

	changes, err := bot.cli.GetPullRequestChanges(org, repo, number)
	if err != nil {
		log.Errorf("get pr changes failed: %v", err)
		return nil, err
	}

	filePath := cfg.FilePath
	branch := cfg.FileBranch

	content, err := bot.cli.GetPathContent(org, repo, filePath, branch)
	if err != nil {
		log.Errorf("get file %s/%s/%s failed, err: %v", org, repo, filePath, err)
		return nil, err
	}

	c, err := base64.StdEncoding.DecodeString(content.Content)
	if err != nil {
		log.Errorf("decode string err: %v", err)
		return nil, err
	}

	var r Relation

	err = yaml.Unmarshal(c, &r)
	if err != nil {
		log.Errorf("yaml unmarshal failed: %v", err)
		return nil, err
	}

	owners := sets.NewString()
	var mo []Maintainer
	for _, c := range changes {
		for _, f := range r.Relations {
			for _, ff := range f.Path {
				if strings.Contains(c.Filename, ff) {
					mo = append(mo, f.Owner...)
				}
				if strings.Contains(ff, "/*/") {
					reg := regexp.MustCompile(strings.Replace(ff, "/*/", "/[^\\s]+/", -1))
					if ok := reg.MatchString(c.Filename); ok {
						mo = append(mo, f.Owner...)
					}
				}
			}
		}
	}

	for _, m := range mo {
		owners.Insert(m.GiteeID)
	}

	return owners, nil
}
