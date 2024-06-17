package atomgitclient

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"golang.org/x/oauth2"

	"github.com/opensourceways/go-atomgit/atomgit"
)

type client struct {
	c *atomgit.Client
}

func NewClient(getToken func() []byte) Client {
	ts := oauth2.StaticTokenSource(&oauth2.Token{
		AccessToken: string(getToken()),
	})
	tc := oauth2.NewClient(context.Background(), ts)

	return client{atomgit.NewClient(tc)}
}

func (cl client) AddPRLabel(pr *PRIssue, label string) error {
	_, _, err := cl.c.Issues.AddLabelsToIssue(
		context.Background(),
		pr.Org, pr.Repo, pr.Number, []string{label},
	)

	return err
}

func (cl client) RemovePRLabel(pr *PRIssue, label string) error {
	r, err := cl.c.Issues.RemoveLabelForIssue(
		context.Background(),
		pr.Org, pr.Repo, pr.Number, label,
	)
	if err != nil && r != nil && r.StatusCode == 404 {
		return nil
	}

	return err
}

func (cl client) CreatePRComment(pr *PRIssue, comment string) error {
	ic := atomgit.PullRequestComment{
		Body: atomgit.String(comment),
	}
	_, _, err := cl.c.PullRequests.CreateComment(
		context.Background(),
		pr.Org, pr.Repo, pr.Number, &ic,
	)

	return err
}

func (cl client) CreatePRCommentReply(pr *PRIssue, comment, commentID string) error {
	_, _, err := cl.c.PullRequests.CreateCommentInReplyTo(
		context.Background(),
		pr.Org, pr.Repo, pr.Number, comment, commentID,
	)

	return err
}

func (cl client) DeletePRComment(org, repo, commentId string) error {
	_, err := cl.c.PullRequests.DeleteComment(context.Background(), org, repo, commentId)

	return err
}

func (cl client) GetPRComments(pr *PRIssue) ([]*atomgit.PullRequestComment, error) {
	comments := []*atomgit.PullRequestComment{}

	opt := &atomgit.PullRequestListCommentsOptions{}
	opt.Page = 1

	for {
		v, resp, err := cl.c.PullRequests.ListComments(context.Background(), pr.Org, pr.Repo, pr.Number, opt)
		if err != nil {
			return comments, err
		}

		comments = append(comments, v...)

		link := parseLinks(resp.Header.Get("Link"))["next"]
		if link == "" {
			break
		}

		pagePath, err := url.Parse(link)
		if err != nil {
			break
		}

		p := pagePath.Query().Get("page")
		if p == "" {
			break
		}

		page, err := strconv.Atoi(p)
		if err != nil {
			break
		}
		opt.Page = page
	}

	return comments, nil
}

func (cl client) GetPRCommits(pr *PRIssue) ([]*atomgit.RepositoryCommit, error) {
	commits := []*atomgit.RepositoryCommit{}

	f := func() error {
		opt := &atomgit.ListOptions{}
		opt.Page = 1

		for {
			v, resp, err := cl.c.PullRequests.ListCommits(context.Background(), pr.Org, pr.Repo, pr.Number, nil)
			if err != nil {
				return err
			}

			commits = append(commits, v...)

			link := parseLinks(resp.Header.Get("Link"))["next"]
			if link == "" {
				break
			}

			pagePath, err := url.Parse(link)
			if err != nil {
				return fmt.Errorf("failed to parse 'next' link: %v", err)
			}

			p := pagePath.Query().Get("page")
			if p == "" {
				return fmt.Errorf("failed to get 'page' on link: %s", p)
			}

			page, err := strconv.Atoi(p)
			if err != nil {
				return err
			}

			opt.Page = page
		}

		return nil
	}

	err := f()

	return commits, err
}

func (cl client) UpdatePR(pr *PRIssue, request *atomgit.PullRequest) (*atomgit.PullRequest, error) {

	pull, _, err := cl.c.PullRequests.Edit(context.Background(), pr.Org, pr.Repo, pr.Number, request)
	if err != nil {
		return nil, err
	}

	return pull, nil
}

