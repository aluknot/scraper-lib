package github

type RepoMetadata struct {
	Owner       string
	Name        string
	FullName    string
	Description string
	Stars       int
	Forks       int
	OpenIssues  int
	Language    string
	License     string
	Topics      []string
	IsPrivate   bool
	IsFork      bool
	Homepage    string
}

type ReadmeContent struct {
	Content string
}

type ProfileMetadata struct {
	Username    string
	Name        string
	Bio         string
	Followers   int
	Following   int
	PublicRepos int
	Company     string
	Location    string
	Blog        string
	Twitter     string
	Email       string
}

type ReleasesContent struct {
	Releases []Release
}

type Release struct {
	Tag  string
	Name string
	Date string
	Body string
	URL  string
}

type ContentType string

const (
	ContentTypeRepo    ContentType = "repo"
	ContentTypeProfile ContentType = "profile"
	ContentTypeRelease ContentType = "release"
)

func (c ContentType) String() string {
	return string(c)
}
