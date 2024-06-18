package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/opensourceways/community-robot-lib/atomgitclient"

	"github.com/opensourceways/go-atomgit/atomgit"

	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/util/sets"
	"sigs.k8s.io/yaml"
)

const (
	msgPRConflicts        = "PR conflicts to the target branch."
	msgMissingLabels      = "PR does not have these lables: %s"
	msgInvalidLabels      = "PR should remove these labels: %s"
	msgNotEnoughLGTMLabel = "PR needs %d lgtm labels and now gets %d"
	msgFrozenWithOwner    = "The target branch of PR has been frozen and it can be merge only by branch owners: @%s"
)

var regCheckPr = regexp.MustCompile(`(?mi)^/check-pr\s*$`)

func (bot *robot) handleCheckPR(p *parameter) error {
	if !regCheckPr.MatchString(p.commentContent) {
		return nil
	}

	return bot.tryMerge(p, true)
}

func (bot *robot) tryMerge(p *parameter, addComment bool) error {
	h := &mergeHelper{
		arg:    p,
		method: bot.genMergeMethod(p),
		cli:    bot.cli,
	}

	if r, ok := h.canMerge(p.log); !ok {
		if len(r) > 0 && addComment {
			claYesLabel := ""
			for _, labelForMerge := range p.bcf.LabelsForMerge {
				if strings.Contains(labelForMerge, "-cla/yes") {
					claYesLabel = labelForMerge
					break
				}
			}
			comment := fmt.Sprintf("@%s, this pr is not mergeable and the reasons are below:\n%s\n\n***lgtm***: "+
				"A label mandatory for merging a pull request. The repository collaborators can comment '/lgtm' to "+
				"add the label. The creator of a pull request can comment '/lgtm cancel' to remove the label, but "+
				"cannot run the '/lgtm' command to add the label.\n***approved***: A label mandatory for merging a "+
				"pull request. The repository collaborators can comment '/approve' to add the label and comment "+
				"'/approve cancel' to remove the label.\n***%s***:  A label mandatory for merging a pull request. "+
				"The author of each commit of a pull request must sign the Contributor License Agreement (CLA). "+
				"Otherwise, the pull request will fail to be merged. After signing the CLA, the author can comment "+
				"'/check-cla' to check the CLA status again.\n***wait_confirm***: A label for confirming pull request "+
				"merging. A pull request with this label cannot be automatically merged. This label is added because "+
				"members (including maintainers, committers, and repository administrators) are to be added to "+
				"**sig-info.yaml** in the pull request. To remove the label, all members to be added must comment "+
				"'/lgtm' in the pull request.",
				p.commentator, strings.Join(r, "\n"), claYesLabel)
			return bot.cli.CreatePRComment(p.prArg, comment)
		}

		return nil
	}

	if err := h.merge(); err != nil {
		includeErr := "there are conflicting files"
		if !strings.Contains(err.Error(), includeErr) {
			return err
		}

		return bot.cli.CreatePRComment(p.prArg, fmt.Sprintf(prCanNotMergeNotice, p.author, h.method, err.Error()))
	}

	return nil
}

func (bot *robot) handleLabelUpdate(p *parameter) error {
	if p.action != "updated_label" {
		return nil
	}

	methodOfMerge := bot.genMergeMethod(p)

	h := &mergeHelper{
		arg:    p,
		method: bot.genMergeMethod(p),
		cli:    bot.cli,
	}

	if _, ok := h.canMerge(p.log); ok {
		if err := h.merge(); err != nil {
			includeErr := "there are conflicting files"
			if !strings.Contains(err.Error(), includeErr) {
				return err
			}

			return bot.cli.CreatePRComment(p.prArg, fmt.Sprintf(prCanNotMergeNotice, p.author, methodOfMerge, err.Error()))
		} else {
			return nil
		}
	}

	return nil
}

type mergeHelper struct {
	arg             *parameter
	cli             iClient
	method, trigger string
}

