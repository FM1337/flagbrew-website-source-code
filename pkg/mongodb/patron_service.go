package mongo

import (
	"context"

	"github.com/FM1337/flagbrew-website-source-code/pkg/helpers"
	"github.com/FM1337/flagbrew-website-source-code/pkg/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// patronSrv satisfies the models.PatronService interface.
type patronSrv struct {
	srv *mongoSrv
}

func (s *mongoSrv) NewPatronService() *patronSrv {
	return &patronSrv{srv: s}
}

func (s *patronSrv) IsPatron(ctx context.Context, code string) bool {
	result := s.srv.patrons.FindOne(ctx, bson.M{"code": code})
	if result.Err() != nil {
		if result.Err() != mongo.ErrNoDocuments {
			helpers.LogToSentry(result.Err())
		}
		return false
	}
	return true
}

func (s *patronSrv) GetPatronDiscord(ctx context.Context, code string) (id string, err error) {
	result := s.srv.patrons.FindOne(ctx, bson.M{"code": code})
	if result.Err() != nil {
		return id, result.Err()
	}
	patron := &models.Patron{}
	err = result.Decode(&patron)
	return patron.DiscordID, err
}