func (cl client) GetPullRequests(pr *PRIssue) ([]*atomgit.PullRequest, error) {
	var prs []*atomgit.PullRequest
	f := func() error {
		opt := &atomgit.ListOptions{}
		opt.Page = 1

		for {
			v, resp, err := cl.c.PullRequests.List(context.Background(), pr.Org, pr.Repo,
				&atomgit.PullRequestListOptions{ListOptions: *opt})
			if err != nil {
				return err
			}

			prs = append(prs, v...)

			link := parseLinks(resp.Header.Get("Link"))["next"]
			if link == "" {
				break
			}

			pagePath, err := url.Parse(link)
			if err != nil {
				return fmt.Errorf("failed to parse 'next' link: %v", err)
			}

			p := pagePath.Query().Get("page")
			if p == "" {
				return fmt.Errorf("failed to get 'page' on link: %s", p)
			}

			page, err := strconv.Atoi(p)
			if err != nil {
				return err
			}

			opt.Page = page
		}

		return nil
	}

	err := f()

	return prs, err
}

func (cl client) ListCollaborator(pr *PRIssue) ([]*atomgit.User, error) {
	var collaborator []*atomgit.User

	f := func() error {
		opt := &atomgit.ListOptions{}
		opt.Page = 1

		for {
			v, resp, err := cl.c.Repositories.ListCollaborators(context.Background(), pr.Org, pr.Repo,
				&atomgit.ListCollaboratorsOptions{ListOptions: *opt})
			if err != nil {
				return err
			}

			collaborator = append(collaborator, v...)

			link := parseLinks(resp.Header.Get("Link"))["next"]
			if link == "" {
				break
			}

			pagePath, err := url.Parse(link)
			if err != nil {
				return fmt.Errorf("failed to parse 'next' link: %v", err)
			}

			p := pagePath.Query().Get("page")
			if p == "" {
				return fmt.Errorf("failed to get 'page' on link: %s", p)
			}

			page, err := strconv.Atoi(p)
			if err != nil {
				return err
			}

			opt.Page = page
		}

		return nil
	}

	err := f()

	return collaborator, err
}

func (cl client) IsCollaborator(pr *PRIssue, login string) (bool, error) {
	b, _, err := cl.c.Repositories.IsCollaborator(context.Background(), pr.Org, pr.Repo, login)
	if err != nil {
		return false, err
	}
	return b, nil
}

func (cl client) RemoveRepoMember(pr *PRIssue, login string) error {
	_, err := cl.c.Repositories.RemoveCollaborator(context.Background(), pr.Org, pr.Repo, login)
	if err != nil {
		return err
	}

	return nil
}

func (cl client) AddRepoMember(pr *PRIssue, login, permission string) error {
	_, _, err := cl.c.Repositories.AddCollaborator(context.Background(), pr.Org, pr.Repo, login,
		&atomgit.RepositoryAddCollaboratorOptions{Permission: permission})
	if err != nil {
		return err
	}

	return nil
}

func (cl client) GetPullRequestChanges(pr *PRIssue) ([]*atomgit.CommitFile, error) {
	var files []*atomgit.CommitFile

	f := func() error {
		opt := &atomgit.ListOptions{}
		opt.Page = 1

		for {
			v, resp, err := cl.c.PullRequests.ListFiles(context.Background(), pr.Org, pr.Repo, pr.Number, opt)
			if err != nil {
				return err
			}

			files = append(files, v...)

			link := parseLinks(resp.Header.Get("Link"))["next"]
			if link == "" {
				break
			}

			pagePath, err := url.Parse(link)
			if err != nil {
				return fmt.Errorf("failed to parse 'next' link: %v", err)
			}

			p := pagePath.Query().Get("page")
			if p == "" {
				return fmt.Errorf("failed to get 'page' on link: %s", p)
			}

			page, err := strconv.Atoi(p)
			if err != nil {
				return err
			}

			opt.Page = page
		}

		return nil
	}

	err := f()

	return files, err
}

func (cl client) GetPRLabels(pr *PRIssue) ([]string, error) {
	pull, _, err := cl.c.PullRequests.Get(context.Background(), pr.Org, pr.Repo, pr.Number)
	if err != nil {
		return nil, err
	}

	labels := make([]string, len(pull.Labels))
	for _, p := range pull.Labels {
		labels = append(labels, *p.Name)
	}

	return labels, nil
}

