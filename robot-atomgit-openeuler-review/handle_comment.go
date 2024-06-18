package main

import (
	"encoding/base64"
	"fmt"
	"regexp"
	"strings"

	"k8s.io/apimachinery/pkg/util/sets"

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

func (bot *robot) removeInvalidCLA(p *parameter) error {
	if p.commentContent != removeClaCommand {
		return nil
	}

	hasPermission, err := bot.hasPermission(p, false)
	if err != nil {
		return err
	}

	if !hasPermission {
		return bot.cli.CreatePRComment(p.prArg, fmt.Sprintf(msgNoPermissionToRemoveCla, p.commentator,
			"remove", removeLabel))
	}

	return bot.cli.RemovePRLabel(p.prArg, removeLabel)
}

func (bot *robot) handleRebase(p *parameter) error {
	if p.commentContent != rebaseCommand {
		return nil
	}

	hasPermission, err := bot.hasPermission(p, false)
	if err != nil {
		return err
	}

	if !hasPermission {
		return nil
	}

	prLabels := p.realPR.GetLabels()
	for i, j := 0, len(prLabels); i < j; i++ {
		if *prLabels[i].Name == "merge/squash" {
			return bot.cli.CreatePRComment(p.prArg,
				"Please use **/squash cancel** to remove **merge/squash** label, and try **/rebase** again")
		}
	}

	return bot.cli.AddPRLabel(p.prArg, "merge/rebase")
}

func (bot *robot) handleFlattened(p *parameter) error {
	if p.commentContent != squashCommand {
		return nil
	}

	hasPermission, err := bot.hasPermission(p, false)
	if err != nil {
		return err
	}

	if !hasPermission {
		return nil
	}

	prLabels := p.realPR.GetLabels()
	for i, j := 0, len(prLabels); i < j; i++ {
		if *prLabels[i].Name == "merge/rebase" {
			return bot.cli.CreatePRComment(p.prArg,
				"Please use **/rebase cancel** to remove **merge/rebase** label, and try **/squash** again")
		}
	}

	return bot.cli.AddPRLabel(p.prArg, "merge/squash")
}

func (bot *robot) doRetest(p *parameter) error {
	if p.action != "updated" {
		return nil
	}

	return bot.cli.CreatePRComment(p.prArg, retestCommand)
}

func (bot *robot) checkReviewer(p *parameter) error {
	if p.bcf.UnableCheckingReviewerForPR || p.action != "opened" {
		return nil
	}

	if p.realPR != nil && len(p.realPR.Assignees) > 0 {
		return nil
	}

	return bot.cli.CreatePRComment(p.prArg, fmt.Sprintf(msgNotSetReviewer, p.author))
}

func (bot *robot) clearLabel(p *parameter) error {
	if p.action != "updated" {
		return nil
	}

	labels := sets.New[string]()
	lbs := p.realPR.GetLabels()
	for i, j := 0, len(lbs); i < j; i++ {
		labels.Insert(*lbs[i].Name)
	}
	lb := getLGTMLabelsOnPR(labels)

	for i, j := 0, len(lbs); i < j; i++ {
		if *lbs[i].Name == approvedLabel {
			lb = append(lb, approvedLabel)
			break
		}
	}

	if len(lb) > 0 {

		merr := utils.NewMultiErrors()
		for _, l := range lb {
			if err := bot.cli.RemovePRLabel(p.prArg, l); err != nil {
				return err
			}
		}
		if merr != nil {
			return merr.Err()
		}

		return bot.cli.CreatePRComment(p.prArg, fmt.Sprintf(commentClearLabel, strings.Join(lb, ", ")))
	}

	return nil
}

func (bot *robot) genMergeMethod(p *parameter) string {
	mergeMethod := "merge"

	prLabels := p.realPR.GetLabels()
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
	filePath := fmt.Sprintf("sig/%s/%s/%s/%s", sig, p.prArg.Org, strings.ToLower(p.prArg.Repo[0:1]), fmt.Sprintf("%s.yaml", p.prArg.Repo))

	c, err := bot.cli.GetPathContent("openeuler", "community", filePath, "master")
	if err != nil {
		p.log.Infof("get repo %s failed, because of %v", fmt.Sprintf("%s-%s", p.prArg.Org, p.prArg.Repo), err)

		return mergeMethod
	}

	mergeMethod = bot.decodeRepoYaml(c, p.log)

	return mergeMethod
}

func (bot *robot) removeRebase(p *parameter) error {
	if p.commentContent != removeRebase {
		return nil
	}

	hasPermission, err := bot.hasPermission(p, false)
	if err != nil {
		return err
	}

	if !hasPermission {
		return nil
	}

	return bot.cli.RemovePRLabel(p.prArg, "merge/rebase")
}

func (bot *robot) removeFlattened(p *parameter) error {
	if p.commentContent != removeSquash {
		return nil
	}

	hasPermission, err := bot.hasPermission(p, false)
	if err != nil {
		return err
	}

	if !hasPermission {
		return nil
	}

	return bot.cli.RemovePRLabel(p.prArg, "merge/squash")
}

func (bot *robot) handleACK(p *parameter) error {

	if !ackCommand.MatchString(p.commentContent) {
		return nil
	}

	if p.prArg.Org != "openeuler" && p.prArg.Repo != "kernel" {
		return nil
	}

	hasPermission, err := bot.hasPermission(p, false)
	if err != nil {
		return err
	}

	if !hasPermission {
		return nil
	}

	return bot.cli.AddPRLabel(p.prArg, ackLabel)
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
