package main

import (
	"strings"

	"github.com/opensourceways/community-robot-lib/atomgitclient"

	"github.com/opensourceways/go-atomgit/atomgit"

	"github.com/opensourceways/community-robot-lib/utils"
	"k8s.io/apimachinery/pkg/util/sets"
)

type iRepoLabelHelper interface {
	getLabelsOfRepo() ([]string, error)
	isCollaborator(string) (bool, error)
	createLabelsOfRepo(missing []string) error
}

type repoLabelHelper struct {
	cli    iClient
	org    string
	repo   string
	add    []string
	remove []string
}

func (h *repoLabelHelper) isCollaborator(commenter string) (bool, error) {
	return h.cli.IsCollaborator(atomgitclient.BuildPRIssue(h.org, h.repo, 0), commenter)
}

func (h *repoLabelHelper) getLabelsOfRepo() ([]string, error) {
	labels, err := h.cli.GetRepositoryLabels(atomgitclient.BuildPRIssue(h.org, h.repo, 0))
	if err != nil {
		return nil, err
	}
	return labels, nil
}

func (h *repoLabelHelper) createLabelsOfRepo(labels []string) error {
	mErr := utils.MultiError{}

	for _, v := range labels {
		if err := h.cli.CreateRepoLabel(h.org, h.repo, v); err != nil {
			mErr.AddError(err)
		}
	}

	return mErr.Err()
}

type labelHelper interface {
	addLabels([]string) error
	removeLabels([]string) error
	getCurrentLabels() sets.String
	addComment(string) error

	iRepoLabelHelper
}

type issueLabelHelper struct {
	*repoLabelHelper

	number int
	labels sets.String
}

func (h *issueLabelHelper) addLabels(label []string) error {
	return h.cli.AddIssueLabel(atomgitclient.BuildPRIssue(h.org, h.repo, h.number), label)
}

func (h *issueLabelHelper) removeLabels(label []string) error {
	errs := utils.NewMultiErrors()
	var e error
	for _, l := range label {
		e = h.cli.RemoveIssueLabel(atomgitclient.BuildPRIssue(h.org, h.repo, h.number), l)
		if e != nil {
			errs.AddError(e)
		}
	}
	return errs.Err()
}

func (h *issueLabelHelper) getCurrentLabels() sets.String {
	return h.labels
}

func (h *issueLabelHelper) addComment(comment string) error {

	return h.cli.CreateIssueComment(atomgitclient.BuildPRIssue(h.org, h.repo, h.number), comment)
}

type prLabelHelper struct {
	*repoLabelHelper

	number    int
	labels    []*atomgit.Label
	commenter string
	commitID  string
}

func (h *prLabelHelper) addLabels(label []string) error {
	errs := utils.NewMultiErrors()
	var e error
	for _, l := range label {
		e = h.cli.AddPRLabel(atomgitclient.BuildPRIssue(h.org, h.repo, h.number), l)
		if e != nil {
			errs.AddError(e)
		}
	}
	return errs.Err()
}

func (h *prLabelHelper) removeLabels(label []string) error {
	errs := utils.NewMultiErrors()
	var e error
	for _, l := range label {
		e = h.cli.RemovePRLabel(atomgitclient.BuildPRIssue(h.org, h.repo, h.number), l)
		if e != nil {
			errs.AddError(e)
		}
	}
	return errs.Err()
}

func (h *prLabelHelper) getCurrentLabels() sets.String {
	s := sets.String{}
	for _, l := range h.labels {
		s.Insert(*l.Name)
	}
	return s
}

func (h *prLabelHelper) addComment(comment string) error {
	return h.cli.CreatePRCommentReply(atomgitclient.BuildPRIssue(h.org, h.repo, h.number), comment, h.commitID)
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
