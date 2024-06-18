package main

import (
	"fmt"
	"regexp"
	"strings"

	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/opensourceways/community-robot-lib/utils"
)

const (
	// the gitee platform limits the maximum length of label to 20.
	labelLenLimit = 20
	lgtmLabel     = "lgtm"

	commentAddLGTMBySelf            = "***lgtm*** can not be added in your self-own pull request. :astonished:"
	commentClearLabel               = `New code changes of pr are detected and remove these labels ***%s***. :flushed: `
	commentNoPermissionForLgtmLabel = `Thanks for your review, ***%s***, your opinion is very important to us.:wave:
The maintainers will consider your advice carefully.`
	commentNoPermissionForLabel = `
***@%s*** has no permission to %s ***%s*** label in this pull request. :astonished:
Please contact to the collaborators in this repository.`
	commentAddLabel = `***%s*** was added to this pull request by: ***%s***. :wave: 
**NOTE:** If this pull request is not merged while all conditions are met, comment "/check-pr" to try again. :smile: `
	commentRemovedLabel = `***%s*** was removed in this pull request by: ***%s***. :flushed: `
)

var (
	regAddLgtm    = regexp.MustCompile(`(?mi)^/lgtm\s*$`)
	regRemoveLgtm = regexp.MustCompile(`(?mi)^/lgtm cancel\s*$`)
)

func (bot *robot) handleLGTM(p *parameter) error {

	if regAddLgtm.MatchString(p.commentContent) {
		return bot.addLGTM(p)
	}

	if regRemoveLgtm.MatchString(p.commentContent) {
		return bot.removeLGTM(p)
	}

	return nil
}

func (bot *robot) addLGTM(p *parameter) error {

	if p.author == p.commentator {
		return bot.cli.CreatePRComment(p.prArg, commentAddLGTMBySelf)
	}

	v, err := bot.hasPermission(p, p.bcf.CheckPermissionBasedOnSigOwners)
	if err != nil {
		return err
	}
	if !v {
		return bot.cli.CreatePRComment(p.prArg, fmt.Sprintf(commentNoPermissionForLgtmLabel, p.commentator))
	}

	label := genLGTMLabel(p.commentator, p.bcf.LgtmCountsRequired)
	if label != lgtmLabel {
		if err = bot.createLabelIfNeed(p, label); err != nil {
			p.log.WithError(err).Errorf("create repo label: %s", label)
		}
	}

	if err = bot.cli.AddPRLabel(p.prArg, label); err != nil {
		return err
	}

	err = bot.cli.CreatePRComment(p.prArg, fmt.Sprintf(commentAddLabel, label, p.commentator))
	if err != nil {
		p.log.Error(err)
	}

	return bot.tryMerge(p, false)
}

func (bot *robot) removeLGTM(p *parameter) error {

	if p.author != p.commentator {
		v, err := bot.hasPermission(p, p.bcf.CheckPermissionBasedOnSigOwners)
		if err != nil {
			return err
		}
		if !v {
			return bot.cli.CreatePRComment(p.prArg, fmt.Sprintf(commentNoPermissionForLabel, p.commentator, "remove", lgtmLabel))
		}

		l := genLGTMLabel(p.commentator, p.bcf.LgtmCountsRequired)
		if err = bot.cli.RemovePRLabel(p.prArg, l); err != nil {
			return err
		}

		return bot.cli.CreatePRComment(p.prArg, fmt.Sprintf(commentRemovedLabel, l, p.commentator))
	}

	// the author of pr can remove all of lgtm[-login name] kind labels
	labels := sets.New[string]()
	lbs := p.realPR.GetLabels()
	for i, j := 0, len(lbs); i < j; i++ {
		labels.Insert(*lbs[i].Name)
	}
	if v := getLGTMLabelsOnPR(labels); len(v) > 0 {
		errs := utils.NewMultiErrors()
		var e error
		for _, lb := range v {
			e = bot.cli.RemovePRLabel(p.prArg, lb)

			if e != nil {
				errs.AddError(e)
			}
		}
		return errs.Err()
	}

	return nil
}

func (bot *robot) createLabelIfNeed(p *parameter, label string) error {
	repoLabels, err := bot.cli.GetRepositoryLabels(p.prArg)
	if err != nil {
		return err
	}

	for _, v := range repoLabels {
		if v == label {
			return nil
		}
	}

	return bot.cli.CreateRepoLabel(p.prArg.Org, p.prArg.Repo, label)
}

func genLGTMLabel(commenter string, lgtmCount uint) string {
	if lgtmCount <= 1 {
		return lgtmLabel
	}

	l := fmt.Sprintf("%s-%s", lgtmLabel, strings.ToLower(commenter))
	if len(l) > labelLenLimit {
		return l[:labelLenLimit]
	}

	return l
}

func getLGTMLabelsOnPR(labels sets.Set[string]) []string {
	var r []string

	for lb, _ := range labels {
		if strings.HasPrefix(lb, lgtmLabel) {
			r = append(r, lb)
		}
	}

	return r
}
