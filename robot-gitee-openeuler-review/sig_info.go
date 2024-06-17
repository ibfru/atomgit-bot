package main

// SigInfo struct.
type SigInfo struct {
	Name         string       `json:"name,omitempty"`
	Description  string       `json:"description,omitempty"`
	MailingList  string       `json:"mailing_list,omitempty"`
	MeetingURL   string       `json:"meeting_url,omitempty"`
	MatureLevel  string       `json:"mature_level,omitempty"`
	Mentors      []Mentor     `json:"mentors,omitempty"`
	Maintainers  []Maintainer `json:"maintainers,omitempty"`
	Repositories []RepoAdmin  `json:"repositories,omitempty"`
	Branches     []Branches   `json:"branches,omitempty"`
}

// Maintainer struct.
type Maintainer struct {
	GiteeID      string `json:"gitee_id,omitempty"`
	Name         string `json:"name,omitempty"`
	Organization string `json:"organization,omitempty"`
	Email        string `json:"email,omitempty"`
}

// RepoAdmin struct.
type RepoAdmin struct {
	Repo         []string      `json:"repo,omitempty"`
	Admins       []Admin       `json:"admins,omitempty"`
	Committers   []Committer   `json:"committers,omitempty"`
	Contributors []Contributor `json:"contributors,omitempty"`
}

// Contributor struct.
type Contributor struct {
	GiteeID      string `json:"gitee_id,omitempty"`
	Name         string `json:"name,omitempty"`
	Organization string `json:"organization,omitempty"`
	Email        string `json:"email,omitempty"`
}

// Mentor struct.
type Mentor struct {
	GiteeID      string `json:"gitee_id,omitempty"`
	Name         string `json:"name,omitempty"`
	Organization string `json:"organization,omitempty"`
	Email        string `json:"email,omitempty"`
}

// Committer struct.
type Committer struct {
	GiteeID      string `json:"gitee_id,omitempty"`
	Name         string `json:"name,omitempty"`
	Organization string `json:"organization,omitempty"`
	Email        string `json:"email,omitempty"`
}

// Admin struct.
type Admin struct {
	GiteeID      string `json:"gitee_id,omitempty"`
	Name         string `json:"name,omitempty"`
	Organization string `json:"organization,omitempty"`
	Email        string `json:"email,omitempty"`
}

// Repository struct.
type Repository struct {
	Name        string   `json:"name,omitempty"`
	Description string   `json:"description,omitempty"`
	MergeMethod string   `json:"merge_method,omitempty"`
	Branches    []Branch `json:"branches,omitempty"`
	Type        string   `json:"type,omitempty"`
}

// Branch struct.
type Branch struct {
	Name       string `json:"name,omitempty"`
	CreateFrom string `json:"create_from,omitempty"`
	Type       string `json:"type,omitempty"`
}

// Branches struct.
type Branches struct {
	RepoBranch []RepoBranch `json:"repo_branch,omitempty"`
	Keeper     []Keeper     `json:"keeper,omitempty"`
}

// RepoBranch struct
type RepoBranch struct {
	Repo   string `json:"repo"`
	Branch string `json:"branch"`
}

// Keeper struct.
type Keeper struct {
	GiteeID string `json:"gitee_id,omitempty"`
	Name    string `json:"name,omitempty"`
	Email   string `json:"email,omitempty"`
}
