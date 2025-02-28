package daemon

import (
	"context"
	"fmt"
	"time"

	"github.com/FM1337/flagbrew-website-source-code/pkg/helpers"
	"github.com/FM1337/flagbrew-website-source-code/pkg/models"
)

type GitHubDaemon struct {
	ghAPI   models.GitHubAPI
	ghSrv   models.GitHubService
	enabled bool
	runtime time.Duration
}

func NewGitHubDaemon(githubAPI *models.GitHubAPI, githubService *models.GitHubService) *GitHubDaemon {
	return &GitHubDaemon{ghAPI: *githubAPI, ghSrv: *githubService, enabled: false, runtime: 1 * time.Hour}
}

func (d *GitHubDaemon) Start() {
	d.enabled = true
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			if !d.enabled {
				break
			}
			mutex.Lock()
			members, err := d.ghAPI.FetchMembers()
			if err != nil {
				helpers.LogToSentry(err)
				fmt.Printf("Error fetching GitHub members, %s", err.Error())
			} else {
				for _, member := range members {
					if member.Username == "BernardoGiordano" {
						continue
					}
					err = d.ghSrv.UpsertMember(context.Background(), member)
					if err != nil {
						helpers.LogToSentry(err)
						fmt.Printf("Error upserting GitHub member, %s", err.Error())
					}
				}
			}
			repos, err := d.ghAPI.FetchRepos()
			if err != nil {
				helpers.LogToSentry(err)
				fmt.Printf("Error fetching GitHub repos, %s", err.Error())
			} else {
				for _, repo := range repos {
					err = d.ghSrv.UpsertRepo(context.Background(), repo)
					if err != nil {
						helpers.LogToSentry(err)
						fmt.Printf("Error upserting GitHub repo, %s", err.Error())
					}
				}
			}
			mutex.Unlock()
			time.Sleep(d.runtime)
		}
	}()
}

func (d *GitHubDaemon) Stop() {
	d.enabled = false
}
