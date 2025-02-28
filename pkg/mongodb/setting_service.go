package mongo

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/FM1337/flagbrew-website-source-code/pkg/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// settingSrv satisfies the models.SettingService interface.
type settingSrv struct {
	srv *mongoSrv
}

func (s *mongoSrv) NewSettingService() *settingSrv {
	return &settingSrv{srv: s}
}

func (s *settingSrv) InsertSetting(ctx context.Context, setting *models.Setting) (err error) {
	_, err = s.srv.settings.InsertOne(ctx, setting)
	return err
}

func (s *settingSrv) UpdateSetting(ctx context.Context, name string, value interface{}) (err error) {
	_, err = s.srv.settings.UpdateOne(ctx, bson.M{"map_key": name}, bson.M{"$set": bson.M{"value": value, "modified_date": time.Now()}})
	return err
}

func (s *settingSrv) DeleteSetting(ctx context.Context, name string) (err error) {
	_, err = s.srv.settings.DeleteOne(ctx, bson.M{"name": name, "system_variable": false})
	if err != nil && err == mongo.ErrNoDocuments {
		err = fmt.Errorf("Either no setting for that name exists, or you tried to delete a system variable")
	}
	return err
}

func (s *settingSrv) ListSettings(ctx context.Context, query bson.M, page, limit int, sort bson.M) (settings []*models.Setting, pages int, total int64, err error) {
	pages = 1
	skip := 0
	count, err := s.srv.settings.CountDocuments(ctx, query)

	if err != nil {
		return settings, 0, 0, errorWrapper(err)
	}

	// If the count is greater than the perPage variable, then we have more than 1 page!
	if count > int64(limit) {
		pages = int(math.Ceil((float64(count) / float64(limit))))
		skip = (page - 1) * limit
	}

	cursor, err := s.srv.settings.Find(ctx, query, options.Find().SetSort(sort).SetSkip(int64(skip)).SetLimit(int64(limit)))
	if err != nil {
		return settings, pages, 0, errorWrapper(err)
	}

	err = cursor.All(ctx, &settings)

	return settings, pages, count, err
}

func (s *settingSrv) LoadDefaults(ctx context.Context) (err error) {
	// This is dumb, Go why you make me do this?
	var settings []interface{}
	for _, setting := range models.DefaultSettings {
		settings = append(settings, setting)
	}
	_, err = s.srv.settings.InsertMany(ctx, settings)
	return err
}
