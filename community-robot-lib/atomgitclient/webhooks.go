package atomgitclient

import "fmt"

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