func (m *mergeHelper) merge() error {

	n, urs := m.arg.realPR.GetNeedReview()
	if n != 0 {

		// TODO
		if n == 1 && m.arg.realPR.GetAssignee() == nil {
			err := m.cli.AssignPR(m.arg.prArg, urs)
			if err != nil {
				return err
			}
		}

		// TODO Assignees

		err := m.cli.AssignPR(m.arg.prArg, urs)
		if err != nil {
			return err
		}
	}

	desc := m.genMergeDesc()

	bodyStr := ""
	if m.arg.prArg.Org == "openeuler" && m.arg.prArg.Repo == "kernel" {

		if m.arg.commentator == "openeuler-sync-bot" {
			bodySlice := strings.Split(m.arg.commentContent, "\n")
			originPR := strings.Split(strings.Replace(bodySlice[1], "### ", "", -1), "1. ")[1]
			syncRelatedPR := bodySlice[2]

			relatedPRNumber, _ := strconv.Atoi(strings.Replace(
				strings.Split(syncRelatedPR, "/")[6], "\r", "", -1))
			relatedOrg := strings.Split(syncRelatedPR, "/")[3]
			relatedRepo := strings.Split(syncRelatedPR, "/")[4]
			relatedPR, _ := m.cli.GetSinglePR(atomgitclient.BuildPRIssue(relatedOrg, relatedRepo, relatedPRNumber))
			relatedDesc := relatedPR.Body

			bodyStr = fmt.Sprintf("\n%s \n%s \n \n%s", originPR, syncRelatedPR, relatedDesc)
		} else {
			bodyStr = m.arg.commentContent
		}

		return m.cli.MergePR(
			m.arg.prArg,
			fmt.Sprintf("\n%s \n \n%s \n \n%s \n%s",
				fmt.Sprintf("Merge Pull Request from: @%s", m.arg.commentator),
				bodyStr, fmt.Sprintf("Link:%s", m.arg.realPR.GetHTMLURL()), desc),
			&atomgit.PullRequestOptions{
				MergeMethod: string(m.arg.bcf.MergeMethod),
				SHA:         *m.arg.realPR.MergeCommitSHA,
			},
		)
	}

	return m.cli.MergePR(
		m.arg.prArg,
		fmt.Sprintf("\n%s", desc),
		&atomgit.PullRequestOptions{
			MergeMethod: string(m.method),
			SHA:         *m.arg.realPR.MergeCommitSHA,
		},
	)
}

func (m *mergeHelper) canMerge(log *logrus.Entry) ([]string, bool) {
	if !m.arg.realPR.GetMergeable() {
		return []string{msgPRConflicts}, false
	}

	ops, err := m.cli.ListOperationLogs(m.arg.prArg)
	if err != nil {
		return []string{}, false
	}

	labels := m.getPRLabels()
	for label := range labels {
		for _, l := range m.arg.bcf.LabelsNotAllowMerge {
			if l == label {
				return []string{}, false
			}
		}
	}

	if r := isLabelMatched(labels, m.arg.bcf, ops, log); len(r) > 0 {
		return r, false
	}

	freeze, err := m.getFreezeInfo(log)
	if err != nil {
		return nil, false
	}

	if freeze == nil || !freeze.isFrozen() {
		return nil, true
	}

	if m.trigger == "" {
		return nil, false
	}

	if freeze.isOwner(m.trigger) {
		return nil, true
	}

	return []string{
		fmt.Sprintf(msgFrozenWithOwner, strings.Join(freeze.Owner, " , @")),
	}, false
}

func (m *mergeHelper) getFreezeInfo(log *logrus.Entry) (*freezeItem, error) {
	branch := m.arg.realPR.GetBase().GetRef()
	for _, v := range m.arg.bcf.FreezeFile {
		fc, err := m.getFreezeContent(v)
		if err != nil {
			log.Errorf("get freeze file:%s, err:%s", v.toString(), err.Error())
			return nil, err
		}

		if v := fc.getFreezeItem(m.arg.prArg.Org, branch); v != nil {
			return v, nil
		}
	}

	return nil, nil
}

func (m *mergeHelper) getFreezeContent(f freezeFile) (freezeContent, error) {
	var fc freezeContent

	c, err := m.cli.GetPathContent(f.Owner, f.Repo, f.Path, f.Branch)
	if err != nil {
		return fc, err
	}

	s, err := c.GetContent()
	if err != nil {
		return fc, err
	}

	err = yaml.Unmarshal([]byte(s), &fc)

	return fc, err
}

