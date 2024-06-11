package options

import (
	"flag"
	"fmt"
)

// GiteeOptions holds options for interacting with Gitee.
type GiteeOptions struct {
	TokenPath     string
	RepoCacheDir  string
	CacheRepoOnPV bool
}

// NewGiteeOptions creates a GiteeOptions with default values.
func NewGiteeOptions() *GiteeOptions {
	return &GiteeOptions{}
}

// AddFlags injects Gitee options into the given FlagSet.
func (o *GiteeOptions) AddFlags(fs *flag.FlagSet) {
	o.addFlags("/etc/gitee/oauth", fs)
}

// AddFlagsWithoutDefaultGiteeTokenPath injects Gitee options into the given
// Flagset without setting a default for for the giteeTokenPath, allowing to
// use an anonymous Gitee client
func (o *GiteeOptions) AddFlagsWithoutDefaultGiteeTokenPath(fs *flag.FlagSet) {
	o.addFlags("", fs)
}

func (o *GiteeOptions) addFlags(defaultGiteeTokenPath string, fs *flag.FlagSet) {
	fs.StringVar(&o.TokenPath, "gitee-token-path", defaultGiteeTokenPath, "Path to the file containing the Gitee OAuth secret.")
	fs.StringVar(&o.RepoCacheDir, "repo-cache-dir", "", "Path to which clone repo.")
	fs.BoolVar(&o.CacheRepoOnPV, "cache-repo-on-pv", false, "Specify whether to cache repo on persistent volume.")
}

// Validate validates Gitee options.
func (o GiteeOptions) Validate() error {
	if o.CacheRepoOnPV && o.RepoCacheDir == "" {
		return fmt.Errorf("must set repo-cache-dir if caching repo on persistent volume")
	}
	return nil
}
