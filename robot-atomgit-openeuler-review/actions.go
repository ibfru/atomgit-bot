package main

import (
	"encoding/base64"
	"fmt"
	"regexp"
	"strings"

	"github.com/opensourceways/community-robot-lib/atomgitclient"
	"github.com/opensourceways/community-robot-lib/utils"

	"github.com/opensourceways/go-atomgit/atomgit"
	"github.com/sirupsen/logrus"
	"sigs.k8s.io/yaml"
)

const (
	retestCommand              = "/retest"
	removeClaCommand           = "/cla cancel"
	rebaseCommand              = "/rebase"
	removeRebase               = "/rebase cancel"
	removeSquash               = "/squash cancel"
	baseMergeMethod            = "merge"
	squashCommand              = "/squash"
	removeLabel                = "openeuler-cla/yes"
	ackLabel                   = "Acked"
	msgNotSetReviewer          = "**@%s** Thank you for submitting a PullRequest. It is detected that you have not set a reviewer, please set a one."
	msgNoPermissionToRemoveCla = "**@%s** has no permission to %s ***%s*** label in this pull request. :astonished:\nPlease contact to the collaborators in this repository."
	prCanNotMergeNotice        = "**@%s** This pull request can not be merged by %s. :astonished:\nPlease check the error message: %s"
)

var (
	regAck     = regexp.MustCompile(`(?mi)^/ack\s*$`)
	ackCommand = regexp.MustCompile(`(?mi)^/ack\s*$`)
)

type param struct {
	prIssue               *atomgitclient.PRIssue
	pr                    *atomgit.PullRequest
	cnf                   *botConfig
	log                   *logrus.Entry
	commentator, commitID string
	author                string
}

func (bot *robot) removeInvalidCLA(e *atomgit.PullRequestReviewCommentEvent, cfg *botConfig, log *logrus.Entry) error {
	if e.GetPullRequest().GetState() != "open" || e.GetComment().GetBody() != removeClaCommand {
		return nil
	}

	org, repo := e.GetRepo().GetOrgAndRepo()
	p := &param{
		prIssue:     atomgitclient.BuildPRIssue(org, repo, e.GetPullRequest().GetNumber()),
		pr:          e.GetPullRequest(),
		cnf:         cfg,
		log:         log,
		commentator: e.GetComment().GetUser().GetLogin(),
		commitID:    e.GetComment().GetCommitID(),
	}

	hasPermission, err := bot.hasPermission(p, false)
	if err != nil {
		return err
	}

	if !hasPermission {
		return bot.cli.CreatePRComment(p.prIssue, fmt.Sprintf(msgNoPermissionToRemoveCla, p.commentator,
			"remove", removeLabel))
	}

	return bot.cli.RemovePRLabel(p.prIssue, removeLabel)
}

func (bot *robot) handleRebase(e *atomgit.PullRequestReviewCommentEvent, cfg *botConfig, log *logrus.Entry) error {
	if e.GetPullRequest().GetState() != "open" || e.GetComment().GetBody() != rebaseCommand {
		return nil
	}

	org, repo := e.GetRepo().GetOrgAndRepo()
	p := &param{
		prIssue:     atomgitclient.BuildPRIssue(org, repo, e.GetPullRequest().GetNumber()),
		pr:          e.GetPullRequest(),
		cnf:         cfg,
		log:         log,
		commentator: e.GetComment().GetUser().GetLogin(),
		commitID:    e.GetComment().GetCommitID(),
	}

	hasPermission, err := bot.hasPermission(p, false)
	if err != nil {
		return err
	}

	if !hasPermission {
		return nil
	}

	prLabels := e.GetPullRequest().GetLabels()
	for i, j := 0, len(prLabels); i < j; i++ {
		if *prLabels[i].Name == "merge/squash" {
			return bot.cli.CreatePRComment(p.prIssue,
				"Please use **/squash cancel** to remove **merge/squash** label, and try **/rebase** again")
		}
	}

	return bot.cli.AddPRLabel(p.prIssue, "merge/rebase")
}

