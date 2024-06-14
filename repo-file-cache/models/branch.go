package models

import (
	"fmt"
	"strings"
)

const seperator = "::"

func BranchToKey(b *Branch) string {
	return fmt.Sprintf(
		"%s%s%s%s%s%s%s",
		b.Platform, seperator,
		b.Org, seperator,
		b.Repo, seperator,
		b.Branch,
	)
}

func ParseBranch(s string) (Branch, IModelError) {
	v := strings.Split(s, seperator)
	if len(v) != 4 {
		return Branch{}, ErrInvalidBranchKey.toModelError()
	}

	return Branch{
		Platform: v[0],
		Org:      v[1],
		Repo:     v[2],
		Branch:   v[3],
	}, nil
}