func (m *mergeHelper) getPRLabels() sets.Set[string] {
	labels := sets.New[string]()
	lbs := m.arg.realPR.GetLabels()
	for i, j := 0, len(lbs); i < j; i++ {
		labels.Insert(*lbs[i].Name)
	}

	if m.arg.commentator == "" {
		return labels
	}

	prLabels, err := m.cli.GetPRLabels(m.arg.prArg)
	if err != nil {
		return labels
	} else {
		labels.Clear()
	}

	for _, v := range prLabels {
		labels.Insert(v)
	}

	return labels
}

func (m *mergeHelper) genMergeDesc() string {
	comments, err := m.cli.GetPRComments(m.arg.prArg)
	if err != nil || len(comments) == 0 {
		return ""
	}

	f := func(comment *atomgit.PullRequestComment, reg *regexp.Regexp) bool {
		return reg.MatchString(comment.GetBody()) &&
			comment.UpdatedAt == comment.CreatedAt &&
			*comment.User.Login != m.arg.author
	}

	f2 := func(comment *atomgit.PullRequestComment, reg *regexp.Regexp) bool {
		return reg.MatchString(comment.GetBody()) &&
			*comment.User.Login != m.arg.author
	}

	reviewers := sets.NewString()
	signers := sets.NewString()
	ackers := sets.NewString()

	for _, c := range comments {
		if m.arg.prArg.Org == "openeuler" && m.arg.prArg.Repo == "kernel" {
			if f2(c, regAddLgtm) {
				reviewers.Insert(*c.User.Login)
			}

			if f2(c, regAddApprove) {
				signers.Insert(*c.User.Login)
			}

			if f2(c, regAck) {
				ackers.Insert(*c.User.Login)
			}
		}

		if f(c, regAddLgtm) {
			reviewers.Insert(*c.User.Login)
		}

		if f(c, regAddApprove) {
			signers.Insert(*c.User.Login)
		}
	}

	if len(signers) == 0 && len(reviewers) == 0 && len(ackers) == 0 {
		return ""
	}

	// kernel return the name and email address
	if m.arg.prArg.Org == "openeuler" && m.arg.prArg.Repo == "kernel" {
		content, e := m.cli.GetPathContent("openeuler", "community", "sig/Kernel/sig-info.yaml", "master")
		if e != nil {
			return ""
		}

		c, e1 := content.GetContent()
		if e1 != nil {
			return ""
		}

		var s SigInfo

		if err = yaml.Unmarshal([]byte(c), &s); err != nil {
			return ""
		}

		nameEmail := make(map[string]string, 80)
		contributorNameEmail := make(map[string]string, 20)
		for _, ms := range s.Maintainers {
			nameEmail[ms.GiteeID] = fmt.Sprintf("%s <%s>", ms.Name, ms.Email)
		}

		for _, i := range s.Repositories {
			for _, j := range i.Committers {
				nameEmail[j.GiteeID] = fmt.Sprintf("%s <%s>", j.Name, j.Email)
			}

			for _, k := range i.Contributors {
				contributorNameEmail[k.GiteeID] = fmt.Sprintf("%s <%s>", k.Name, k.Email)
			}
		}

		reviewersInfo := sets.NewString()
		for r, _ := range reviewers {
			if v, ok := nameEmail[r]; ok {
				reviewersInfo.Insert(v)
			}

			if v, ok := contributorNameEmail[r]; ok {
				reviewersInfo.Insert(v)
			}
		}

		signersInfo := sets.NewString()
		for s, _ := range signers {
			if v, ok := nameEmail[s]; ok {
				signersInfo.Insert(v)
			}
		}

		ackersInfo := sets.NewString()
		for a, _ := range ackers {
			if v, ok := nameEmail[a]; ok {
				ackersInfo.Insert(v)
			}
		}

		reviewedUserInfo := make([]string, 0)
		for _, item := range reviewersInfo.UnsortedList() {
			reviewedUserInfo = append(reviewedUserInfo, fmt.Sprintf("Reviewed-by: %s \n", item))
		}

		signedOffUserInfo := make([]string, 0)
		for _, item := range signersInfo.UnsortedList() {
			signedOffUserInfo = append(signedOffUserInfo, fmt.Sprintf("Signed-off-by: %s \n", item))
		}

		ackeByUserInfo := make([]string, 0)
		for _, item := range ackersInfo.UnsortedList() {
			ackeByUserInfo = append(ackeByUserInfo, fmt.Sprintf("Acked-by: %s \n", item))
		}

		return fmt.Sprintf(
			"\n%s%s%s",
			strings.Join(reviewedUserInfo, ""),
			strings.Join(signedOffUserInfo, ""),
			strings.Join(ackeByUserInfo, ""),
		)
	}

	return fmt.Sprintf(
		"From: @%s \nReviewed-by: @%s \nSigned-off-by: @%s \n",
		m.arg.author,
		strings.Join(reviewers.UnsortedList(), ", @"),
		strings.Join(signers.UnsortedList(), ", @"),
	)
}

