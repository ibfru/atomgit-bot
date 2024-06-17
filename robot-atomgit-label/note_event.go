package main

import (
	"fmt"
	"strings"

	"github.com/opensourceways/community-robot-lib/utils"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/util/sets"
)

func (bot *robot) handleLabelsByComment(
	lh *labelHelper,
	cfg *botConfig,
	log *logrus.Entry,
) error {

	add := newLabelSet(lh.add)
	remove := newLabelSet(lh.remove)
	if v := add.intersection(remove); len(v) > 0 {
		return lh.addComment(fmt.Sprintf(
			"conflict labels(%s) exit", strings.Join(add.origin(v), ", "),
		))
	}

	merr := utils.NewMultiErrors()

	if remove.count() > 0 {
		if _, err := lh.filterRemoveLabels(remove); err != nil {
			merr.AddError(err)
		}
	}

	if add.count() > 0 {
		err := lh.filterAddLabels(add, lh.commentator, cfg, log)
		if err != nil {
			merr.AddError(err)
		}
	}
	return merr.Err()
}

func (l *labelHelper) filterAddLabels(toAdd *labelSet, commentator string, cfg *botConfig, log *logrus.Entry) error {
	canAdd, missing, err := l.checkLabelsToAdd(toAdd, commentator, cfg, log)
	if err != nil {
		return err
	}

	merr := utils.NewMultiErrors()

	if len(canAdd) > 0 {
		ls := sets.NewString(canAdd...).Difference(l.getCurrentLabels())
		if ls.Len() > 0 {
			if err := l.addLabels(ls.UnsortedList()); err != nil {
				merr.AddError(err)
			}
		}
	}

	if len(missing) > 0 {
		msg := fmt.Sprintf(
			"The label(s) `%s` cannot be applied, because the repository doesn't have them",
			strings.Join(missing, ", "),
		)

		if err = l.addComment(msg); err != nil {
			merr.AddError(err)
		}
	}

	return merr.Err()
}

func (l *labelHelper) checkLabelsToAdd(
	toAdd *labelSet,
	commentator string,
	cfg *botConfig,
	log *logrus.Entry,
) ([]string, []string, error) {
	v, err := l.getLabelsOfRepo()
	if err != nil {
		return nil, nil, err
	}
	repoLabels := newLabelSet(v)

	missing := toAdd.difference(repoLabels)
	if len(missing) == 0 {
		return repoLabels.origin(toAdd.toList()), nil, nil
	}

	var canAdd []string
	if len(missing) < toAdd.count() {
		canAdd = repoLabels.origin(toAdd.intersection(repoLabels))
	}

	missing = toAdd.origin(missing)

	if !cfg.AllowCreatingLabelsByCollaborator {
		return canAdd, missing, nil
	}

	b, err := l.isCollaborator(commentator)
	if err != nil {
		return nil, nil, err
	}
	if b {
		if e := l.createLabelsOfRepo(missing); e != nil {
			log.Error(e)
		}

		return append(canAdd, missing...), nil, nil
	}
	return canAdd, missing, nil
}

func (l *labelHelper) filterRemoveLabels(toRemove *labelSet) ([]string, error) {
	v, err := l.getLabelsOfRepo()
	if err != nil {
		return nil, err
	}
	repoLabels := newLabelSet(v)

	ls := l.getCurrentLabels().Intersection(sets.NewString(
		repoLabels.origin(toRemove.toList())...)).UnsortedList()

	if len(ls) == 0 {
		return nil, nil
	}
	return ls, l.removeLabels(ls)
}

func (bot *robot) clearLabelCaseByPRCodeUpdate(lh *labelHelper, cfg *botConfig) error {
	//if sdk.GetPullRequestAction(e) != sdk.PRActionChangedSourceBranch {
	//	return nil
	//}

	labels := lh.getCurrentLabels()
	toRemove := getClearLabels(labels, cfg)
	if len(toRemove) == 0 {
		return nil
	}

	errs := utils.NewMultiErrors()
	var e error
	for _, lb := range toRemove {
		e = bot.cli.RemovePRLabel(lh.prIssue, lb)
		if e != nil {
			errs.AddError(e)
		}
	}
	if errs != nil {
		return errs.Err()
	}

	comment := fmt.Sprintf(
		"This pull request source branch has changed, so removes the following label(s): %s.",
		strings.Join(toRemove, ", "),
	)

	return bot.cli.CreatePRComment(lh.prIssue, comment)
}

func getClearLabels(labels sets.String, cfg *botConfig) []string {
	var r []string

	all := labels
	if len(cfg.ClearLabels) > 0 {
		v := all.Intersection(sets.NewString(cfg.ClearLabels...))
		if v.Len() > 0 {
			r = v.UnsortedList()
			all = all.Difference(v)
		}
	}

	exp := cfg.clearLabelsByRegexp
	if exp != nil {
		for k := range all {
			if exp.MatchString(k) {
				r = append(r, k)
			}
		}
	}

	return r
}
