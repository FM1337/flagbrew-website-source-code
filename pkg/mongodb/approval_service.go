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

type approvalSvc struct {
	srv     *mongoSrv
	logSrv  models.LogService
	gpssSrv models.GPSSService
}

type restrictService struct {
	srv    *mongoSrv
	logSrv models.LogService
}

func (s *mongoSrv) NewApprovalSvc(ls *models.LogService, gpss *models.GPSSService) (*approvalSvc, *restrictService) {
	return &approvalSvc{srv: s, logSrv: *ls, gpssSrv: *gpss}, &restrictService{srv: s, logSrv: *ls}
}

func (s *approvalSvc) Approve(ctx context.Context, code, user string) (err error) {
	// approve the pokemon
	result := s.srv.gpss.FindOneAndUpdate(ctx, bson.M{"download_code": code, "approved": false, "deleted": false}, bson.M{"$set": bson.M{"approved": true}})
	if result.Err() != nil {
		return result.Err()
	}

	tmpPkmn := &models.GPSSPokemon{}
	err = result.Decode(&tmpPkmn)

	if err != nil {
		return err
	}

	// Check if pokemon in group
	if tmpPkmn.InGroup {
		// Oh great, now we have to get every bundle that has this pokemon in it and isn't approved
		hasBundles := true
		results, err := s.srv.bundles.Find(ctx, bson.M{"download_codes": code, "approved": false})
		if err != nil {
			if err != mongo.ErrNoDocuments {
				return err
			}
			hasBundles = false
		}

		if hasBundles {
			// decode
			bundles := []*models.GPSSBundlePokemon{}
			err = results.All(ctx, &bundles)
			if err != nil {
				return err
			}

			// loop through each bundle's download codes
			for _, bundle := range bundles {
				approved := true
				for _, dc := range bundle.DownloadCodes {
					if dc == code {
						// skip this one as it is already approved
						continue
					}
					if !s.gpssSrv.PokemonApproved(ctx, dc) {
						// pokemon ain't approved, which means we can't set this bundle to approved just yet.
						approved = false
					}
				}
				if approved {
					// update the bundle to be approved
					result := s.srv.bundles.FindOneAndUpdate(ctx, bson.M{"download_code": bundle.DownloadCode}, bson.M{"$set": bson.M{"approved": true}})
					if result.Err() != nil {
						return result.Err()
					}
					// update the log for it
					err = s.logSrv.UpdateLog(ctx, bson.M{"log_type": "gpss_bundle_upload", "download_code": bundle.DownloadCode}, bson.M{"$set": bson.M{"approved": true}})
					if err != nil {
						return err
					}
				}
			}
		}
	}

	// if no errors, then update the log
	err = s.logSrv.UpdateLog(ctx, bson.M{"log_type": "gpss_upload", "download_code": code}, bson.M{"$set": bson.M{"approved": true, "approved_by": user}})
	return err
}

func (s *approvalSvc) ApproveBundle(ctx context.Context, code, user string) (err error) {
	// approve the bundle
	result := s.srv.bundles.FindOneAndUpdate(ctx, bson.M{"download_code": code}, bson.M{"$set": bson.M{"approved": true}})
	if result.Err() != nil {
		return err
	}

	// okay no errors? Good let's approve all the pokemon in the bundle
	bundle := &models.GPSSBundlePokemon{}

	err = result.Decode(&bundle)
	if err != nil {
		return err
	}

	for _, dc := range bundle.DownloadCodes {
		// call the approve pokemon function
		err = s.Approve(ctx, dc, user)
		if err != nil {
			return err
		}
	}

	// Okay no errors good.
	// Time to upsert a log
	err = s.logSrv.UpdateLog(ctx, bson.M{"log_type": "gpss_bundle_upload", "download_code": code}, bson.M{"$set": bson.M{"approved": true}})
	return err
}

