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

	return lh.addComment("label created successful.")

	merr := utils.NewMultiErrors()

	if remove.count() > 0 {
		if _, err := lh.filterRemoveLabels(remove); err != nil {
			merr.AddError(err)
		}
	}

	if add.count() > 0 {
		err := filterAddLabels(lh, add, lh.commenter, cfg, log)
		if err != nil {
			merr.AddError(err)
		}
	}
	return merr.Err()
}

func addLabels(lh labelHelper, toAdd *labelSet, commenter string, cfg *botConfig, log *logrus.Entry) error {
	canAdd, missing, err := checkLabelsToAdd(lh, toAdd, commenter, cfg, log)
	if err != nil {
		return err
	}

	merr := utils.NewMultiErrors()

	if len(canAdd) > 0 {
		ls := sets.NewString(canAdd...).Difference(lh.getCurrentLabels())
		if ls.Len() > 0 {
			if err := lh.addLabels(ls.UnsortedList()); err != nil {
				merr.AddError(err)
			}
		}
	}

	if len(missing) > 0 {
		msg := fmt.Sprintf(
			"The label(s) `%s` cannot be applied, because the repository doesn't have them",
			strings.Join(missing, ", "),
		)

		if err := lh.addComment(msg); err != nil {
			merr.AddError(err)
		}
	}

	return merr.Err()
}

func checkLabelsToAdd(
	h labelHelper,
	toAdd *labelSet,
	commenter string,
	cfg *botConfig,
	log *logrus.Entry,
) ([]string, []string, error) {
	v, err := h.getLabelsOfRepo()
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

	b, err := h.isCollaborator(commenter)
	if err != nil {
		return nil, nil, err
	}
	if b {
		if err := h.createLabelsOfRepo(missing); err != nil {
			log.Error(err)
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
