package main

import (
	"encoding/base64"
	"fmt"
	"regexp"
	"strings"

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

func (bot *robot) removeInvalidCLA(e *atomgit.NoteEvent, cfg *botConfig, log *logrus.Entry) error {
	if !e.IsPullRequest() ||
		!e.IsPROpen() ||
		!e.IsCreatingCommentEvent() ||
		e.GetComment().GetBody() != removeClaCommand {
		return nil
	}

	org, repo := e.GetOrgRepo()
	number := e.GetPRNumber()
	commenter := e.GetCommenter()

	hasPermission, err := bot.hasPermission(org, repo, commenter, false, e.GetPullRequest(), cfg, log)
	if err != nil {
		return err
	}

	if !hasPermission {
		return bot.cli.CreatePRComment(org, repo, number, fmt.Sprintf(msgNoPermissionToRemoveCla, commenter,
			"remove", removeLabel))
	}

	return bot.cli.RemovePRLabel(org, repo, number, removeLabel)
}

func (bot *robot) handleRebase(e *atomgit.NoteEvent, cfg *botConfig, log *logrus.Entry) error {
	if !e.IsPullRequest() ||
		!e.IsPROpen() ||
		!e.IsCreatingCommentEvent() ||
		e.GetComment().GetBody() != rebaseCommand {
		return nil
	}

	org, repo := e.GetOrgRepo()
	number := e.GetPRNumber()
	commenter := e.GetCommenter()

	hasPermission, err := bot.hasPermission(org, repo, commenter, false, e.GetPullRequest(), cfg, log)
	if err != nil {
		return err
	}

	if !hasPermission {
		return nil
	}

	prLabels := e.GetPRLabelSet()
	if _, ok := prLabels["merge/squash"]; ok {
		return bot.cli.CreatePRComment(org, repo, number,
			"Please use **/squash cancel** to remove **merge/squash** label, and try **/rebase** again")
	}

	return bot.cli.AddPRLabel(org, repo, number, "merge/rebase")
}

func (bot *robot) handleFlattened(e *atomgit.NoteEvent, cfg *botConfig, log *logrus.Entry) error {
	if !e.IsPullRequest() ||
		!e.IsPROpen() ||
		!e.IsCreatingCommentEvent() ||
		e.GetComment().GetBody() != squashCommand {
		return nil
	}

	org, repo := e.GetOrgRepo()
	number := e.GetPRNumber()
	commenter := e.GetCommenter()

	hasPermission, err := bot.hasPermission(org, repo, commenter, false, e.GetPullRequest(), cfg, log)
	if err != nil {
		return err
	}

	if !hasPermission {
		return nil
	}

	prLabels := e.GetPRLabelSet()
	if _, ok := prLabels["merge/rebase"]; ok {
		return bot.cli.CreatePRComment(org, repo, number,
			"Please use **/rebase cancel** to remove **merge/rebase** label, and try **/squash** again")
	}

	return bot.cli.AddPRLabel(org, repo, number, "merge/squash")
}

func (bot *robot) doRetest(e *atomgit.PullRequestEvent) error {
	if atomgit.GetPullRequestAction(e) != atomgit.PRActionChangedSourceBranch {
		return nil
	}

	org, repo := e.GetOrgRepo()

	return bot.cli.CreatePRComment(org, repo, e.GetPRNumber(), retestCommand)
}

func (bot *robot) checkReviewer(e *atomgit.PullRequestEvent, cfg *botConfig) error {
	if cfg.UnableCheckingReviewerForPR || atomgit.GetPullRequestAction(e) != atomgit.ActionOpen {
		return nil
	}

	if e.GetPullRequest() != nil && len(e.GetPullRequest().Assignees) > 0 {
		return nil
	}

	org, repo := e.GetOrgRepo()

	return bot.cli.CreatePRComment(
		org, repo, e.GetPRNumber(),
		fmt.Sprintf(msgNotSetReviewer, e.GetPRAuthor()),
	)
}

func (bot *robot) clearLabel(e *atomgit.PullRequestEvent) error {
	if atomgit.GetPullRequestAction(e) != atomgit.PRActionChangedSourceBranch {
		return nil
	}

	labels := e.GetPRLabelSet()
	v := getLGTMLabelsOnPR(labels)

	if labels.Has(approvedLabel) {
		v = append(v, approvedLabel)
	}

	if len(v) > 0 {
		org, repo := e.GetOrgRepo()
		number := e.GetPRNumber()

		if err := bot.cli.RemovePRLabels(org, repo, number, v); err != nil {
			return err
		}

		return bot.cli.CreatePRComment(
			org, repo, number,
			fmt.Sprintf(commentClearLabel, strings.Join(v, ", ")),
		)
	}

	return nil
}

func (bot *robot) genMergeMethod(e *atomgit.PullRequestHook, org, repo string, log *logrus.Entry) string {
	mergeMethod := "merge"

	prLabels := e.LabelsToSet()
	sigLabel := ""

	for p := range prLabels {
		if strings.HasPrefix(p, "merge/") {
			if strings.Split(p, "/")[1] == "squash" {
				return "squash"
			}

			return strings.Split(p, "/")[1]
		}

		if strings.HasPrefix(p, "sig/") {
			sigLabel = p
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

func (bot *robot) removeRebase(e *atomgit.NoteEvent, cfg *botConfig, log *logrus.Entry) error {
	if !e.IsPullRequest() ||
		!e.IsPROpen() ||
		!e.IsCreatingCommentEvent() ||
		e.GetComment().GetBody() != removeRebase {
		return nil
	}

	org, repo := e.GetOrgRepo()
	number := e.GetPRNumber()

	commenter := e.GetCommenter()

	hasPermission, err := bot.hasPermission(org, repo, commenter, false, e.GetPullRequest(), cfg, log)
	if err != nil {
		return err
	}

	if !hasPermission {
		return nil
	}

	return bot.cli.RemovePRLabel(org, repo, number, "merge/rebase")
}

func (bot *robot) removeFlattened(e *atomgit.NoteEvent, cfg *botConfig, log *logrus.Entry) error {
	if !e.IsPullRequest() ||
		!e.IsPROpen() ||
		!e.IsCreatingCommentEvent() ||
		e.GetComment().GetBody() != removeSquash {
		return nil
	}

	org, repo := e.GetOrgRepo()
	number := e.GetPRNumber()

	commenter := e.GetCommenter()

	hasPermission, err := bot.hasPermission(org, repo, commenter, false, e.GetPullRequest(), cfg, log)
	if err != nil {
		return err
	}

	if !hasPermission {
		return nil
	}

	return bot.cli.RemovePRLabel(org, repo, number, "merge/squash")
}

func (bot *robot) handleACK(e *atomgit.NoteEvent, cfg *botConfig, log *logrus.Entry) error {
	if !e.IsPullRequest() ||
		!e.IsPROpen() ||
		!e.IsCreatingCommentEvent() {
		return nil
	}

	if !ackCommand.MatchString(e.GetComment().GetBody()) {
		return nil
	}

	org, repo := e.GetOrgRepo()
	if org != "openeuler" && repo != "kernel" {
		return nil
	}

	number := e.GetPRNumber()

	commenter := e.GetCommenter()

	hasPermission, err := bot.hasPermission(org, repo, commenter, false, e.GetPullRequest(), cfg, log)
	if err != nil {
		return err
	}

	if !hasPermission {
		return nil
	}

	return bot.cli.AddPRLabel(org, repo, number, ackLabel)
}

func (bot *robot) decodeRepoYaml(content atomgit.Content, log *logrus.Entry) string {
	c, err := base64.StdEncoding.DecodeString(content.Content)
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
