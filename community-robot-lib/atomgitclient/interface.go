package atomgitclient

import (
	"fmt"

	"github.com/opensourceways/go-atomgit/atomgit"
)

type PRIssue struct {
	Org    string
	Repo   string
	Number int
}

func BuildPRIssue(org, repo string, number int) *PRIssue {
	return &PRIssue{
		Org:    org,
		Repo:   repo,
		Number: number,
	}
}

func (p *PRIssue) String() string {
	return fmt.Sprintf("%s/%s:%d", p.Org, p.Repo, p.Number)
}

type Client interface {
	AddPRLabel(pr *PRIssue, label string) error
	RemovePRLabel(pr *PRIssue, label string) error
	CreatePRComment(pr *PRIssue, comment string) error
	DeletePRComment(org, repo string, ID int64) error
	GetPRCommits(pr *PRIssue) ([]*atomgit.RepositoryCommit, error)
	GetPRComments(pr *PRIssue) ([]*atomgit.IssueComment, error)
	UpdatePR(pr *PRIssue, request *atomgit.PullRequest) (*atomgit.PullRequest, error)
	GetPullRequests(pr *PRIssue) ([]*atomgit.PullRequest, error)
	ListCollaborator(pr *PRIssue) ([]*atomgit.User, error)
	IsCollaborator(pr *PRIssue, login string) (bool, error)
	RemoveRepoMember(pr *PRIssue, login string) error
	AddRepoMember(pr *PRIssue, login, permission string) error
	GetPullRequestChanges(pr *PRIssue) ([]*atomgit.CommitFile, error)
	GetPRLabels(pr *PRIssue) ([]string, error)
	GetRepositoryLabels(pr *PRIssue) ([]string, error)
	UpdatePRComment(pr *PRIssue, commentID int64, ic *atomgit.IssueComment) error
	ClosePR(pr *PRIssue) error
	ReopenPR(pr *PRIssue) error
	AssignPR(pr *PRIssue, logins []string) error
	UnAssignPR(pr *PRIssue, logins []string) error
	CloseIssue(pr *PRIssue) error
	ReopenIssue(pr *PRIssue) error
	MergePR(pr *PRIssue, commitMessage string, opt *atomgit.PullRequestOptions) error
	GetRepos(org string) ([]*atomgit.Repository, error)
	GetRepo(org, repo string) (*atomgit.Repository, error)
	CreateRepo(org string, r *atomgit.Repository) error
	UpdateRepo(org, repo string, r *atomgit.Repository) error
	CreateRepoLabel(org, repo, label string) error
	GetRepoLabels(org, repo string) ([]string, error)
	AssignSingleIssue(is *PRIssue, login string) error
	UnAssignSingleIssue(is *PRIssue, login string) error
	CreateIssueComment(is *PRIssue, comment string) error
	UpdateIssueComment(is *PRIssue, commentID int64, c *atomgit.IssueComment) error
	ListIssueComments(is *PRIssue) ([]*atomgit.IssueComment, error)
	RemoveIssueLabel(is *PRIssue, label string) error
	AddIssueLabel(is *PRIssue, label []string) error
	GetIssueLabels(is *PRIssue) ([]string, error)
	UpdateIssue(is *PRIssue, iss *atomgit.IssueRequest) error
	GetSingleIssue(is *PRIssue) (*atomgit.Issue, error)
	ListBranches(org, repo string) ([]*atomgit.Branch, error)
	SetProtectionBranch(org, repo, branch string, pre *atomgit.ProtectionRequest) error
	RemoveProtectionBranch(org, repo, branch string) error
	GetDirectoryTree(org, repo, branch string, recursive bool) ([]*atomgit.TreeEntry, error)
	GetPathContent(org, repo, path, branch string) (*atomgit.RepositoryContent, error)
	CreateFile(org, repo, path, branch, commitMSG, sha string, content []byte) error
	GetUserPermissionOfRepo(org, repo, user string) (*atomgit.RepositoryPermissionLevel, error)
	CreateIssue(org, repo string, request *atomgit.IssueRequest) (*atomgit.Issue, error)
	GetRef(org, repo, ref string) (*atomgit.Reference, error)
	CreateBranch(org, repo string, reference *atomgit.Reference) error
	ListOperationLogs(pr *PRIssue) ([]*atomgit.Timeline, error)
	GetEnterprisesMember(org string) ([]*atomgit.User, error)
	GetSinglePR(pr *PRIssue) (*atomgit.PullRequest, error)
	CreatePRCommentReply(pr *PRIssue, comment, commentID string) error
}
