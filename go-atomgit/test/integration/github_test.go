// Copyright 2014 The go-github AUTHORS. All rights reserved.
//
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build integration
// +build integration

package integration

import (
	"context"
	"fmt"
	"github.com/opensourceways/go-atomgit/atomgit"
	"math/rand"
	"net/http"
	"os"
)

var (
	client *atomgit.Client

	// auth indicates whether tests are being run with an OAuth token.
	// Tests can use this flag to skip certain tests when run without auth.
	auth bool
)

func init() {
	token := os.Getenv("GITHUB_AUTH_TOKEN")
	if token == "" {
		print("!!! No OAuth token. Some tests won't run. !!!\n\n")
		client = atomgit.NewClient(nil)
	} else {
		client = atomgit.NewClient(nil).WithAuthToken(token)
		auth = true
	}
}

func checkAuth(name string) bool {
	if !auth {
		fmt.Printf("No auth - skipping portions of %v\n", name)
	}
	return auth
}

func createRandomTestRepository(owner string, autoinit bool) (*atomgit.Repository, error) {
	// determine the owner to use if one wasn't specified
	if owner == "" {
		owner = os.Getenv("GITHUB_OWNER")
		if owner == "" {
			me, _, err := client.Users.Get(context.Background(), "")
			if err != nil {
				return nil, err
			}
			owner = *me.Login
		}
	}

	// create random repo name that does not currently exist
	var repoName string
	for {
		repoName = fmt.Sprintf("test-%d", rand.Int())
		_, resp, err := client.Repositories.Get(context.Background(), owner, repoName)
		if err != nil {
			if resp.StatusCode == http.StatusNotFound {
				// found a non-existent repo, perfect
				break
			}

			return nil, err
		}
	}

	// create the repository
	repo, _, err := client.Repositories.Create(
		context.Background(),
		owner,
		&atomgit.Repository{
			Name:     atomgit.String(repoName),
			AutoInit: atomgit.Bool(autoinit),
		},
	)
	if err != nil {
		return nil, err
	}

	return repo, nil
}