func (s *approvalSvc) Reject(ctx context.Context, code, reason, user string) (err error) {
	// reject the pokemon (aka delete it)
	err = s.gpssSrv.RemovePokemon(ctx, code, true, false)
	if err != nil {
		if err != mongo.ErrNoDocuments {
			return err
		}
	}
	// if no unexpected errors, then update the log
	err = s.logSrv.UpdateLog(ctx, bson.M{"log_type": "gpss_upload", "download_code": code}, bson.M{"$set": bson.M{"rejected": true, "rejected_by": user, "rejected_reason": reason}})
	return err
}

func (s *approvalSvc) ListPending(ctx context.Context, query bson.M, page, limit int, sort bson.M) (pending []*models.GPSSPokemon, pages int, count int64, err error) {
	return s.gpssSrv.ListPokemons(ctx, query, page, limit, sort, false)
}

func (s *restrictService) IsUploaderRestricted(ctx context.Context, ip string) (bool, string) {
	result := s.srv.restricted.FindOne(ctx, bson.M{"ip": ip})
	if result.Err() != nil {
		if result.Err() != mongo.ErrNoDocuments {
			helpers.LogToSentry(result.Err())
		}
		return false, ""
	}
	tmp := &models.RestrictedUploader{}
	err := result.Decode(&tmp)
	if err != nil {
		helpers.LogToSentry(err)
		return true, "Couldn't get reason, see error logs"
	}
	return true, tmp.RestrictedReason
}

func (s *restrictService) RestrictUploader(ctx context.Context, ip, reason, user string) (err error) {
	// Check to make sure the IP isn't already restricted (we'll do this frontside as well)
	if restricted, _ := s.IsUploaderRestricted(ctx, ip); restricted {
		return fmt.Errorf("%s already is restricted", ip)
	}
	// Okay they aren't restricted, let's restrict them
	_, err = s.srv.restricted.InsertOne(ctx, &models.RestrictedUploader{
		IP:               ip,
		RestrictedReason: reason,
		RestrictedBy:     user,
	})
	return err
}

func (s *restrictService) UnrestrictUploader(ctx context.Context, ip, user string) (err error) {
	// Check to make sure the IP isn't already unrestricted (we'll do this frontside as well)
	if restricted, _ := s.IsUploaderRestricted(ctx, ip); !restricted {
		return fmt.Errorf("%s is not restricted", ip)
	}

	// Okay they are restricted, let's get their restricted data and then unrestrict them
	restrict := &models.RestrictedUploader{}
	s.srv.restricted.FindOne(ctx, bson.M{"ip": ip}).Decode(&restrict)

	_, err = s.srv.restricted.DeleteOne(ctx, bson.M{"ip": ip})
	if err != nil {
		return err
	}
	// Okay they're unrestricted, let's log who the hell did it.
	var log interface{}
	log = helpers.GenerateUnrestrictLog(user, restrict)
	err = s.logSrv.UpsertLog(ctx, &log)
	return err
}

// TODO: can we deduplicate this (and the others) code so that everytime we need to make a request with those kind of params, we can just call a single function?
func (s *restrictService) ListRestricted(ctx context.Context, query bson.M, page, limit int, sort bson.M) (restricted []*models.RestrictedUploader, pages int, count int64, err error) {
	pages = 1
	skip := 0
	count, err = s.srv.logs.CountDocuments(ctx, query)

	if err != nil {
		return restricted, 0, 0, errorWrapper(err)
	}

	// If the count is greater than the perPage variable, then we have more than 1 page!
	if count > int64(limit) {
		pages = int(math.Ceil((float64(count) / float64(limit))))
		skip = (page - 1) * limit
	}

	cursor, err := s.srv.restricted.Find(ctx, query, options.Find().SetSort(sort).SetSkip(int64(skip)).SetLimit(int64(limit)))
	if err != nil {
		return restricted, pages, 0, errorWrapper(err)
	}

	err = cursor.All(ctx, &restricted)

	return restricted, pages, count, err
}