func (cl client) GetRepositoryLabels(pr *PRIssue) ([]string, error) {
	var rLabels []*atomgit.Label
	f := func() error {
		opt := &atomgit.ListOptions{}
		opt.Page = 1

		for {
			v, resp, err := cl.c.Issues.ListLabels(context.Background(), pr.Org, pr.Repo, opt)
			if err != nil {
				return err
			}

			rLabels = append(rLabels, v...)

			link := parseLinks(resp.Header.Get("Link"))["next"]
			if link == "" {
				break
			}

			pagePath, err := url.Parse(link)
			if err != nil {
				return fmt.Errorf("failed to parse 'next' link: %v", err)
			}

			p := pagePath.Query().Get("page")
			if p == "" {
				return fmt.Errorf("failed to get 'page' on link: %s", p)
			}

			page, err := strconv.Atoi(p)
			if err != nil {
				return err
			}

			opt.Page = page
		}

		return nil
	}

	err := f()
	if err != nil {
		return nil, err
	}
	labels := make([]string, len(rLabels))
	for _, r := range rLabels {
		labels = append(labels, *r.Name)
	}

	return labels, nil
}

func (cl client) UpdatePRComment(pr *PRIssue, commentID int64, ic *atomgit.IssueComment) error {
	_, _, err := cl.c.Issues.EditComment(context.Background(), pr.Org, pr.Repo, commentID, ic)
	if err != nil {
		return err
	}

	return nil
}

func (cl client) ClosePR(pr *PRIssue) error {
	action := ActionClosed
	_, _, err := cl.c.PullRequests.Edit(context.Background(), pr.Org, pr.Repo, pr.Number, &atomgit.PullRequest{State: &action})
	if err != nil {
		return err
	}

	return nil
}

func (cl client) ReopenPR(pr *PRIssue) error {
	action := "open"
	_, _, err := cl.c.PullRequests.Edit(context.Background(), pr.Org, pr.Repo, pr.Number, &atomgit.PullRequest{State: &action})
	if err != nil {
		return err
	}

	return nil
}

func (cl client) AssignPR(pr *PRIssue, logins []string) error {
	_, _, err := cl.c.Issues.AddAssignees(context.Background(), pr.Org, pr.Repo, pr.Number, logins)
	if err != nil {
		return err
	}

	return nil
}

func (cl client) UnAssignPR(pr *PRIssue, logins []string) error {
	_, _, err := cl.c.Issues.RemoveAssignees(context.Background(), pr.Org, pr.Repo, pr.Number, logins)
	if err != nil {
		return err
	}

	return nil
}

func (cl client) CloseIssue(pr *PRIssue) error {
	action := ActionClosed
	_, _, err := cl.c.Issues.Edit(context.Background(), pr.Org, pr.Repo, pr.Number, &atomgit.IssueRequest{State: &action})
	if err != nil {
		return err
	}

	return nil
}

func (cl client) ReopenIssue(pr *PRIssue) error {
	action := "open"
	_, _, err := cl.c.Issues.Edit(context.Background(), pr.Org, pr.Repo, pr.Number, &atomgit.IssueRequest{State: &action})
	if err != nil {
		return err
	}

	return nil
}

func (cl client) MergePR(pr *PRIssue, commitMessage string, opt *atomgit.PullRequestOptions) error {
	_, _, err := cl.c.PullRequests.Merge(context.Background(), pr.Org, pr.Repo, pr.Number, commitMessage, opt)
	if err != nil {
		return err
	}

	return nil
}

func (cl client) GetRepos(org string) ([]*atomgit.Repository, error) {
	var rps []*atomgit.Repository
	f := func() error {
		opt := &atomgit.ListOptions{}
		opt.Page = 1
		opt.PerPage = 100

		for {
			v, resp, err := cl.c.Repositories.ListByOrg(context.Background(), org, &atomgit.RepositoryListByOrgOptions{ListOptions: *opt})
			if err != nil {
				return err
			}

			rps = append(rps, v...)

			link := parseLinks(resp.Header.Get("Link"))["next"]
			if link == "" {
				break
			}

			pagePath, err := url.Parse(link)
			if err != nil {
				return fmt.Errorf("failed to parse 'next' link: %v", err)
			}

			p := pagePath.Query().Get("page")
			if p == "" {
				return fmt.Errorf("failed to get 'page' on link: %s", p)
			}

			page, err := strconv.Atoi(p)
			if err != nil {
				return err
			}

			opt.Page = page
		}

		return nil
	}

	err := f()

	return rps, err
}

