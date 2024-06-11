package options

import (
	"flag"
)

// AtomGitOptions holds options for interacting with AtomGit.
type AtomGitOptions struct {
	TokenPath      string
	TokenGenerator func() []byte
}

// NewAtomGitOptions creates a AtomGitOptions with default values.
func NewAtomGitOptions() *AtomGitOptions {
	return &AtomGitOptions{}
}

// AddFlags injects AtomGit options into the given FlagSet.
func (o *AtomGitOptions) AddFlags(fs *flag.FlagSet) {
	o.addFlags("/etc/atomgit/oauth", fs)
}

// AddFlagsWithoutDefaultAtomGitTokenPath injects AtomGit options into the given
// Flagset without setting a default for for the AtomGitTokenPath, allowing to
// use an anonymous Gitee client
func (o *AtomGitOptions) AddFlagsWithoutDefaultAtomGitTokenPath(fs *flag.FlagSet) {
	o.addFlags("", fs)
}

func (o *AtomGitOptions) addFlags(defaultAtomGitTokenPath string, fs *flag.FlagSet) {
	fs.StringVar(
		&o.TokenPath,
		"atomgit-token-path",
		defaultAtomGitTokenPath,
		"Path to the file containing the AtomGit OAuth secret.",
	)
}

// Validate validates AtomGit options.
func (o AtomGitOptions) Validate() error {
	return nil
}