func isLabelMatched(labels sets.Set[string], cfg *botConfig, ops []*atomgit.Timeline, log *logrus.Entry) []string {
	var reasons []string

	needs := sets.New[string](approvedLabel)
	needs.Insert(cfg.LabelsForMerge...)

	if ln := cfg.LgtmCountsRequired; ln == 1 {
		needs.Insert(lgtmLabel)
	} else {
		v := getLGTMLabelsOnPR(labels)
		if n := uint(len(v)); n < ln {
			reasons = append(reasons, fmt.Sprintf(msgNotEnoughLGTMLabel, ln, n))
		}
	}

	s := checkLabelsLegal(labels, needs, ops, cfg.LegalOperator, log)
	if s != "" {
		reasons = append(reasons, s)
	}

	if v := needs.Difference(labels); v.Len() > 0 {
		vl := v.UnsortedList()
		var vlp []string
		for _, i := range vl {
			vlp = append(vlp, fmt.Sprintf("***%s***", i))
		}
		reasons = append(reasons, fmt.Sprintf(
			msgMissingLabels, strings.Join(vlp, ", "),
		))
	}

	if len(cfg.MissingLabelsForMerge) > 0 {
		missing := sets.New[string](cfg.MissingLabelsForMerge...)
		if v := missing.Intersection(labels); v.Len() > 0 {
			vl := v.UnsortedList()
			var vlp []string
			for _, i := range vl {
				vlp = append(vlp, fmt.Sprintf("***%s***", i))
			}
			reasons = append(reasons, fmt.Sprintf(
				msgInvalidLabels, strings.Join(vlp, ", "),
			))
		}
	}

	return reasons
}

type labelLog struct {
	label string
	who   string
	t     time.Time
}

func getLatestLog(ops []*atomgit.Timeline, label string, log *logrus.Entry) (labelLog, bool) {
	var t time.Time

	index := -1

	for i := range ops {
		op := ops[i]

		// TODO
		if op.GetEvent() != "1231" || !strings.Contains(op.GetBody(), label) {
			continue
		}

		ut := op.GetCreatedAt()

		if index < 0 || ut.After(t) {
			t = ut.Local()
			index = i
		}
	}

	if index >= 0 {
		if user := ops[index].GetUser(); user != nil && user.GetLogin() != "" {
			return labelLog{
				label: label,
				t:     t,
				who:   user.GetLogin(),
			}, true
		}
	}

	return labelLog{}, false
}

func checkLabelsLegal(labels sets.Set[string], needs sets.Set[string], ops []*atomgit.Timeline, legalOperator string,
	log *logrus.Entry) string {
	f := func(label string) string {
		v, b := getLatestLog(ops, label, log)
		if !b {
			return fmt.Sprintf("The corresponding operation log is missing. you should delete " +
				"the label and add it again by correct way")
		}

		if v.who != legalOperator {
			if strings.HasPrefix(v.label, "openeuler-cla/") {
				return fmt.Sprintf("%s You can't add %s by yourself, "+
					"please remove it and use /check-cla to add it", v.who, v.label)
			}

			return fmt.Sprintf("%s You can't add %s by yourself, please contact the maintainers", v.who, v.label)
		}

		return ""
	}

	v := make([]string, 0, len(labels))

	for label := range labels {
		if ok := needs.Has(label); ok || strings.HasPrefix(label, lgtmLabel) {
			if s := f(label); s != "" {
				v = append(v, fmt.Sprintf("%s: %s", label, s))
			}
		}
	}

	if n := len(v); n > 0 {
		s := "label is"

		if n > 1 {
			s = "labels are"
		}

		return fmt.Sprintf("**The following %s not ready**.\n\n%s", s, strings.Join(v, "\n\n"))
	}

	return ""
}
