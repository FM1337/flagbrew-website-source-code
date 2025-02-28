package mongo

import (
	"context"
	"errors"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/FM1337/flagbrew-website-source-code/pkg/helpers"
	"github.com/FM1337/flagbrew-website-source-code/pkg/models"
	"github.com/chr4/pwgen"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// var migrating = false

type gpssSrv struct {
	srv    *mongoSrv
	logSrv models.LogService
}

func (s *mongoSrv) NewGPSSService(ls *models.LogService) *gpssSrv {
	return &gpssSrv{srv: s, logSrv: *ls}
}

func (s *gpssSrv) ListPokemon(ctx context.Context, code string, approvedOnly bool) (pokemon *models.GPSSPokemon, err error) {
	query := bson.M{"download_code": code, "deleted": false}
	if approvedOnly {
		query["approved"] = true
	}
	result := s.srv.gpss.FindOne(ctx, query)
	err = result.Err()
	if err != nil {
		return nil, err
	}
	result.Decode(&pokemon)
	return pokemon, errorWrapper(err)
}

func (s *gpssSrv) ListPokemons(ctx context.Context, query bson.M, page, limit int, sort bson.M, pksmMode bool) (pokemons []*models.GPSSPokemon, pages int, total int64, err error) {
	pages = 1
	skip := 0
	count, err := s.srv.gpss.CountDocuments(ctx, query)
	if err != nil {
		return pokemons, 0, 0, errorWrapper(err)
	}

	// If the count is greater than the perPage variable, then we have more than 1 page!
	if count > int64(limit) {
		pages = int(math.Ceil((float64(count) / float64(limit))))
		skip = (page - 1) * limit
	}

	o := options.Find()
	o.SetSort(sort).SetSkip(int64(skip)).SetLimit(int64(limit))
	if pksmMode {
		o.SetProjection(bson.M{"base_64": 1, "download_code": 1, "generation": 1, "pokemon.is_legal": 1, "_id": 0, "size": 1, "pokemon.dex_number": 1})
	}

	cursor, err := s.srv.gpss.Find(ctx, query, o)
	if err != nil {
		return pokemons, pages, 0, errorWrapper(err)
	}

	err = cursor.All(ctx, &pokemons)

	return pokemons, pages, count, errorWrapper(err)
}

func (s *gpssSrv) ListBundle(ctx context.Context, code string) (bundle *models.GPSSBundlePokemon, err error) {
	result := s.srv.bundles.FindOne(ctx, bson.M{"download_code": code})
	err = result.Err()
	result.Decode(&bundle)
	return bundle, errorWrapper(err)
}

func (s *gpssSrv) ListBundles(ctx context.Context, query bson.M, page, limit int, sort bson.M) (bundles []*models.GPSSBundlePokemon, pages int, count int64, err error) {
	pages = 1
	skip := 0

	count, err = s.srv.bundles.CountDocuments(ctx, query)
	if err != nil {
		return bundles, 0, 0, errorWrapper(err)
	}

	// If the count is greater than the perPage variable, then we have more than 1 page!
	if count > int64(limit) {
		pages = int(math.Ceil((float64(count) / float64(limit))))
		skip = (page - 1) * limit
	}

	cursor, err := s.srv.bundles.Find(ctx, query, options.Find().SetSort(sort).SetSkip(int64(skip)).SetLimit(int64(limit)))
	if err != nil {
		return bundles, pages, count, errorWrapper(err)
	}

	err = cursor.All(ctx, &bundles)

	return bundles, pages, count, errorWrapper(err)
}

func (s *gpssSrv) UpsertPokemon(ctx context.Context, gpssPokemon *models.GPSSPokemon, header *http.Header, patron bool, patronCode, patronDiscord string, bundleUpload bool, bundleCode string) (uploaded bool, approved bool, downloadCode string, err error) {
	var log interface{}
	// if check if the download_code is set
	if gpssPokemon.DownloadCode != "" {
		// if it's set, do an upsert/update
		_, err = s.srv.gpss.UpdateOne(ctx, bson.M{"_id": gpssPokemon.ID}, bson.M{"$set": gpssPokemon}, options.Update().SetUpsert(true))
		if err != nil {
			return false, false, "", errorWrapper(err)
		}
		return true, gpssPokemon.Approved, gpssPokemon.DownloadCode, errorWrapper(err)
	}

	// look up pokemon in the database with the same base64
	result := &models.GPSSPokemon{}
	if err = s.srv.gpss.FindOne(ctx, bson.M{"base_64": gpssPokemon.Base64, "deleted": false}, options.FindOne()).Decode(&result); err != nil {
		if err != mongo.ErrNoDocuments {
			// Failed to check for duplicates, let's log as failed upload
			log = helpers.GenerateFailedUploadLog(header.Get("IP-For-Logs"), header.Get("source"), header.Get("discord-user"),
				fmt.Sprintf("Unexpected database error when checking for duplicates, error information: %s",
					err.Error()), patron, patronCode, patronDiscord)
			err = s.logSrv.UpsertLog(ctx, &log)

			return false, false, "", errorWrapper(err)
		}
	} else {
		// Duplicate pokemon let's log as failed upload
		// TODO, what should I do if we fail to log here? Post to discord or sentry?
		log = helpers.GenerateFailedUploadLog(header.Get("IP-For-Logs"), header.Get("source"), header.Get("discord-user"),
			fmt.Sprintf("Pokemon already exists in database (%s)", result.DownloadCode), patron, patronCode, patronDiscord)
		err = s.logSrv.UpsertLog(ctx, &log)
		return false, result.Approved, result.DownloadCode, errorWrapper(err)
	}

	// If we hit here, then it's a new pokemon and we should go ahead and insert it.
	// Before continuing, set the download code
	// TODO check that download code is unique
	gpssPokemon.DownloadCode = pwgen.Num(10)

	// Now insert
	_, err = s.srv.gpss.InsertOne(ctx, gpssPokemon)
	if err != nil {
		// Failed to insert pokemon, let's log this as failed upload

		log = helpers.GenerateFailedUploadLog(header.Get("IP-For-Logs"), header.Get("source"), header.Get("discord-user"),
			fmt.Sprintf("Unexpected database error when inserting pokemon, error information: %s",
				err.Error()), patron, patronCode, patronDiscord)
		err = s.logSrv.UpsertLog(ctx, &log)
		return false, false, "", errorWrapper(err)
	}
	approvedBy := ""
	if gpssPokemon.Approved {
		approvedBy = "System"
	}
	log = helpers.GenerateUploadLog(header.Get("IP-For-Logs"), header.Get("source"), header.Get("discord-user"), false, gpssPokemon.Pokemon, gpssPokemon.Approved, approvedBy, patron, bundleUpload, gpssPokemon.DownloadCode, bundleCode, patronCode, patronDiscord) // TODO bundle code stuff
	err = s.logSrv.UpsertLog(ctx, &log)

	uploadType := "individual"
	if bundleUpload {
		uploadType = "bundle"
	}
	helpers.IncreaseUploads(uploadType, []string{strconv.Itoa(gpssPokemon.Generation)}, []string{gpssPokemon.Pokemon.Species}, []string{gpssPokemon.Pokemon.Gender}, []bool{gpssPokemon.Pokemon.IsLegal}, []bool{gpssPokemon.Pokemon.IsShiny}, []bool{gpssPokemon.Pokemon.IsEgg}, gpssPokemon.Patreon)
	return true, gpssPokemon.Approved, gpssPokemon.DownloadCode, errorWrapper(err)
}

// Clean this up later...
func (s *gpssSrv) UpsertBundle(ctx context.Context, bundle *models.GPSSBundlePokemon, patron bool, patronCode, patronDiscord, bundleCode string, header *http.Header) (uploaded bool, downloadCode string, err error) {
	var log interface{}
	// check to see if the bundle's download code is set
	if bundle.DownloadCode != "" {
		// if the code is set then we want to do an update
		_, err = s.srv.bundles.UpdateOne(ctx, bson.M{"_id": bundle.ID}, bson.M{"$set": bundle}, options.Update().SetUpsert(true))
		if err != nil {
			return false, "", errorWrapper(err)
		}
		return true, bundle.DownloadCode, errorWrapper(err)
	}

	result := &models.GPSSBundlePokemon{}
	if err = s.srv.bundles.FindOne(ctx, bson.M{"download_codes": bson.M{"$all": bundle.DownloadCodes, "$size": len(bundle.DownloadCodes)}}, options.FindOne()).Decode(&result); err != nil {
		if err != mongo.ErrNoDocuments {
			return false, "", errorWrapper(err)
		}
	} else {
		err = errors.New("Bundle already uploaded")
		return false, result.DownloadCode, errorWrapper(err)
	}

	// If we hit here, then it's a new bundle and we should go ahead and insert it.
	// Before continuing, set the download code
	// TODO check that download code is unique
	bundle.DownloadCode = bundleCode

	isLegal := true
	pks := []models.GPSSPKSMBundlePokemon{}
	// Check the legality status of all the pokemon in the bundle
	for _, code := range bundle.DownloadCodes {
		pkmn, err := s.ListPokemon(ctx, code, false)
		if err != nil {
			// log this error to sentry
			helpers.LogToSentry(err)
			// set isLegal to false
			isLegal = false
			// break out of loop
			break
		}

		// if no error, then let's check the legality
		if !pkmn.Pokemon.IsLegal && isLegal {
			isLegal = false
		}

		pks = append(pks, models.GPSSPKSMBundlePokemon{
			Base64:     pkmn.Base64,
			Legality:   pkmn.Pokemon.IsLegal,
			Generation: pkmn.Generation,
		})
	}
	// now set the bundle's legality status
	bundle.IsLegal = isLegal
	bundle.Pokemons = pks

	// Now insert
	_, err = s.srv.bundles.InsertOne(ctx, bundle)

	if err != nil {
		return false, "", errorWrapper(err)
	}
	// now update each pokemon in the bundle that already existed and didn't have "in_group" set to true
	_, err = s.srv.gpss.UpdateMany(ctx, bson.M{"download_code": bson.M{"$in": bundle.DownloadCodes}, "in_group": false}, bson.M{"$set": bson.M{"in_group": true}})
	if err != nil {
		return true, bundle.DownloadCode, errorWrapper(err)
	}

	// No errors? Good, let's generate the upsert log now.
	pkmns := []models.Pokemon{}

	pokemons, _, _, err := s.ListPokemons(ctx, bson.M{"download_code": bson.M{"$in": bundle.DownloadCodes}}, 1, 6, bson.M{}, false)
	if err != nil {
		// uh oh error problem
		return false, "", errorWrapper(err)
	}

	// okay no errors? good, let's loop through the pokemons and get the data we need
	for _, pk := range pokemons {
		pkmns = append(pkmns, pk.Pokemon)
	}

	log = helpers.GenerationBundleUpsertLog(header.Get("IP-For-Logs"), header.Get("source"), header.Get("discord-user"), bundle.Patreon, pkmns, bundleCode, patronCode, patronDiscord, bundle.DownloadCodes, bundle.Approved)
	err = s.logSrv.UpsertLog(ctx, &log)

	return true, bundle.DownloadCode, errorWrapper(err)
}

func (s *gpssSrv) RemoveBundle(ctx context.Context, downloadCode string) (err error) {
	// TODO: Once logging service is in place, update the logs to indicate deleted bundle

	// before deleting, get a copy of the download codes in this bundle
	tmpCopy := &models.GPSSBundlePokemon{}
	err = s.srv.bundles.FindOne(ctx, bson.M{"download_code": downloadCode}, options.FindOne().SetProjection(bson.M{"download_codes": 1})).Decode(&tmpCopy)
	if err != nil {
		return errorWrapper(err)
	}
	// loop through the array and search the database for each bundle that contains the download code we're on
	for _, code := range tmpCopy.DownloadCodes {
		amount, err := s.srv.bundles.CountDocuments(ctx, bson.M{"download_codes": code})
		if err != nil {
			return errorWrapper(err)
		}
		// if amount is 1, then we want to update that pokemon's 'InGroup' bool to false
		if amount == 1 {
			err = s.srv.gpss.FindOneAndUpdate(ctx, bson.M{"download_code": code}, bson.M{"$set": bson.M{"in_group": false}}).Err()
			if err != nil {
				return errorWrapper(err)
			}
		}
	}

	_, err = s.srv.bundles.DeleteOne(ctx, bson.M{"download_code": downloadCode})
	return errorWrapper(err)
}

func (s *gpssSrv) RemovePokemon(ctx context.Context, downloadCode string, rejecting, soft bool) (err error) {
	var log interface{}
	query := bson.M{"download_code": downloadCode, "deleted": false}
	if rejecting {
		query["approved"] = false
	}
	// BEFORE DOING ANYTHING, CHECK TO SEE IF THE THING EXISTS
	matches, err := s.srv.gpss.CountDocuments(ctx, query)
	if err != nil {
		return errorWrapper(err)
	}
	// if matches is 0, then we stop right here
	if matches == 0 {
		return errorWrapper(errors.New("pokemon does not exist"))
	}

	inBundles := false
	// first query the bundles collection for any bundles containing in their download_codes array the pokemon's download code
	matches, err = s.srv.bundles.CountDocuments(ctx, bson.M{"download_codes": downloadCode})
	if err != nil {
		return errorWrapper(err)
	}
	if matches > 0 {
		// if there are bundles with this pokemon, then we will need to remove this pokemon's download code from each of them
		inBundles = true
	}
	// if we do have to remove them from bundles, get this out of the way first
	if inBundles {
		// create a bundles slice to hold the bundles containing the pokemon
		bundles := []*models.GPSSBundlePokemon{}

		// fetch each bundle that has the code in it
		results, err := s.srv.bundles.Find(ctx, bson.M{"download_codes": downloadCode})
		if err != nil {
			return errorWrapper(err)
		}
		// decode results
		if err = results.All(ctx, &bundles); err != nil {
			return errorWrapper(err)
		}

		// loop through each bundle
		for _, bundle := range bundles {
			// if the bundle's length before deletion of the pokemon is two (or less which should be impossible but hey), we will delete the bundle
			if len(bundle.DownloadCodes) <= 2 {
				if err = s.RemoveBundle(ctx, bundle.DownloadCode); err != nil {
					return errorWrapper(err)
				}
				// upsert deletion log
				log = helpers.GenerateDeletionLog("System", "Bundle has 1 Pokemon Left and is not a bundle anymore", "Bundle", bundle.DownloadCode)
				if err = s.logSrv.UpsertLog(ctx, &log); err != nil {
					helpers.LogToSentry(err)
				}
			} else {
				// if it's greater than 2 (which it should be as bundles can't be less than 2) then remove it from the slice
				tmpBundleArr := []string{}
				tmpPkmnArray := []models.GPSSPKSMBundlePokemon{}
				approved := true
				for i, tmpCode := range bundle.DownloadCodes {
					// if the code isn't the one we're removing
					if tmpCode != downloadCode {
						// add it to the tmpArray we're replacing
						tmpBundleArr = append(tmpBundleArr, tmpCode)
						tmpPkmnArray = append(tmpPkmnArray, bundle.Pokemons[i])
					}
				}
				// loop through the codes and find out if there are any pokemon that aren't approved yet
				for _, dc := range tmpBundleArr {
					if !s.PokemonApproved(ctx, dc) {
						approved = false
					}
				}
				err = s.srv.bundles.FindOneAndUpdate(ctx, bson.M{"download_code": bundle.DownloadCode}, bson.M{"$set": bson.M{"download_codes": tmpBundleArr, "pokemons": tmpPkmnArray, "count": len(tmpBundleArr), "approved": approved}}).Err()
				if err != nil {
					return errorWrapper(err)
				}
			}
		}
	}

	// Now that the bundle updating/removing is out of the way, we can proceed to remove the pokemon
	if soft {
		_, err = s.srv.gpss.UpdateOne(ctx, bson.M{"download_code": downloadCode}, bson.M{"$set": bson.M{"deleted": true, "in_group": false}})
	} else {
		_, err = s.srv.gpss.DeleteOne(ctx, bson.M{"download_code": downloadCode})
	}
	return errorWrapper(err)
}

func (s *gpssSrv) DownloadPokemon(ctx context.Context, code string, approvedOnly bool) (pokemon *models.GPSSPokemon, err error) {
	query := bson.M{"download_code": code, "deleted": false}
	if approvedOnly {
		query["approved"] = true
	}
	result := s.srv.gpss.FindOneAndUpdate(ctx, query, bson.M{"$inc": bson.M{"current_downloads": 1, "lifetime_downloads": 1}})
	err = result.Err()
	result.Decode(&pokemon)
	return pokemon, errorWrapper(err)
}

func (s *gpssSrv) DownloadBundle(ctx context.Context, code string, approvedOnly bool) (bundledPkmn []*models.GPSSPokemon, err error) {
	query := bson.M{"download_code": code}
	if approvedOnly {
		query["approved"] = true
	}
	result := s.srv.bundles.FindOneAndUpdate(ctx, query, bson.M{"$inc": bson.M{"download_count": 1}})
	err = result.Err()
	if err != nil {
		return bundledPkmn, errorWrapper(err)
	}

	bundle := &models.GPSSBundlePokemon{}

	err = result.Decode(&bundle)
	if err != nil {
		return bundledPkmn, errorWrapper(err)
	}

	// Okay loop through each pokemon and increment by one
	for _, code := range bundle.DownloadCodes {
		pkmn, err := s.DownloadPokemon(ctx, code, approvedOnly)
		if err != nil {
			return bundledPkmn, errorWrapper(err)
		}
		bundledPkmn = append(bundledPkmn, pkmn)
	}
	// now return
	return bundledPkmn, errorWrapper(err)
}

func (s *gpssSrv) ResetOldPokemonDownloads(ctx context.Context) (modified int64, err error) {
	results, err := s.srv.gpss.UpdateMany(ctx, bson.M{"current_downloads": bson.M{"$gt": 0}, "last_reset": bson.M{"$lt": time.Now().Add(-43800 * time.Minute)}, "deleted": false}, bson.M{"$set": bson.M{"current_downloads": 0, "last_reset": time.Now()}})
	return results.ModifiedCount, err
}

func (s *gpssSrv) PokemonExists(ctx context.Context, base64 string) (exists bool, code string) {
	result := s.srv.gpss.FindOne(ctx, bson.M{"base_64": base64, "deleted": false})
	if result.Err() != nil {
		if result.Err() != mongo.ErrNoDocuments {
			// Unexpected error, let's log it!
			helpers.LogToSentry(result.Err())
		}
		return false, ""
	}
	pokemon := &models.GPSSPokemon{}

	err := result.Decode(&pokemon)
	if err != nil {
		// log error to sentry
		helpers.LogToSentry(err)
		return true, "COULD NOT GET DONWLOAD CODE" // one exists, but we couldn't get it
	}
	// No errors, pokemon must exist
	return true, pokemon.DownloadCode
}

func (s *gpssSrv) PokemonApproved(ctx context.Context, code string) (approved bool) {

	pkmn := &models.GPSSPokemon{}

	result := s.srv.gpss.FindOne(ctx, bson.M{"download_code": code, "deleted": false}, options.FindOne().SetProjection(bson.M{"approved": 1}))
	err := result.Err()
	if err != nil {
		if err != mongo.ErrNoDocuments {
			helpers.LogToSentry(err)
			return false
		}
	}
	err = result.Decode(&pkmn)
	if err != nil {
		helpers.LogToSentry(err)
		return false
	}
	return pkmn.Approved
}

func (s *gpssSrv) GetStats(ctx context.Context, approved bool) (pokemon, bundles int64, err error) {
	query := bson.M{"deleted": false}
	if approved {
		query["approved"] = true
	}

	pokemon, err = s.srv.gpss.CountDocuments(ctx, query)
	if err != nil {
		return 0, 0, err
	}

	delete(query, "deleted")
	bundles, err = s.srv.bundles.CountDocuments(ctx, query)
	if err != nil {
		return 0, 0, err
	}
	return pokemon, bundles, err
}

// // This is a migrate function to run all pokemon in the database against CoreAPI again.
// func (s *gpssSrv) NewMigrate() (err error) {
// 	if !migrating {
// 		startTime := time.Now()
// 		migrating = true

// 		cursor, err := s.srv.logs.Find(context.Background(), bson.M{"log_type": "gpss_upload"})
// 		if err != nil {
// 			fmt.Println("495", err)
// 			return err
// 		}

// 		data := []models.GPSSUploadLog{}
// 		err = cursor.All(context.Background(), &data)
// 		if err != nil {
// 			fmt.Println("502", err)
// 			return err
// 		}

// 		sem := make(chan struct{}, 10)

// 		for i, pokemon := range data {
// 			sem <- struct{}{}
// 			go func(log models.GPSSUploadLog, i int) {
// 				defer func() { <-sem }() // receiving from the channel unblocks it

// 				result := s.srv.gpss.FindOne(context.Background(), bson.M{"download_code": log.DownloadCode})
// 				if result.Err() != nil {
// 					fmt.Println("515", err)
// 					return
// 				}

// 				pkmn := models.NewGPSSPokemon{}
// 				err = result.Decode(&pkmn)

// 				if err != nil {
// 					fmt.Println("523", err)
// 					return
// 				}

// 				newLog := models.NewGPSSUploadLog{
// 					ID:              log.ID,
// 					Date:            log.Date,
// 					UploaderIP:      log.UploaderIP,
// 					UploadSource:    log.UploadSource,
// 					UploaderDiscord: log.UploaderDiscord,
// 					Deleted:         log.Deleted,
// 					PokemonData:     pkmn.Pokemon,
// 					Patron:          log.Patron,
// 					PatronCode:      log.PatronCode,
// 					PatronDiscord:   log.PatronDiscord,
// 					DownloadCode:    log.DownloadCode,
// 					Approved:        log.Approved,
// 					ApprovedBy:      log.ApprovedBy,
// 					Rejected:        log.Rejected,
// 					RejectedBy:      log.RejectedBy,
// 					RejectedReason:  log.RejectedReason,
// 					BundleUpload:    log.BundleUpload,
// 					BundleCode:      log.BundleCode,
// 					LogType:         log.LogType,
// 					DBVersion:       2,
// 				}

// 				_, err = s.srv.logs.UpdateByID(context.Background(), log.ID, bson.M{"$set": newLog})
// 				if err != nil {
// 					fmt.Println(err)
// 					return
// 				}

// 			}(pokemon, i)
// 			fmt.Printf("\rMigrated %d/%d", i, len(data))

// 			// formData := make(map[string]string)
// 			// pkmnFile, err := base64.StdEncoding.DecodeString(pokemon.Base64)
// 			// if err != nil {
// 			// 	// log error to sentry
// 			// 	fmt.Println(err)
// 			// 	continue
// 			// }
// 			// // Get the pokemon from CoreAPI
// 			// returned, _, err := helpers.CoreAPIFile(pkmnFile, formData, "/info")
// 			// if err != nil {
// 			// 	// log error to sentry
// 			// 	fmt.Println(err)
// 			// 	continue
// 			// }

// 			// pkmn := models.NewPokemon{}

// 			// err = json.Unmarshal(returned, &pkmn)
// 			// if err != nil {
// 			// 	// log error to sentry
// 			// 	fmt.Println(err)
// 			// 	continue
// 			// }

// 			// size := 0
// 			// if pkmn.PartySize > pkmn.StoredSize {
// 			// 	size = int(pkmn.PartySize)
// 			// } else {
// 			// 	size = int(pkmn.StoredSize)
// 			// }

// 			// // Update the pokemon in the database
// 			// // Use the new GPSS Model
// 			// newGPSS := &models.NewGPSSPokemon{
// 			// 	ID:                pokemon.ID,
// 			// 	Base64:            pokemon.Base64,
// 			// 	DownloadCode:      pokemon.DownloadCode,
// 			// 	Pokemon:           pkmn,
// 			// 	Patreon:           pokemon.Patreon,
// 			// 	InGroup:           pokemon.InGroup,
// 			// 	Size:              size,
// 			// 	Generation:        int(pkmn.Generation),
// 			// 	LifetimeDownloads: pokemon.LifetimeDownloads,
// 			// 	CurrentDownloads:  pokemon.CurrentDownloads,
// 			// 	DBVersion:         5,
// 			// 	UploadDate:        pokemon.UploadDate,
// 			// 	LastReset:         pokemon.LastReset,
// 			// 	Approved:          pokemon.Approved,
// 			// 	Deleted:           pokemon.Deleted,
// 			// }

// 			// _, err = s.srv.gpss.UpdateByID(context.Background(), pokemon.ID, bson.M{"$set": newGPSS})
// 			// if err != nil {
// 			// 	// log error to sentry
// 			// 	fmt.Println(err)
// 			// 	continue
// 			// }

// 		}

// 		for i := 0; i < cap(sem); i++ {
// 			sem <- struct{}{}
// 		}

// 		// fmt.Printf("\rMigrated %d/%d", i, len(data))
// 		fmt.Println("Migrated all pokemon in", time.Since(startTime))
// 		return nil
// 	}
// 	return fmt.Errorf("already migrating")
// }

// // This is a function that will be removed in the future once the site goes live and the database is migrated successfully
// func (s *gpssSrv) Migrate(logData, pokemonData []byte, header *http.Header) (err error) {
// 	if !migrating {
// 		startTime := time.Now()
// 		migrating = true
// 		// Create our indexes
// 		_, err = s.srv.gpss.Indexes().CreateMany(context.Background(), []mongo.IndexModel{
// 			{
// 				Keys: bson.M{"upload_date": 1},
// 			},
// 			{
// 				Keys: bson.M{"pokemon.is_legal": 1},
// 			},
// 			{
// 				Keys: bson.M{"lifetime_downloads": 1},
// 			},
// 		})
// 		if err != nil {
// 			return err
// 		}

// 		_, err = s.srv.bundles.Indexes().CreateMany(context.Background(), []mongo.IndexModel{
// 			{
// 				Keys: bson.M{"upload_date": 1},
// 			},
// 			{
// 				Keys: bson.M{"is_legal": 1},
// 			},
// 			{
// 				Keys: bson.M{"download_count": 1},
// 			},
// 		})

// 		if err != nil {
// 			return err
// 		}

// 		logs := []*models.GPSSMigrateLog{}
// 		err = json.Unmarshal(logData, &logs)

// 		if err != nil {
// 			return err
// 		}

// 		pokemons := []*models.GPSSMigratePokemon{}
// 		err = json.Unmarshal(pokemonData, &pokemons)

// 		if err != nil {
// 			return err
// 		}

// 		if err != nil {
// 			return err
// 		}
// 		lMap := make(map[string]models.GPSSMigrateLog)
// 		for _, log := range logs {
// 			lMap[log.DownloadCode] = *log
// 		}

// 		failed := []string{}
// 		for _, pk := range pokemons {
// 			// Make the form data
// 			formData := make(map[string]string)
// 			formData["generation"] = pk.Generation

// 			// decode the base64 from the pokemon
// 			pkmnFile, err := base64.StdEncoding.DecodeString(pk.Base64)
// 			if err != nil {
// 				failed = append(failed, pk.DownloadCode)
// 				continue
// 			}
// 			// call CoreAPI
// 			data, _, err := helpers.CoreAPIFile(pkmnFile, formData, "/PokemonInfo")
// 			if err != nil {
// 				failed = append(failed, pk.DownloadCode)
// 				continue
// 			}
// 			if helpers.IsArugmentError(errors.New(strings.Split(string(data), "\n")[0])) {
// 				// Try without generation
// 				data, _, err := helpers.CoreAPIFile(pkmnFile, nil, "/PokemonInfo")
// 				if err != nil {
// 					failed = append(failed, pk.DownloadCode)
// 					continue
// 				}
// 				if helpers.IsArugmentError(errors.New(strings.Split(string(data), "\n")[0])) {
// 					failed = append(failed, pk.DownloadCode)
// 					continue
// 				}
// 			}

// 			// create the pokemon object
// 			pkmn := &models.Pokemon{}

// 			err = json.Unmarshal(data, &pkmn)
// 			if err != nil {
// 				failed = append(failed, pk.DownloadCode)
// 				continue
// 			}

// 			// generate the GPSS insert
// 			gpssInsert := &models.GPSSPokemon{
// 				Base64:            pk.Base64,
// 				Pokemon:           *pkmn,
// 				Size:              int(pkmn.Size),
// 				UploadDate:        lMap[pk.DownloadCode].UploadDate.DateTime,
// 				LastReset:         time.Now(),
// 				Generation:        pk.Generation,
// 				LifetimeDownloads: pk.TotalDownloads,
// 				CurrentDownloads:  0,
// 				DBVersion:         3,
// 				Patreon:           lMap[pk.DownloadCode].Patreon,
// 				Approved:          true,
// 				InGroup:           len(pk.GroupCodes) != 0,
// 				Deleted:           false,
// 			}
// 			// Set the IP for the header.
// 			header.Set("IP-For-Logs", lMap[pk.DownloadCode].IP)
// 			header.Set("source", "Migration From Original GPSS")

// 			patreonCode := ""
// 			if gpssInsert.Patreon {
// 				patreonCode = "UNKNOWN"
// 			}
// 			// insert into the database
// 			_, _, _, err = s.UpsertPokemon(context.Background(), gpssInsert, header, false, patreonCode, "", false, "")
// 			if err != nil {
// 				failed = append(failed, pk.DownloadCode)
// 				continue
// 			}
// 		}
// 		runtime := time.Since(startTime)
// 		success, context := helpers.GenerateSentryEventLogContext([]string{"duration", "failed_upload", "amount_to_be_migrated"}, []interface{}{runtime, failed, len(pokemons)})
// 		if success {
// 			helpers.LogToSentryWithContext(sentry.LevelInfo, "migration of GPSS 1 to GPSS 2 is complete", context)
// 		}
// 		return err
// 	}
// 	return fmt.Errorf("already migrating")
// }

// func (s *gpssSrv) ListGenerationDownloads(ctx context.Context) (downloads map[string]float64, err error) {
// 	// db.gpss.aggregate([
// 	// 	{
// 	// 		$group: {
// 	// 			_id: '$generation',
// 	// 			count: { $sum: '$lifetime_downloads' }
// 	// 		}
// 	// 	}
// 	// ])

// 	// db.gpss.aggregate([
// 	// 	{$match: {'pokemon.is_shiny': true}},
// 	// 	{
// 	// 		$group: {
// 	// 			_id: '$generation',
// 	// 			count: { $sum: 1 }
// 	// 		}
// 	// 	}
// 	// ])

// 	// db.gpss.aggregate([
// 	// 	{ "$group": { "_id": "$pokemon.is_shiny", "count": { "$sum": 1 } } }
// 	//   ])

// 	// pokemon.is_shiny
// }

func (s *gpssSrv) ListCountForFieldStat(ctx context.Context, field string, downloads bool) (uploads map[string]float64, err error) {
	sum := bson.D{{"$sum", 1}}
	if downloads {
		sum = bson.D{{"$sum", "$lifetime_downloads"}}
	}

	result, err := s.srv.gpss.Aggregate(ctx, mongo.Pipeline{
		{
			{"$group", bson.D{
				{"_id", field},
				{"count", sum}},
			}},
	},
	)
	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}
	results := []map[string]interface{}{}

	if err = result.All(ctx, &results); err != nil {
		fmt.Println(err.Error())
		return nil, err
	}
	uploads = make(map[string]float64)
	for _, gen := range results {
		var key string

		switch v := gen["_id"].(type) {
		case bool:
			key = strconv.FormatBool(v)
		case string:
			key = v
		default:
			key = fmt.Sprintf("%v", v)
		}

		uploads[key] = float64(gen["count"].(int32))
	}
	return uploads, err
}