func (bot *robot) handleFlattened(e *atomgit.PullRequestReviewCommentEvent, cfg *botConfig, log *logrus.Entry) error {
	if e.GetPullRequest().GetState() != "open" || e.GetComment().GetBody() != squashCommand {
		return nil
	}

	org, repo := e.GetRepo().GetOrgAndRepo()
	p := &param{
		prIssue:     atomgitclient.BuildPRIssue(org, repo, e.GetPullRequest().GetNumber()),
		pr:          e.GetPullRequest(),
		cnf:         cfg,
		log:         log,
		commentator: e.GetComment().GetUser().GetLogin(),
		commitID:    e.GetComment().GetCommitID(),
	}

	hasPermission, err := bot.hasPermission(p, false)
	if err != nil {
		return err
	}

	if !hasPermission {
		return nil
	}

	prLabels := e.GetPullRequest().GetLabels()
	for i, j := 0, len(prLabels); i < j; i++ {
		if *prLabels[i].Name == "merge/rebase" {
			return bot.cli.CreatePRComment(p.prIssue,
				"Please use **/rebase cancel** to remove **merge/rebase** label, and try **/squash** again")
		}
	}

	return bot.cli.AddPRLabel(p.prIssue, "merge/squash")
}

func (bot *robot) doRetest(e *atomgit.PullRequestEvent) error {
	if e.GetAction() != "updated" {
		return nil
	}

	org, repo := e.GetRepo().GetOrgAndRepo()

	return bot.cli.CreatePRComment(atomgitclient.BuildPRIssue(org, repo, e.GetPullRequest().GetNumber()), retestCommand)
}

func (bot *robot) checkReviewer(e *atomgit.PullRequestEvent, cfg *botConfig) error {
	if cfg.UnableCheckingReviewerForPR || e.GetAction() != "opened" {
		return nil
	}

	if e.GetPullRequest() != nil && len(e.GetPullRequest().Assignees) > 0 {
		return nil
	}

	org, repo := e.GetRepo().GetOrgAndRepo()

	return bot.cli.CreatePRComment(
		atomgitclient.BuildPRIssue(org, repo, e.GetPullRequest().GetNumber()),
		fmt.Sprintf(msgNotSetReviewer, e.GetSender().GetLogin()),
	)
}

func (bot *robot) clearLabel(e *atomgit.PullRequestEvent) error {
	if e.GetAction() != "updated" {
		return nil
	}

	labels := e.GetPullRequest().GetLabels()
	lb := getLGTMLabelsOnPR(labels)

	for i, j := 0, len(labels); i < j; i++ {
		if *labels[i].Name == approvedLabel {
			lb = append(lb, approvedLabel)
			break
		}
	}

	if len(lb) > 0 {
		org, repo := e.GetRepo().GetOrgAndRepo()
		pr := atomgitclient.BuildPRIssue(org, repo, e.GetPullRequest().GetNumber())

		merr := utils.NewMultiErrors()
		for _, l := range lb {
			if err := bot.cli.RemovePRLabel(pr, l); err != nil {
				return err
			}
		}
		if merr != nil {
			return merr.Err()
		}

		return bot.cli.CreatePRComment(
			pr,
			fmt.Sprintf(commentClearLabel, strings.Join(lb, ", ")),
		)
	}

	return nil
}

func (bot *robot) genMergeMethod(e *atomgit.PullRequest, org, repo string, log *logrus.Entry) string {
	mergeMethod := "merge"

	prLabels := e.GetLabels()
	sigLabel := ""

	for i, j := 0, len(prLabels); i < j; i++ {
		if strings.HasPrefix(*prLabels[i].Name, "merge/") {
			if strings.Split(*prLabels[i].Name, "/")[1] == "squash" {
				return "squash"
			}

			return strings.Split(*prLabels[i].Name, "/")[1]
		}

		if strings.HasPrefix(*prLabels[i].Name, "sig/") {
			sigLabel = *prLabels[i].Name
		}
	}

	if sigLabel == "" {
		return mergeMethod
	}

	sig := strings.Split(sigLabel, "/")[1]
	filePath := fmt.Sprintf("sig/%s/%s/%s/%s", sig, org, strings.ToLower(repo[0:1]), fmt.Sprintf("%s.yaml", repo))

	c, err := bot.cli.GetPathContent("openeuler", "community", filePath, "master")
	if err != nil {
		log.Infof("get repo %s failed, because of %v", fmt.Sprintf("%s-%s", org, repo), err)

		return mergeMethod
	}

	mergeMethod = bot.decodeRepoYaml(c, log)

	return mergeMethod
}