func (cl client) GetRepo(org, repo string) (*atomgit.Repository, error) {
	r, _, err := cl.c.Repositories.Get(context.Background(), org, repo)
	if err != nil {
		return nil, err
	}

	return r, nil
}

func (cl client) CreateRepo(org string, r *atomgit.Repository) error {
	_, _, err := cl.c.Repositories.Create(context.Background(), org, r)
	if err != nil {
		return err
	}

	return nil
}

func (cl client) UpdateRepo(org, repo string, r *atomgit.Repository) error {
	_, _, err := cl.c.Repositories.Edit(context.Background(), org, repo, r)
	if err != nil {
		return err
	}

	return nil
}

func (cl client) CreateRepoLabel(org, repo, label string) error {
	_, _, err := cl.c.Issues.CreateLabel(context.Background(), org, repo, &atomgit.Label{Name: &label})
	if err != nil {
		return err
	}

	return nil
}

func (cl client) GetRepoLabels(org, repo string) ([]string, error) {
	var lbs []*atomgit.Label
	f := func() error {
		opt := &atomgit.ListOptions{}
		opt.Page = 1

		for {
			v, resp, err := cl.c.Issues.ListLabels(context.Background(), org, repo, opt)
			if err != nil {
				return err
			}

			lbs = append(lbs, v...)

			link := parseLinks(resp.Header.Get("Link"))["next"]
			if link == "" {
				break
			}

			pagePath, err := url.Parse(link)
			if err != nil {
				return fmt.Errorf("failed to parse 'next' link: %v", err)
			}

			p := pagePath.Query().Get("page")
			if p == "" {
				return fmt.Errorf("failed to get 'page' on link: %s", p)
			}

			page, err := strconv.Atoi(p)
			if err != nil {
				return err
			}

			opt.Page = page
		}

		return nil
	}

	err := f()
	if err != nil {
		return nil, err
	}

	labels := make([]string, len(lbs))
	for _, l := range lbs {
		labels = append(labels, *l.Name)
	}

	return labels, nil
}

func (cl client) AssignSingleIssue(is *PRIssue, login string) error {
	_, _, err := cl.c.Issues.AddAssignees(context.Background(), is.Org, is.Repo, is.Number, []string{login})
	if err != nil {
		return err
	}

	return nil
}

func (cl client) UnAssignSingleIssue(is *PRIssue, login string) error {
	_, _, err := cl.c.Issues.RemoveAssignees(context.Background(), is.Org, is.Repo, is.Number, []string{login})
	if err != nil {
		return err
	}

	return nil
}

func (cl client) CreateIssueComment(is *PRIssue, comment string) error {
	ic := atomgit.IssueComment{
		Body: atomgit.String(comment),
	}
	_, _, err := cl.c.Issues.CreateComment(context.Background(), is.Org, is.Repo, is.Number, &ic)
	if err != nil {
		return err
	}

	return nil
}

func (cl client) UpdateIssueComment(is *PRIssue, commentID int64, c *atomgit.IssueComment) error {
	_, _, err := cl.c.Issues.EditComment(context.Background(), is.Org, is.Repo, commentID, c)
	if err != nil {
		return err
	}

	return nil
}

func (cl client) ListIssueComments(is *PRIssue) ([]*atomgit.IssueComment, error) {
	var comments []*atomgit.IssueComment

	opt := &atomgit.IssueListCommentsOptions{}
	opt.Page = 1

	for {
		v, resp, err := cl.c.Issues.ListComments(context.Background(), is.Org, is.Repo, is.Number, opt)
		if err != nil {
			return comments, err
		}

		comments = append(comments, v...)

		link := parseLinks(resp.Header.Get("Link"))["next"]
		if link == "" {
			break
		}

		pagePath, err := url.Parse(link)
		if err != nil {
			break
		}

		p := pagePath.Query().Get("page")
		if p == "" {
			break
		}

		page, err := strconv.Atoi(p)
		if err != nil {
			break
		}
		opt.Page = page
	}

	return comments, nil
}

