package mongo

import (
	"context"
	"fmt"
	"math"

	"github.com/FM1337/flagbrew-website-source-code/pkg/helpers"
	"github.com/FM1337/flagbrew-website-source-code/pkg/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// logSrv satisfies the models.LogService interface.
type logSrv struct {
	srv *mongoSrv
}

func (s *mongoSrv) NewLogService() *logSrv {
	return &logSrv{srv: s}
}

func (s *logSrv) UpsertLog(ctx context.Context, l *interface{}) (err error) {
	_, err = s.srv.logs.InsertOne(ctx, l)
	if err != nil {
		helpers.LogToSentry(err)
	}
	return err
}

func (s *logSrv) ListLogs(ctx context.Context, query bson.M, page, limit int, sort bson.M) (logs interface{}, pages int, count int64, err error) {
	pages = 1
	skip := 0
	count, err = s.srv.logs.CountDocuments(ctx, query)

	if err != nil {
		return logs, 0, 0, errorWrapper(err)
	}

	// If the count is greater than the perPage variable, then we have more than 1 page!
	if count > int64(limit) {
		pages = int(math.Ceil((float64(count) / float64(limit))))
		skip = (page - 1) * limit
	}

	if query["log_type"] == "gpss_upload" || query["log_type"] == "gpss_bundle_upload" {
		query["db_version"] = 2
	}

	cursor, err := s.srv.logs.Find(ctx, query, options.Find().SetSort(sort).SetSkip(int64(skip)).SetLimit(int64(limit)))
	if err != nil {
		return logs, pages, 0, errorWrapper(err)
	}
	switch query["log_type"] {
	case "gpss_upload":
		logs = []*models.GPSSUploadLog{}
	case "gpss_failed_upload":
		logs = []*models.GPSSFailedUploadLog{}
	case "banned":
		logs = []*models.Ban{}
	case "setting_change":
		logs = []*models.SettingChangeLog{}
	case "gpss_deletion":
		logs = []*models.GPSSDeletionLog{}
	case "unban":
		logs = []*models.UnbanLog{}
	case "gpss_clean":
		logs = []*models.GPSSCleanLog{}
	case "build_delete":
		logs = []*models.PatreonBuildDeleteLog{}
	case "unrestrict":
		logs = []*models.UnrestrictedLog{}
	case "gpss_bundle_upload":
		logs = []*models.GPSSBundleUploadLog{}
	case "word_delete":
		logs = []*models.WordDeleteLog{}
	default:
		return nil, 0, 0, fmt.Errorf("Unknown log type %s", query["log_type"])
	}
	err = cursor.All(ctx, &logs)

	return logs, pages, count, err
}

func (s *logSrv) GetLog(ctx context.Context, query bson.M) (log interface{}, err error) {
	if query["log_type"] == "gpss_upload" || query["log_type"] == "gpss_bundle_upload" {
		query["db_version"] = 2
	}

	result := s.srv.logs.FindOne(ctx, query)
	if err != nil {
		return log, err
	}

	switch query["log_type"] {
	case "gpss_upload":
		log = &models.GPSSUploadLog{}
	case "gpss_failed_upload":
		log = &models.GPSSFailedUploadLog{}
	case "banned":
		log = &models.Ban{}
	case "setting_change":
		log = &models.SettingChangeLog{}
	case "gpss_deletion":
		log = &models.GPSSDeletionLog{}
	case "unban":
		log = &models.UnbanLog{}
	case "gpss_clean":
		log = &models.GPSSCleanLog{}
	case "build_delete":
		log = &models.PatreonBuildDeleteLog{}
	case "unrestrict":
		log = &models.UnrestrictedLog{}
	case "gpss_bundle_upload":
		log = &models.GPSSBundleUploadLog{}
	case "word_delete":
		log = &models.WordDeleteLog{}
	default:
		return nil, fmt.Errorf("Unknown log type %s", query["log_type"])
	}

	err = result.Decode(log)
	return log, err
}

func (s *logSrv) UpdateLog(ctx context.Context, query, update bson.M) (err error) {
	result := s.srv.logs.FindOneAndUpdate(ctx, query, update)
	err = result.Err()

	return err
}

// USED ONLY FOR EMERGERCIES
func (s *logSrv) DeleteLog(ctx context.Context, query bson.M) (err error) {
	_, err = s.srv.logs.DeleteOne(ctx, query)
	return err
}
