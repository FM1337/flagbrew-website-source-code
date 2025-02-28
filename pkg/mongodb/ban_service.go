package mongo

import (
	"context"
	"fmt"
	"math"

	"github.com/FM1337/flagbrew-website-source-code/pkg/helpers"
	"github.com/FM1337/flagbrew-website-source-code/pkg/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// banSrv satisfies the models.BanService interface.
type banSrv struct {
	srv *mongoSrv
}

func (s *mongoSrv) NewBanService() *banSrv {
	return &banSrv{srv: s}
}

func (s *banSrv) Ban(ctx context.Context, ban *models.Ban) (err error) {
	// Make sure a ban doesn't already exist for this IP
	if b, _ := s.ListBan(ctx, bson.M{"ip": ban.IP}); b != nil {
		err = fmt.Errorf("Error banning IP: %s is already banned", ban.IP)
		return err
	}
	_, err = s.srv.bans.InsertOne(ctx, ban)
	if err != nil {
		helpers.LogToSentry(err)
	}
	return err
}

func (s *banSrv) Unban(ctx context.Context, ip string) (err error) {
	_, err = s.srv.bans.DeleteOne(ctx, bson.M{"ip": ip})
	if err != nil {
		helpers.LogToSentry(err)
	}
	return err
}

func (s *banSrv) ListBan(ctx context.Context, query bson.M) (ban *models.Ban, err error) {
	result := s.srv.bans.FindOne(ctx, query)
	err = result.Err()
	if err != nil {
		if err != mongo.ErrNoDocuments {
			helpers.LogToSentry(err)
		}
		return nil, err
	}

	err = result.Decode(&ban)
	if err != nil {
		helpers.LogToSentry(err)
	}

	return ban, err
}

func (s *banSrv) ListBans(ctx context.Context, query bson.M, page, limit int, sort bson.M) (bans []*models.Ban, pages int, totalBans int64, err error) {
	pages = 1
	skip := 0
	count, err := s.srv.bans.CountDocuments(ctx, query)

	if err != nil {
		return bans, 0, 0, errorWrapper(err)
	}

	// If the count is greater than the perPage variable, then we have more than 1 page!
	if count > int64(limit) {
		pages = int(math.Ceil((float64(count) / float64(limit))))
		skip = (page - 1) * limit
	}

	cursor, err := s.srv.bans.Find(ctx, query, options.Find().SetSort(sort).SetSkip(int64(skip)).SetLimit(int64(limit)))
	if err != nil {
		return bans, pages, 0, errorWrapper(err)
	}

	err = cursor.All(ctx, &bans)

	return bans, pages, count, err
}