func (cl client) RemoveIssueLabel(is *PRIssue, label string) error {
	_, err := cl.c.Issues.RemoveLabelForIssue(context.Background(), is.Org, is.Repo, is.Number, label)
	if err != nil {
		return err
	}

	return nil
}

func (cl client) AddIssueLabel(is *PRIssue, label []string) error {
	_, _, err := cl.c.Issues.AddLabelsToIssue(context.Background(), is.Org, is.Repo, is.Number, label)
	if err != nil {
		return err
	}

	return nil
}

func (cl client) GetIssueLabels(is *PRIssue) ([]string, error) {
	var lbs []*atomgit.Label
	f := func() error {
		opt := &atomgit.ListOptions{}
		opt.Page = 1

		for {
			v, resp, err := cl.c.Issues.ListLabelsByIssue(context.Background(), is.Org, is.Repo, is.Number, opt)
			if err != nil {
				return err
			}

			lbs = append(lbs, v...)

			link := parseLinks(resp.Header.Get("Link"))["next"]
			if link == "" {
				break
			}

			pagePath, err := url.Parse(link)
			if err != nil {
				return fmt.Errorf("failed to parse 'next' link: %v", err)
			}

			p := pagePath.Query().Get("page")
			if p == "" {
				return fmt.Errorf("failed to get 'page' on link: %s", p)
			}

			page, err := strconv.Atoi(p)
			if err != nil {
				return err
			}

			opt.Page = page
		}

		return nil
	}

	err := f()
	if err != nil {
		return nil, err
	}

	labels := make([]string, len(lbs))
	for _, l := range lbs {
		labels = append(labels, *l.Name)
	}

	return labels, nil
}

func (cl client) UpdateIssue(is *PRIssue, iss *atomgit.IssueRequest) error {
	_, _, err := cl.c.Issues.Edit(context.Background(), is.Org, is.Repo, is.Number, iss)
	if err != nil {
		return err
	}

	return nil
}

func (cl client) GetSingleIssue(is *PRIssue) (*atomgit.Issue, error) {
	issue, _, err := cl.c.Issues.Get(context.Background(), is.Org, is.Repo, is.Number)
	if err != nil {
		return nil, err
	}

	return issue, nil
}

func (cl client) ListBranches(org, repo string) ([]*atomgit.Branch, error) {
	var brs []*atomgit.Branch
	f := func() error {
		opt := &atomgit.ListOptions{}
		opt.Page = 1

		for {
			v, resp, err := cl.c.Repositories.ListBranches(context.Background(), org, repo,
				&atomgit.BranchListOptions{ListOptions: *opt})
			if err != nil {
				return err
			}

			brs = append(brs, v...)

			link := parseLinks(resp.Header.Get("Link"))["next"]
			if link == "" {
				break
			}

			pagePath, err := url.Parse(link)
			if err != nil {
				return fmt.Errorf("failed to parse 'next' link: %v", err)
			}

			p := pagePath.Query().Get("page")
			if p == "" {
				return fmt.Errorf("failed to get 'page' on link: %s", p)
			}

			page, err := strconv.Atoi(p)
			if err != nil {
				return err
			}

			opt.Page = page
		}

		return nil
	}

	err := f()
	if err != nil {
		return nil, err
	}

	labels := make([]string, len(brs))
	for _, b := range brs {
		labels = append(labels, *b.Name)
	}

	return brs, nil
}

func (cl client) SetProtectionBranch(org, repo, branch string, pre *atomgit.ProtectionRequest) error {
	_, _, err := cl.c.Repositories.UpdateBranchProtection(context.Background(), org, repo, branch, pre)
	if err != nil {
		return err
	}

	return nil
}

func (cl client) RemoveProtectionBranch(org, repo, branch string) error {
	_, err := cl.c.Repositories.RemoveBranchProtection(context.Background(), org, repo, branch)
	if err != nil {
		return err
	}

	return nil
}

func (cl client) GetDirectoryTree(org, repo, branch string, recursive bool) ([]*atomgit.TreeEntry, error) {
	trees, _, err := cl.c.Git.GetTree(context.Background(), org, repo, branch, recursive)
	if err != nil {
		return nil, err
	}

	return trees.Entries, nil
}

