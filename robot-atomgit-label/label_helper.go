package main

import (
	"strings"

	"github.com/opensourceways/community-robot-lib/atomgitclient"

	"github.com/opensourceways/go-atomgit/atomgit"

	"github.com/opensourceways/community-robot-lib/utils"
	"k8s.io/apimachinery/pkg/util/sets"
)

const (
	Issue = iota
	PullRequest
)

type labelHelper struct {
	cli                   iClient
	flag                  int
	prIssue               *atomgitclient.PRIssue
	labels                []*atomgit.Label
	commentator, commitID string
	add, remove           []string // add labels and remove labels
}

type iLabelHelper interface {
	addLabels([]string) error
	removeLabels([]string) error
	getCurrentLabels() sets.String
	addComment(string) error

	getLabelsOfRepo() ([]string, error)
	isCollaborator(string) (bool, error)
	createLabelsOfRepo(missing []string) error
}

func (l *labelHelper) isCollaborator(commenter string) (bool, error) {
	return l.cli.IsCollaborator(l.prIssue, commenter)
}

func (l *labelHelper) getLabelsOfRepo() ([]string, error) {
	labels, err := l.cli.GetRepositoryLabels(l.prIssue)
	if err != nil {
		return nil, err
	}
	return labels, nil
}

func (l *labelHelper) createLabelsOfRepo(labels []string) error {
	mErr := utils.MultiError{}

	for _, lb := range labels {
		if err := l.cli.CreateRepoLabel(l.prIssue.Org, l.prIssue.Repo, lb); err != nil {
			mErr.AddError(err)
		}
	}

	return mErr.Err()
}

func (l *labelHelper) addLabels(label []string) error {
	if l.flag == Issue {
		return l.cli.AddIssueLabel(l.prIssue, label)
	}

	errs := utils.NewMultiErrors()
	var e error
	for _, lb := range label {
		e = l.cli.AddPRLabel(l.prIssue, lb)
		if e != nil {
			errs.AddError(e)
		}
	}
	return errs.Err()
}

func (l *labelHelper) removeLabels(label []string) error {
	errs := utils.NewMultiErrors()
	var e error
	for _, lb := range label {
		if l.flag == Issue {
			e = l.cli.RemoveIssueLabel(l.prIssue, lb)
		} else {
			e = l.cli.RemovePRLabel(l.prIssue, lb)
		}

		if e != nil {
			errs.AddError(e)
		}
	}
	return errs.Err()
}

func (l *labelHelper) getCurrentLabels() sets.String {
	s := sets.String{}
	for _, lb := range l.labels {
		s.Insert(*lb.Name)
	}
	return s
}

func (l *labelHelper) addComment(comment string) error {
	if l.flag == Issue {
		return l.cli.CreateIssueComment(l.prIssue, comment)
	} else {
		return l.cli.CreatePRCommentReply(l.prIssue, comment, l.commitID)
	}
}

type labelSet struct {
	m map[string]string
	s sets.String
}

func (ls *labelSet) count() int {
	return len(ls.m)
}

func (ls *labelSet) toList() []string {
	return ls.s.UnsortedList()
}

func (ls *labelSet) origin(data []string) []string {
	r := make([]string, 0, len(data))
	for _, item := range data {
		if v, ok := ls.m[item]; ok {
			r = append(r, v)
		}
	}
	return r
}

func (ls *labelSet) intersection(ls1 *labelSet) []string {
	return ls.s.Intersection(ls1.s).UnsortedList()
}

func (ls *labelSet) difference(ls1 *labelSet) []string {
	return ls.s.Difference(ls1.s).UnsortedList()
}

func newLabelSet(data []string) *labelSet {
	m := map[string]string{}
	v := make([]string, len(data))
	for i := range data {
		v[i] = strings.ToLower(data[i])
		m[v[i]] = data[i]
	}

	return &labelSet{
		m: m,
		s: sets.NewString(v...),
	}
}
