package dbmodels

import (
	"path/filepath"
)

const (
	FileCurrentDir    = "."
	FilePathSeparator = string(filepath.Separator)
)

type Branch struct {
	Platform string `json:"platform" required:"true"`
	Org      string `json:"org" required:"true"`
	Repo     string `json:"repo" required:"true"`
	Branch   string `json:"branch" required:"true"`
}

func (b Branch) IsMissingParam() (string, bool) {
	v := true
	if b.Platform == "" {
		return "platform", v
	}

	if b.Org == "" {
		return "org", v
	}

	if b.Repo == "" {
		return "repo", v
	}

	if b.Branch == "" {
		return "branch", v
	}
	return "", false
}

type File struct {
	Path FilePath `json:"path" required:"true"`
	SHA  string   `json:"sha" required:"true"`

	// Allow empty file
	Content string `json:"content,omitempty"`
}

func (f File) Name() string {
	return f.Path.Name()
}

func (f File) Dir() string {
	return f.Path.Dir()
}

func (f *File) IsMissingParam() (string, bool) {
	if f == nil {
		return "", false
	}

	v := true

	if f.Path == "" {
		return "path", v
	}

	if f.SHA == "" {
		return "sha", v
	}
	return "", false
}

type FilePath string

func (f FilePath) Name() string {
	return filepath.Base(string(f))
}

func (f FilePath) Dir() string {
	return filepath.Dir(filepath.Clean(string(f)))
}

func (f FilePath) FullPath() string {
	return string(f)
}
