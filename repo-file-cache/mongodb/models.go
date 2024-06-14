package mongodb

const (
	fieldBranch    = "branch"
	fieldName      = "name"
	fieldFiles     = "files"
	fieldFilesItem = "files."
	fieldFilesRef  = "$files"
	fieldFilesVal  = "files.v.v"
	fieldSHA       = "sha"
	fieldKey       = "k"
	fieldValue     = "v"
)

type dFiles struct {
	// Branch stands for a branch of repo. The format of it is platform:org:repo:branch.
	Branch string `bson:"branch" json:"branch" required:"true"`

	// Name is the file name.
	Name string `bson:"name" json:"name" required:"true"`

	BranchSHA string `bson:"sha" json:"sha" required:"true"`

	// The key of Files is the path of file.
	// Specially the key of file which is in the root directory is '/'.
	// The path can't include character of '.'.
	Files map[string]dFile `bson:"files" json:"files" required:"true"`
}

type dFile struct {
	SHA     string `bson:"k" json:"k" required:"true"`
	Content string `bson:"v" json:"v" required:"true"`
}