func (cl client) GetPathContent(org, repo, path, branch string) (*atomgit.RepositoryContent, error) {
	fc, _, _, err := cl.c.Repositories.GetContents(context.Background(), org, repo, path,
		&atomgit.RepositoryContentGetOptions{Ref: branch})
	if err != nil {
		return nil, err
	}

	return fc, nil
}

func (cl client) CreateFile(org, repo, path, branch, commitMSG, sha string, content []byte) error {
	_, _, err := cl.c.Repositories.CreateFile(context.Background(), org, repo, path,
		&atomgit.RepositoryContentFileOptions{Content: content, Message: &commitMSG, Branch: &branch, SHA: &sha})

	if err != nil {
		return err
	}

	return nil
}

func (cl client) GetUserPermissionOfRepo(org, repo, user string) (*atomgit.RepositoryPermissionLevel, error) {
	permission, _, err := cl.c.Repositories.GetPermissionLevel(context.Background(), org, repo, user)
	if err != nil {
		return nil, err
	}

	return permission, nil
}

func (cl client) CreateIssue(org, repo string, request *atomgit.IssueRequest) (*atomgit.Issue, error) {
	is, _, err := cl.c.Issues.Create(context.Background(), org, repo, request)
	if err != nil {
		return nil, err
	}

	return is, nil
}

func (cl client) GetRef(org, repo, ref string) (*atomgit.Reference, error) {
	r, _, err := cl.c.Git.GetRef(context.Background(), org, repo, ref)
	if err != nil {
		return nil, err
	}

	return r, nil
}

func (cl client) CreateBranch(org, repo string, reference *atomgit.Reference) error {
	_, _, err := cl.c.Git.CreateRef(context.Background(), org, repo, reference)
	if err != nil {
		return err
	}

	return nil
}

func (cl client) ListOperationLogs(pr *PRIssue) ([]*atomgit.Timeline, error) {
	var t []*atomgit.Timeline
	f := func() error {
		opt := &atomgit.ListOptions{}
		opt.Page = 1

		for {
			v, resp, err := cl.c.Issues.ListIssueTimeline(context.Background(), pr.Org, pr.Repo, pr.Number, opt)
			if err != nil {
				return err
			}

			t = append(t, v...)

			link := parseLinks(resp.Header.Get("Link"))["next"]
			if link == "" {
				break
			}

			pagePath, err := url.Parse(link)
			if err != nil {
				return fmt.Errorf("failed to parse 'next' link: %v", err)
			}

			p := pagePath.Query().Get("page")
			if p == "" {
				return fmt.Errorf("failed to get 'page' on link: %s", p)
			}

			page, err := strconv.Atoi(p)
			if err != nil {
				return err
			}

			opt.Page = page
		}

		return nil
	}

	err := f()
	if err != nil {
		return nil, err
	}

	return t, nil
}

func (cl client) GetEnterprisesMember(org string) ([]*atomgit.User, error) {
	var t []*atomgit.User
	f := func() error {
		opt := &atomgit.ListOptions{}
		opt.Page = 1

		for {
			v, resp, err := cl.c.Organizations.ListMembers(context.Background(), org,
				&atomgit.ListMembersOptions{ListOptions: *opt})
			if err != nil {
				return err
			}

			t = append(t, v...)

			link := parseLinks(resp.Header.Get("Link"))["next"]
			if link == "" {
				break
			}

			pagePath, err := url.Parse(link)
			if err != nil {
				return fmt.Errorf("failed to parse 'next' link: %v", err)
			}

			p := pagePath.Query().Get("page")
			if p == "" {
				return fmt.Errorf("failed to get 'page' on link: %s", p)
			}

			page, err := strconv.Atoi(p)
			if err != nil {
				return err
			}

			opt.Page = page
		}

		return nil
	}

	err := f()
	if err != nil {
		return nil, err
	}

	return t, nil
}

func (cl client) GetSinglePR(pr *PRIssue) (*atomgit.PullRequest, error) {
	p, _, err := cl.c.PullRequests.Get(context.Background(), pr.Org, pr.Repo, pr.Number)
	if err != nil {
		return nil, err
	}

	return p, nil
}
