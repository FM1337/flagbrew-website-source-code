package api

import (
	"context"
	"strings"

	"github.com/FM1337/flagbrew-website-source-code/pkg/helpers"
	"github.com/FM1337/flagbrew-website-source-code/pkg/models"
	"github.com/google/go-github/github"
	"github.com/russross/blackfriday/v2"
	"golang.org/x/oauth2"
)

var git *github.Client
var ctx context.Context

var excludedRepos = []string{"FlagBot", "golem", ".github", "memecrypto"}

func initalizeGitHubAPI(accessToken string) {
	ctx = context.Background()
	staticToken := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: accessToken})
	tokenClient := oauth2.NewClient(ctx, staticToken)
	git = github.NewClient(tokenClient)
}

type GitHubAPI struct {
}

func NewGitHubAPI(accessToken string) *GitHubAPI {
	initalizeGitHubAPI(accessToken)
	return &GitHubAPI{}
}

func (*GitHubAPI) FetchMembers() (members []*models.Member, err error) {
	m, _, e := git.Organizations.ListMembers(ctx, "FlagBrew", &github.ListMembersOptions{
		PublicOnly: false,
	})
	if e != nil {
		helpers.LogToSentry(e)
		return members, e
	}
	for _, mem := range m {
		username := mem.GetLogin()
		memberInfo, _, e := git.Users.Get(ctx, username)
		if e != nil {
			helpers.LogToSentry(e)
			return members, e
		}
		members = append(members, &models.Member{
			Avatar:   memberInfo.GetAvatarURL(),
			Bio:      memberInfo.GetBio(),
			GitHubID: memberInfo.GetID(),
			Name:     memberInfo.GetName(),
			Username: username,
			URL:      memberInfo.GetHTMLURL(),
		})
	}
	return members, err
}

func (*GitHubAPI) FetchRepos() (repos []*models.Repo, err error) {
	r, _, err := git.Repositories.ListByOrg(ctx, "FlagBrew", nil)
	if err != nil {
		helpers.LogToSentry(err)
		return repos, err
	}
	for _, repo := range r {
		if repo.GetArchived() || func(repo string) bool {
			for _, excluded := range excludedRepos {
				if repo == excluded {
					return true
				}
			}
			return false
		}(repo.GetName()) {
			continue
		}
		tmpRepo := &models.Repo{}

		tmpRepo.RepoID = repo.GetID()
		tmpRepo.Stars = repo.GetStargazersCount()
		tmpRepo.Description = repo.GetDescription()
		tmpRepo.Name = repo.GetName()
		tmpRepo.Size = repo.GetSize()
		tmpRepo.URL = repo.GetHTMLURL()
		tmpRepo.Forks = repo.GetForksCount()

		if readme, _, _ := git.Repositories.GetReadme(ctx, "FlagBrew", tmpRepo.Name, nil); readme != nil {
			readme, err := readme.GetContent()
			if err != nil {
				helpers.LogToSentry(err)
				return repos, err
			}
			tmpRepo.ReadMe = string(blackfriday.Run([]byte(readme)))
		}
		if commits, resp, _ := git.Repositories.ListCommits(ctx, "FlagBrew", tmpRepo.Name, &github.CommitsListOptions{
			ListOptions: github.ListOptions{
				PerPage: 1,
			},
		}); commits != nil {
			tmpRepo.Commits = resp.LastPage
		}
		if releases, _, _ := git.Repositories.ListReleases(ctx, "FlagBrew", tmpRepo.Name, nil); releases != nil {
			latest := true
			for _, release := range releases {
				if latest && !release.GetDraft() {
					tmpRepo.LatestReleaseURL = release.GetHTMLURL()
				}

				for _, asset := range release.Assets {
					if strings.HasSuffix(asset.GetName(), ".cia") && latest && !release.GetDraft() {
						tmpRepo.LatestReleaseCIA = asset.GetBrowserDownloadURL()
					}
					tmpRepo.TotalDownloads += asset.GetDownloadCount()
				}
				if !release.GetDraft() && latest {
					latest = false
				}
			}
		}
		repos = append(repos, tmpRepo)
	}
	return repos, err
}