// func (s *gpssSrv) ListGenerationUploads(ctx context.Context) (uploads map[string]float64, err error) {
// 	result, err := s.srv.gpss.Aggregate(ctx, mongo.Pipeline{
// 		{
// 			{"$group", bson.D{
// 				{"_id", "$generation"},
// 				{"count", bson.D{
// 					{"$sum", 1},
// 				}},
// 			}},
// 		},
// 	})
// 	if err != nil {
// 		fmt.Println(err.Error())
// 		return nil, err
// 	}
// 	results := []map[string]interface{}{}

// 	if err = result.All(ctx, &results); err != nil {
// 		fmt.Println(err.Error())
// 		return nil, err
// 	}
// 	uploads = make(map[string]float64)
// 	for _, gen := range results {
// 		uploads[gen["_id"].(string)] = float64(gen["count"].(int32))
// 	}
// 	return uploads, err
// }

func (s *gpssSrv) RandomPokemon(ctx context.Context, amount int, generations []string) ([]*models.GPSSRandomPokemon, error) {

	// Get sample of 6 pokemon that are only approved and not deleted
	result, err := s.srv.gpss.Aggregate(ctx, mongo.Pipeline{
		{
			{"$match", bson.D{{"$and", []bson.D{
				{{"approved", true}},
				{{"deleted", false}},
				{{"generation", bson.D{{"$in", generations}}}},
			}}}}},
		{
			{"$sample", bson.D{{"size", amount}}},
		},
	})

	if err != nil {
		return nil, err
	}

	results := []models.GPSSPokemon{}
	if err = result.All(ctx, &results); err != nil {
		return nil, err
	}
	var pkmns []*models.GPSSRandomPokemon
	for _, p := range results {
		// Append the base64 encoded pokemon bytes and generation to the list
		pkmns = append(pkmns, &models.GPSSRandomPokemon{
			Base64:     p.Base64,
			Generation: p.Generation,
		})
		// Increase the download count for this pokemon (call the download function)
		s.DownloadPokemon(ctx, p.DownloadCode, true) // We don't actually care about grabbing the pokemon, we're just using this function to increase the download count because I'm too lazy to duplicate code. :P

	}

	return pkmns, nil
}
