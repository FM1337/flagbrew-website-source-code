package mongo

import (
	"context"

	"github.com/FM1337/flagbrew-website-source-code/pkg/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ghSrv satisfies the models.GithubService interface.
type ghSrv struct {
	srv *mongoSrv
}

func (s *mongoSrv) NewGitHubService() *ghSrv {
	return &ghSrv{srv: s}
}

func (s *ghSrv) UpsertMember(ctx context.Context, member *models.Member) (err error) {
	_, err = s.srv.members.UpdateOne(ctx, bson.M{"github_id": member.GitHubID}, bson.M{"$set": member}, options.Update().SetUpsert(true))
	return errorWrapper(err)
}

func (s *ghSrv) ListMembers(ctx context.Context) (members []*models.Member, err error) {
	result, err := s.srv.members.Find(ctx, bson.M{})
	result.All(ctx, &members)
	return members, errorWrapper(err)
}

func (s *ghSrv) UpsertRepo(ctx context.Context, repo *models.Repo) (err error) {
	_, err = s.srv.repos.UpdateOne(ctx, bson.M{"repo_id": repo.RepoID}, bson.M{"$set": repo}, options.Update().SetUpsert(true))
	return errorWrapper(err)
}

func (s *ghSrv) ListRepos(ctx context.Context) (repos []*models.Repo, err error) {
	results, err := s.srv.repos.Find(ctx, bson.M{})
	err = results.All(ctx, &repos)
	return repos, errorWrapper(err)
}

func (s *ghSrv) ListRepo(ctx context.Context, name string) (repo *models.Repo, err error) {
	result := s.srv.repos.FindOne(ctx, bson.M{"name": name})

	err = result.Decode(&repo)
	return repo, errorWrapper(err)
}