func (bot *robot) removeRebase(e *atomgit.PullRequestReviewCommentEvent, cfg *botConfig, log *logrus.Entry) error {
	if e.GetPullRequest().GetState() != "open" || e.GetComment().GetBody() != removeRebase {
		return nil
	}

	org, repo := e.GetRepo().GetOrgAndRepo()
	p := &param{
		prIssue:     atomgitclient.BuildPRIssue(org, repo, e.GetPullRequest().GetNumber()),
		pr:          e.GetPullRequest(),
		cnf:         cfg,
		log:         log,
		commentator: e.GetComment().GetUser().GetLogin(),
		commitID:    e.GetComment().GetCommitID(),
	}

	hasPermission, err := bot.hasPermission(p, false)
	if err != nil {
		return err
	}

	if !hasPermission {
		return nil
	}

	return bot.cli.RemovePRLabel(p.prIssue, "merge/rebase")
}

func (bot *robot) removeFlattened(e *atomgit.PullRequestReviewCommentEvent, cfg *botConfig, log *logrus.Entry) error {
	if e.GetPullRequest().GetState() != "open" || e.GetComment().GetBody() != removeSquash {
		return nil
	}

	org, repo := e.GetRepo().GetOrgAndRepo()
	p := &param{
		prIssue:     atomgitclient.BuildPRIssue(org, repo, e.GetPullRequest().GetNumber()),
		pr:          e.GetPullRequest(),
		cnf:         cfg,
		log:         log,
		commentator: e.GetComment().GetUser().GetLogin(),
		commitID:    e.GetComment().GetCommitID(),
	}

	hasPermission, err := bot.hasPermission(p, false)
	if err != nil {
		return err
	}

	if !hasPermission {
		return nil
	}

	return bot.cli.RemovePRLabel(p.prIssue, "merge/squash")
}

func (bot *robot) handleACK(e *atomgit.PullRequestReviewCommentEvent, cfg *botConfig, log *logrus.Entry) error {
	if e.GetPullRequest().GetState() != "open" {
		return nil
	}

	if !ackCommand.MatchString(e.GetComment().GetBody()) {
		return nil
	}

	org, repo := e.GetRepo().GetOrgAndRepo()
	if org != "openeuler" && repo != "kernel" {
		return nil
	}

	p := &param{
		prIssue:     atomgitclient.BuildPRIssue(org, repo, e.GetPullRequest().GetNumber()),
		pr:          e.GetPullRequest(),
		cnf:         cfg,
		log:         log,
		commentator: e.GetComment().GetUser().GetLogin(),
		commitID:    e.GetComment().GetCommitID(),
	}

	hasPermission, err := bot.hasPermission(p, false)
	if err != nil {
		return err
	}

	if !hasPermission {
		return nil
	}

	return bot.cli.AddPRLabel(p.prIssue, ackLabel)
}

func (bot *robot) decodeRepoYaml(content *atomgit.RepositoryContent, log *logrus.Entry) string {
	c, err := base64.StdEncoding.DecodeString(*content.Content)
	if err != nil {
		log.WithError(err).Error("decode file")

		return baseMergeMethod
	}

	var r Repository
	if err = yaml.Unmarshal(c, &r); err != nil {
		log.WithError(err).Error("code yaml file")

		return baseMergeMethod
	}

	if r.MergeMethod != "" {
		if r.MergeMethod == "rebase" || r.MergeMethod == "squash" {
			return r.MergeMethod
		}
	}

	return baseMergeMethod
}
