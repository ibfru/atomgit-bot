package atomgitclient

import (
	"context"
	atomgitlib "github.com/opensourceways/go-atomgit/atomgit"
	"golang.org/x/oauth2"
)

type client struct {
	c *atomgitlib.Client
}

func NewClient(getToken func() []byte) AtomGitClient {
	ts := oauth2.StaticTokenSource(&oauth2.Token{
		AccessToken: string(getToken()),
	})
	tc := oauth2.NewClient(context.Background(), ts)

	return client{atomgitlib.NewClient(tc)}
}
