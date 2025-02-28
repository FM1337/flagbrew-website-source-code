package models

import "context"

type GitHubService interface {
	UpsertMember(ctx context.Context, m *Member) error
	UpsertRepo(ctx context.Context, r *Repo) error
	ListMembers(ctx context.Context) ([]*Member, error)
	ListRepos(ctx context.Context) ([]*Repo, error)
	ListRepo(ctx context.Context, repo string) (*Repo, error)
}

type GitHubAPI interface {
	FetchMembers() ([]*Member, error)
	FetchRepos() ([]*Repo, error)
}

type GitHubDaemon interface {
	Start()
	Stop()
}

type Member struct {
	GitHubID int64  `bson:"github_id" json:"github_id"`
	Avatar   string `bson:"avatar" json:"avatar"`
	Bio      string `bson:"bio" json:"bio"`
	Name     string `bson:"name" json:"name"`
	Username string `bson:"username" json:"username"`
	URL      string `bson:"url" json:"url"`
}

type Repo struct {
	RepoID           int64  `bson:"repo_id" json:"repo_id"`
	Name             string `bson:"name" json:"name"`
	ReadMe           string `bson:"read_me" json:"read_me"`
	Size             int    `bson:"size" json:"size"`
	Description      string `bson:"description" json:"description"`
	Commits          int    `bson:"commits" json:"commits"`
	Forks            int    `bson:"forks" json:"forks"`
	Stars            int    `bson:"stars" json:"stars"`
	TotalDownloads   int    `bson:"total_downloads" json:"total_downloads"`
	LatestReleaseCIA string `bson:"latest_release_cia" json:"latest_release_cia"`
	LatestReleaseURL string `bson:"latest_release_url" json:"latest_release_url"`
	URL              string `bson:"url" json:"url"`
}
