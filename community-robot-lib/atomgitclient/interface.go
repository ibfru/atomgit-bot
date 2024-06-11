package atomgitclient

import (
	"fmt"
)

type PRInfo struct {
	Org    string
	Repo   string
	Number int
}

func (p PRInfo) String() string {
	return fmt.Sprintf("%s/%s:%d", p.Org, p.Repo, p.Number)
}

type AtomGitClient interface {
}
