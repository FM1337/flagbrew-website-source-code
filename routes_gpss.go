package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/FM1337/flagbrew-website-source-code/pkg/helpers"
	"github.com/FM1337/flagbrew-website-source-code/pkg/models"
	"github.com/chr4/pwgen"
	"github.com/getsentry/sentry-go"
	"github.com/go-chi/chi/v5"
	"github.com/lrstanley/pt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func gpssDisabled(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var log interface{}
		var err error
		// get patreon code
		patreonCode := r.Header.Get("patreon")
		patreon := false
		patreonDiscordID := ""
		if patreonCode != "" && len(patreonCode) == 22 {
			patreon = svcPatron.IsPatron(r.Context(), patreonCode)
			if patreon {
				patreonDiscordID, err = svcPatron.GetPatronDiscord(r.Context(), patreonCode)
				if err != nil {
					// Can't die here, so just log what we can
					helpers.LogToSentry(err)
					patreonDiscordID = "UNKNOWN"
				}
			}
		} else {
			patreonCode = ""
		}

		ip := helpers.GetIP(r, legacyKey)
		if ip == "" {
			log = helpers.GenerateFailedUploadLog("", r.Header.Get("source"), r.Header.Get("discord-user"),
				fmt.Sprintf("Failed to get Uploader IP"), patreon, patreonCode, patreonDiscordID)
			svcLogs.UpsertLog(r.Context(), &log)
			helpers.HttpError(w, r, http.StatusInternalServerError, fmt.Errorf("Networking error, please try again later"), false, true, false)
			return
		}

		switch true {
		case strings.Contains(r.URL.Path, "bundle"):
			if !loadedSettings["bundles_support"].Value.(bool) {
				helpers.HttpError(w, r, http.StatusServiceUnavailable, fmt.Errorf("GPSS Bundle support is currently disabled"), false, false, false)
				return
			}
			fallthrough
		case strings.Contains(r.URL.Path, "upload"):
			if strings.Contains(r.URL.Path, "upload") {
				// Check if GPSS Uploading is enabled
				if !loadedSettings["gpss_upload_enabled"].Value.(bool) {
					log = helpers.GenerateFailedUploadLog(ip, r.Header.Get("source"), r.Header.Get("discord-user"),
						fmt.Sprintf("GPSS Uploading was diabled at the time of the upload attempt"), patreon, patreonCode, patreonDiscordID)
					svcLogs.UpsertLog(r.Context(), &log)
					helpers.HttpError(w, r, http.StatusServiceUnavailable, fmt.Errorf("GPSS Uploading currently disabled, try again later"), false, false, false)
					return
				}
			}
			fallthrough
		case strings.Contains(r.URL.Path, "download"):
			if strings.Contains(r.URL.Path, "download") {
				// Check if GPSS Downloading is enabled
				if !loadedSettings["gpss_download_enabled"].Value.(bool) {
					helpers.HttpError(w, r, http.StatusServiceUnavailable, fmt.Errorf("Downloading from GPSS is currently disabled, please try again later"), false, false, false)
					return
				}
			}
		}

		next.ServeHTTP(w, r)
	})
}

// proccesses a Pokemon and inserts it into the database.
func processPokemon(file multipart.File, generation, source, ip string, r *http.Request, bundleUpload bool, bundleCode string) (uploaded, approved bool, code string, logToSentry bool, b64 string, err error) {
	if source == "" {
		return false, false, "", true, "", fmt.Errorf("Upload source is missing")
	}

	pkmn := bytes.Buffer{}
	io.Copy(&pkmn, file)

	b64 = base64.StdEncoding.WithPadding('=').EncodeToString(pkmn.Bytes())
	// Check if Pokemon already exists
	exists, code := svcGPSS.PokemonExists(r.Context(), b64)
	if exists {
		return false, svcGPSS.PokemonApproved(r.Context(), code), code, false, b64, fmt.Errorf("Your Pokemon is already uploaded")
	}

	formData := make(map[string]string)
	intGen := 0
	if generation != "" {
		switch generation {
		case "6":
			intGen = 6
		case "7":
			intGen = 7
		case "8":
			intGen = 8
		case "BDSP":
			intGen = 8
		case "LGPE":
			intGen = 8
		default:
			num, err := strconv.Atoi(generation)
			if err != nil {
				return false, false, "", true, "", fmt.Errorf("Invalid generation")
			}
			intGen = num
		}

	}

	data, err := apiGPSS.GetPokemonInfo(pkmn.Bytes(), formData)
	if err != nil {
		return false, false, "", true, "", fmt.Errorf("Failed to get pkmn info from CoreAPI, error details: %s", err.Error())
	}

	patreon, patreonCode, patreonDiscordID := getPatronInfo(r)
	size := int64(0)

	if data.PartySize > data.StoredSize {
		size = data.PartySize
	} else {
		size = data.StoredSize
	}

	pokemon := models.GPSSPokemon{}
	pokemon.Pokemon = *data
	pokemon.Base64 = b64
	pokemon.UploadDate = time.Now()
	pokemon.LastReset = pokemon.UploadDate
	pokemon.DBVersion = 4
	if intGen != 0 {
		pokemon.Generation = intGen
	} else {
		switch size {
		case 44:
			pokemon.Generation = 1
		case 48:
			pokemon.Generation = 2
		case 100:
			pokemon.Generation = 3
		case 236:
			pokemon.Generation = 4
		case 220:
			pokemon.Generation = 5
		case 260:
			// check dex number
			if pokemon.Pokemon.DexNumber > 721 {
				pokemon.Generation = 7
			} else {
				pokemon.Generation = 6
			}
		default:
			pokemon.Generation = pokemon.Pokemon.Generation
		}
	}
	pokemon.InGroup = false
	pokemon.Size = int(size)
	pokemon.Patreon = patreon
	pokemon.Deleted = false

	// Check to see if word filter is tripped
	lookup := []string{pokemon.Pokemon.HandlingTrainer, pokemon.Pokemon.Ot, pokemon.Pokemon.Nickname}
	match, strict, index, err := svcFilter.CheckWords(r.Context(), lookup, "any", true)
	if err != nil {
		return false, false, "", true, "", err
	}

	if match && strict {
		return false, false, "", true, "", fmt.Errorf("%s was matched by our word filter with a strictly banned word and thus was rejected automatically", lookup[index])
	}

	restrictCheck, _ := svcRestrict.IsUploaderRestricted(r.Context(), ip)
	successfulLookup, isBad := helpers.ProfanityLookup(cli.Flags.API.Neutrino.Key, lookup)
	switch true {
	case isBad:
	case !successfulLookup:
	case loadedSettings["gpss_restrict_enabled"].Value.(bool):
	case match:
	case restrictCheck:
		_, context := helpers.GenerateSentryEventLogContext([]string{"code", "pokemon", "ip", "source", "pending_reason"}, []interface{}{
			pokemon.DownloadCode,
			pokemon,
			ip,
			source,
			[]interface{}{
				fmt.Sprintf("isBad: %v", isBad),
				fmt.Sprintf("!successfulLookup: %v", !successfulLookup),
				fmt.Sprintf("loadedSettings['gpss_restrict_enabled']: %v", loadedSettings["gpss_restrict_enabled"].Value.(bool)),
				fmt.Sprintf("match: %v", match),
				fmt.Sprintf("restrictCheck: %v", restrictCheck),
			},
		})
		helpers.LogToSentryWithContext(sentry.LevelWarning, "a pokemon requires manual approval", context)
		pokemon.Approved = false
	default:
		pokemon.Approved = true
	}

	r.Header.Set("IP-For-Logs", ip)
	uploaded, approved, code, err = svcGPSS.UpsertPokemon(r.Context(), &pokemon, &r.Header, patreon, patreonCode, patreonDiscordID, bundleUpload, bundleCode)
	if !uploaded && err == nil {
		// we should never hit here, so we'll log this one
		helpers.LogToSentry(fmt.Errorf("pokemon is already uploaded, but the earlier check was able to be bypassed, how? Pokemon: %v, IP: %s", pokemon, ip))
		err = fmt.Errorf("your Pokemon is already uploaded")
	}

	if err == nil && approved && !bundleUpload {
		go helpers.DiscordPostGPSS("individual", cli.Flags.WebHooks.Discord, cli.Flags.SiteURL, pokemon)
	}

	return uploaded, approved, code, false, b64, err
}

func getPatronInfo(r *http.Request) (patreon bool, patreonCode, patreonDiscordID string) {
	var err error
	patreonCode = r.Header.Get("patreon")
	patreon = false
	patreonDiscordID = ""
	if patreonCode != "" && len(patreonCode) == 22 {
		patreon = svcPatron.IsPatron(r.Context(), patreonCode)
		if patreon {
			patreonDiscordID, err = svcPatron.GetPatronDiscord(r.Context(), patreonCode)
			if err != nil {
				// Can't die here, so just log what we can
				helpers.LogToSentry(err)
				patreonDiscordID = "UNKNOWN"
			}
		}
	} else {
		patreonCode = ""
	}

	return patreon, patreonCode, patreonDiscordID
}

func registerGPSSRoutes(r chi.Router) {
	r.Get("/stats", func(w http.ResponseWriter, r *http.Request) {
		approvedOnly := false
		if _, err := getUserInfo(r); err != nil {
			approvedOnly = true
		}
		pokemon, bundles, err := svcGPSS.GetStats(r.Context(), approvedOnly)
		if err != nil {
			// Set up Context
			success, context := helpers.GenerateSentryEventLogContext([]string{"ip", "path", "headers", "error"}, []interface{}{helpers.GetIP(r, legacyKey), r.URL.Path, r.Header, err})
			// log with context
			if success {
				helpers.LogToSentryWithContext(sentry.LevelWarning, "An error was encountered when getting gpss stats from the database", context)
			}
			helpers.HttpError(w, r, http.StatusInternalServerError, err, false, true, true)
			return
		}
		pt.JSON(w, r, pt.M{"pokemon": pokemon, "bundles": bundles})
	})
	r.With(rateLimitByIp(10, 15*time.Second), gpssDisabled).Route("/upload", func(r chi.Router) {
		r.Post("/pokemon", func(w http.ResponseWriter, r *http.Request) {
			startTime := time.Now()
			var log interface{}
			var err error
			r.Body = http.MaxBytesReader(w, r.Body, 40000)
			ip := helpers.GetIP(r, legacyKey)
			patreon, patreonCode, patreonDiscordID := getPatronInfo(r)

			file, _, err := r.FormFile("pkmn")
			if err != nil {
				log = helpers.GenerateFailedUploadLog(ip, r.Header.Get("source"), r.Header.Get("discord-user"),
					fmt.Sprintf("Failed to read pkmn from request body, error details: %s", err.Error()), patreon, patreonCode, patreonDiscordID)
				svcLogs.UpsertLog(r.Context(), &log)
				pt.JSON(w, r, pt.M{"uploaded": false, "code": "", "error": err.Error()})
				return
			}

			uploaded, approved, code, logToSentry, _, err := processPokemon(file, r.Header.Get("generation"), r.Header.Get("source"), ip, r, false, "")
			if err != nil {
				if logToSentry {
					w.WriteHeader(http.StatusInternalServerError)
					helpers.LogToSentry(err)
				}
				log = helpers.GenerateFailedUploadLog(ip, r.Header.Get("source"), r.Header.Get("discord-user"),
					err.Error(), patreon, patreonCode, patreonDiscordID)
				svcLogs.UpsertLog(r.Context(), &log)
				pt.JSON(w, r, pt.M{"uploaded": uploaded, "code": code, "error": err.Error()})
				return
			}

			// No errors
			err = fmt.Errorf("no errors")
			if !approved {
				err = fmt.Errorf("your pokemon is being held for manual review")
			}
			go helpers.MeasureUploadTime("individual", time.Since(startTime))
			pt.JSON(w, r, pt.M{"uploaded": uploaded, "approved": approved, "code": code, "error": err.Error()})
		})

		r.Post("/bundle", func(w http.ResponseWriter, r *http.Request) {
			startTime := time.Now()
			var log interface{}
			var err error
			r.Body = http.MaxBytesReader(w, r.Body, 40000)
			ip := helpers.GetIP(r, legacyKey)
			patreon, patreonCode, patreonDiscordID := getPatronInfo(r)

			missingData, headers := helpers.GetRequiredData(w, r, "header", []string{"count", "generations"}, true)
			if missingData {
				return
			}

			// Read how many pokemon we want in the bundle
			// Max of 6, min of 2
			// Make sure that count is an int
			count, err := strconv.Atoi(headers["count"])
			if err != nil || (count < 1 || count > 6) {
				log = helpers.GenerateFailedUploadLog(ip, r.Header.Get("source"), r.Header.Get("discord-user"),
					fmt.Sprintf("Invalid count header of %v", headers["count"]), patreon, patreonCode, patreonDiscordID)
				svcLogs.UpsertLog(r.Context(), &log)
				helpers.HttpError(w, r, http.StatusBadRequest, fmt.Errorf("Your count header of %v is either not an int or it is greater than 6 or less than 1", headers["count"]), false, true, false)
				return
			}

			// Okay count is good, now let's split the generations by their seperator
			generations := strings.Split(headers["generations"], ",")
			// generations slice should equal the count
			if len(generations) != count {
				log = helpers.GenerateFailedUploadLog(ip, r.Header.Get("source"), r.Header.Get("discord-user"),
					fmt.Sprintf("Invalid generations slice count of %d when count is %d", len(generations), count), patreon, patreonCode, patreonDiscordID)
				svcLogs.UpsertLog(r.Context(), &log)
				helpers.HttpError(w, r, http.StatusBadRequest, fmt.Errorf("You've provided a count of %d but only provided %d generations", count, len(generations)), false, true, false)
				return
			}

			var minGen, maxGen float32
			// okay, now let's make sure the generations are valid

			for _, gen := range generations {
				// conver to uppercase for LGPE
				gen = strings.ToUpper(gen)
				if val, ok := models.ValidGenerations[gen]; ok {
					// value okay? good, let's do some checks
					if minGen == 0 && maxGen == 0 {
						// okay let's set minGen and maxGen (and their strings) to this gen first
						minGen = val
						maxGen = val
						// if we get here, then next loop so we don't run the if checks below
						continue
					}
					// check if this current gen is less than minGen
					if val < minGen {
						// set minGen (and its string) to the current value
						minGen = val
					}

					// check if this current gen is greater than maxGen
					if val > maxGen {
						// set maxGen (and its string) to the current value
						maxGen = val
					}
				} else {
					// We have an invalid generation and should stop here
					log = helpers.GenerateFailedUploadLog(ip, r.Header.Get("source"), r.Header.Get("discord-user"),
						fmt.Sprintf("Invalid generation provided, %s is not a valid generation", gen), patreon, patreonCode, patreonDiscordID)
					svcLogs.UpsertLog(r.Context(), &log)
					helpers.HttpError(w, r, http.StatusBadRequest, fmt.Errorf("%s is not a valid generation", gen), false, true, false)
					return
				}
			}

			// files will store our files, we don't want to upload if there's missing data.
			files := []multipart.File{}

			for i := 0; i < count; i++ {
				// Each pokemon file should be set as pkmn# (so pkmn1, pkmn2 and so on)
				// If one isn't there, we stop the upload.

				file, fileHeader, err := r.FormFile(fmt.Sprintf("pkmn%d", i+1)) // arrays start at 0, but humans count starting from 1 usually, it's weird.
				if err != nil {
					log = helpers.GenerateFailedUploadLog(ip, r.Header.Get("source"), r.Header.Get("discord-user"),
						fmt.Sprintf("Encountered error reading pkmn #%d of bundle upload, error details: %s", i+1, err.Error()), patreon, patreonCode, patreonDiscordID)
					svcLogs.UpsertLog(r.Context(), &log)

					// Set up Context
					success, context := helpers.GenerateSentryEventLogContext([]string{"ip", "path", "headers", "error"}, []interface{}{helpers.GetIP(r, legacyKey), r.URL.Path, r.Header, err})
					// log with context
					if success {
						helpers.LogToSentryWithContext(sentry.LevelWarning, "An error was encountered when reading a pokemon from a bundle upload", context)
					}
					helpers.HttpError(w, r, http.StatusBadRequest, err, false, false, true)
					return
				}

				if fileHeader.Size > 600 {
					log = helpers.GenerateFailedUploadLog(ip, r.Header.Get("source"), r.Header.Get("discord-user"),
						fmt.Sprintf("Encountered a bigger filesize than 600 bytes with pkmn #%d of bundle upload, file size is %d", i+1, fileHeader.Size), patreon, patreonCode, patreonDiscordID)
					svcLogs.UpsertLog(r.Context(), &log)

					// Set up Context
					success, context := helpers.GenerateSentryEventLogContext([]string{"ip", "path", "headers", "error"}, []interface{}{helpers.GetIP(r, legacyKey), r.URL.Path, r.Header, err})
					// log with context
					if success {
						helpers.LogToSentryWithContext(sentry.LevelWarning, "An error was encountered when uploading a pokemon that was too large in a bundle", context)
					}
					helpers.HttpError(w, r, http.StatusBadRequest, fmt.Errorf("Pokemon %d is too big, filesize was %d", i+1, fileHeader.Size), false, false, true)
					return
				}

				// append that file to the array
				files = append(files, file)
			}

			// downloadCodes slice for each pokemon's download code
			downloadCodes := []string{}
			alreadyUploaded := []string{}
			notApproved := []string{}
			uploadedPokemon := []string{}
			problemEncountered := false
			badPokemon := 0

			// Generate the bundle code here (because of how I did logs, I am stupid)
			bundleCode := pwgen.Num(10)

			// Okay we're here, now let's process each file

			for i, file := range files {
				uploaded, approved, code, logToSentry, _, err := processPokemon(file, strings.ToUpper(generations[i]), r.Header.Get("source"), ip, r, true, bundleCode)
				if err != nil {
					if err != fmt.Errorf("Your Pokemon is already uploaded") {
						if logToSentry {
							helpers.LogToSentry(err)
							problemEncountered = true
							badPokemon = i + 1
							break
						}
					}
					alreadyUploaded = append(alreadyUploaded, code)
				}

				if !approved {
					notApproved = append(notApproved, code)
				}

				if uploaded {
					uploadedPokemon = append(uploadedPokemon, code)
				}
				// if we get to here, then append the (new/old) download code to the slice
				downloadCodes = append(downloadCodes, code)
			}

			if problemEncountered {
				// Delete the pokemon that weren't already there.
				for _, code := range uploadedPokemon {
					err = svcGPSS.RemovePokemon(r.Context(), code, false, false)
					if err != nil {
						helpers.LogToSentry(err)
					}
					// Delete the logs of those
					err = svcLogs.DeleteLog(r.Context(), bson.M{"log_type": "gpss_upload", "download_code": code})
					if err != nil {
						helpers.LogToSentry(err)
					}
				}
				log = helpers.GenerateFailedUploadLog(ip, r.Header.Get("source"), r.Header.Get("discord-user"),
					fmt.Sprintf("Encountered an issue with CoreAPI on a bundled Pokemon"), patreon, patreonCode, patreonDiscordID)
				svcLogs.UpsertLog(r.Context(), &log)
				helpers.HttpError(w, r, http.StatusInternalServerError, fmt.Errorf("Something was wrong with pokemon #%d in your bundle, the previous pokemon were uploaded okay", badPokemon), false, false, false)
				return
			}

			// Okay no problems? Excellent, now let's prepare the bundle object!
			// The IsLegal field should be calculated inside the UpsertBundle function.
			bundle := &models.GPSSBundlePokemon{
				DownloadCodes: downloadCodes,
				UploadDate:    time.Now(),
				Patreon:       patreon,
				Count:         count,
				MinGen:        int(minGen),
				MaxGen:        int(maxGen),
				Approved:      len(notApproved) == 0, // if it is 0, then we've approved, otherwise it's not approved yet.
			}
			// set the IP for the logs
			r.Header.Set("IP-For-Logs", ip)
			// Upsert the bundle
			uploaded, code, err := svcGPSS.UpsertBundle(r.Context(), bundle, patreon, patreonCode, patreonDiscordID, bundleCode, &r.Header)
			if err != nil {
				if err.Error() != "Bundle already uploaded" {
					log = helpers.GenerateFailedUploadLog(ip, r.Header.Get("source"), r.Header.Get("discord-user"),
						fmt.Sprintf("Encountered a database issue of %s", err.Error()), patreon, patreonCode, patreonDiscordID)
					svcLogs.UpsertLog(r.Context(), &log)
					helpers.LogToSentry(err)
					helpers.HttpError(w, r, http.StatusInternalServerError, fmt.Errorf("Something went wrong with your upload, please try again later"), false, false, false)
					return
				} else {
					log = helpers.GenerateFailedUploadLog(ip, r.Header.Get("source"), r.Header.Get("discord-user"),
						fmt.Sprintf("Bundle is already uploaded"), patreon, patreonCode, patreonDiscordID)
					svcLogs.UpsertLog(r.Context(), &log)
				}
			} else {
				err = fmt.Errorf("No errors")
				if !bundle.Approved {
					err = fmt.Errorf("your bundle is being held for manual review")
				}
			}
			go helpers.DiscordPostGPSS("bundle", cli.Flags.WebHooks.Discord, cli.Flags.SiteURL, bundle)
			go helpers.MeasureUploadTime("bundle", time.Since(startTime))
			pt.JSON(w, r, pt.M{"uploaded": uploaded, "approved": bundle.Approved, "code": code, "error": err.Error()})
		})
	})

	r.With(gpssDisabled).Route("/search", func(r chi.Router) {
		r.With(rateLimitByIp(30, time.Second*10)).Post("/pokemon", func(w http.ResponseWriter, r *http.Request) {

			page := 1    // default to 1
			amount := 12 // default to 12

			// If we have a page number provided and it is an int
			if number, err := strconv.Atoi(r.URL.Query().Get("page")); err == nil {
				// set it to the page number
				page = number
			}

			// If we have an amount number provided and it is an int
			if number, err := strconv.Atoi(r.URL.Query().Get("amount")); err == nil {
				// make sure it isn't greater than...30 and that it is greater than or equal to 0
				if number <= 30 && number >= 1 {
					amount = number
				}
			}

			query, sort, ok, _ := helpers.ParseGPSSQuery(w, r)
			if !ok {
				return
			}

			query["deleted"] = false

			// We only want approved pokemon for non-logged-in users
			if _, err := getUserInfo(r); err != nil {
				query["approved"] = true
			}

			pksmMode := false
			if r.Header.Get("pksm-mode") == "yes" {
				pksmMode = true
				amount = 30
			}
			pokemon, pages, count, err := svcGPSS.ListPokemons(r.Context(), query, page, amount, sort, pksmMode)
			if err != nil {
				helpers.LogToSentry(err)
				pt.JSON(w, r, pt.M{"error": err.Error()})
				return
			}
			if pokemon == nil {
				pokemon = []*models.GPSSPokemon{}
			}

			// undo this when legacy support is no longer needed
			pokemonKey := "pokemon"
			totalKey := "total"
			if r.Header.Get("Flagbrew-V0-Legacy") == legacyKey && pksmMode {
				pokemonKey = "results"
				totalKey = "total_pkm"
			}

			if pksmMode {
				pkmn := []pt.M{}
				for _, pk := range pokemon {
					generation := "0"
					switch pk.Size {
					case 44:
						generation = "1"
					case 48:
						generation = "2"
					case 100:
						generation = "3"
					case 236:
						generation = "4"
					case 220:
						generation = "5"
					case 260:
						// check dex number
						if pk.Pokemon.DexNumber > 721 {
							generation = "7"
						} else {
							generation = "6"
						}
					default:
						generation = strconv.Itoa(pk.Pokemon.Generation)
					}
					pkmn = append(pkmn, pt.M{"base_64": pk.Base64, "legal": pk.Pokemon.IsLegal, "code": pk.DownloadCode, "generation": generation})
				}
				pt.JSON(w, r, pt.M{pokemonKey: pkmn, "pages": pages, "page": page, totalKey: count})
			} else {
				pt.JSON(w, r, pt.M{"pokemon": pokemon, "pages": pages, "page": page, "total": count})
			}
		})
		r.With(rateLimitByIp(10, time.Second*15)).Post("/bundles", func(w http.ResponseWriter, r *http.Request) {
			// I might remove the searching stuff if this goes bad.
			page := 1    // default to 1
			amount := 12 // default to 12

			// If we have a page number provided and it is an int
			if number, err := strconv.Atoi(r.URL.Query().Get("page")); err == nil {
				// set it to the page number
				page = number
			}

			// If we have an amount number provided and it is an int
			if number, err := strconv.Atoi(r.URL.Query().Get("amount")); err == nil {
				// make sure it isn't greater than...30 and that it is greater than or equal to 0
				if number <= 30 && number >= 1 {
					amount = number
				}
			}

			query, sort, ok, generations := helpers.ParseGPSSQuery(w, r)
			if !ok {
				return
			}

			approvedOnly := false
			// We only want approved pokemon for non-logged-in users
			if _, err := getUserInfo(r); err != nil {
				query["approved"] = true
				approvedOnly = true
			}
			query["in_group"] = true

			// pksmMode := false
			// if r.Header.Get("pksm-mode") == "yes" {
			// 	pksmMode = true
			// }
			// set PKSM mode to true to reduce memory hopefully
			pokemon, _, _, err := svcGPSS.ListPokemons(r.Context(), query, 1, 1000000, sort, true)
			if err != nil {
				helpers.LogToSentry(err)
				pt.JSON(w, r, pt.M{"error": err.Error()})
				return
			}

			downloadCodes := []string{}

			// loop through all the returned Pokemons
			for _, pk := range pokemon {
				// append to downloadCodes
				downloadCodes = append(downloadCodes, pk.DownloadCode)
			}
			bQuery := bson.M{"download_codes": bson.M{"$in": downloadCodes}}
			if approvedOnly {
				bQuery["approved"] = true
			}
			if generations != nil {
				bQuery["min_gen"] = bson.M{"$in": generations}
				bQuery["max_gen"] = bson.M{"$in": generations}
			}
			if r.Header.Get("pksm-mode") == "yes" {
				amount = 5
			}

			// Now query for the bundles
			bundles, pages, count, err := svcGPSS.ListBundles(r.Context(), bQuery, page, amount, sort)
			if err != nil {
				helpers.LogToSentry(err)
				pt.JSON(w, r, pt.M{"error": err.Error()})
				return
			}

			if bundles == nil {
				bundles = []*models.GPSSBundlePokemon{}
			}

			if r.Header.Get("pksm-mode") == "yes" {
				results := []pt.M{}
				for _, bundle := range bundles {
					mons := []pt.M{}
					for _, mon := range bundle.Pokemons {
						mons = append(mons, pt.M{"base_64": mon.Base64, "legality": mon.Legality, "generation": strconv.Itoa(mon.Generation)})
						// fmt.Println("mon", mon)
					}
					results = append(results, pt.M{"download_codes": bundle.DownloadCodes, "download_code": bundle.DownloadCode, "pokemons": mons, "patreon": bundle.Patreon, "min_gen": strconv.Itoa(bundle.MinGen), "max_gen": strconv.Itoa(bundle.MaxGen), "count": len(mons), "is_legal": bundle.IsLegal})
				}
				pt.JSON(w, r, pt.M{"bundles": results, "pages": pages, "page": page, "total": count})
				return
			}

			pt.JSON(w, r, pt.M{"bundles": bundles, "pages": pages, "page": page, "total": count})
		})
	})

	r.With(rateLimitByIp(10, 15*time.Second), gpssDisabled).Route("/download", func(r chi.Router) {
		startTime := time.Now()
		r.Get("/pokemon/{code}", func(w http.ResponseWriter, r *http.Request) {
			// check the download code
			codeStr := chi.URLParam(r, "code")
			if len(codeStr) != 10 {
				helpers.HttpError(w, r, 400, fmt.Errorf("Download code is invalid"), false, false, false)
				return
			}
			// the code should be an int
			_, err := strconv.Atoi(codeStr)
			if err != nil {
				helpers.HttpError(w, r, 400, fmt.Errorf("Download code is invalid"), false, false, false)
				return
			}
			// Only authed users can download non-approved pokemon
			approvedOnly := false
			if _, err := getUserInfo(r); err != nil {
				approvedOnly = true
			}

			// get the pokemon
			pokemon, err := svcGPSS.DownloadPokemon(r.Context(), codeStr, approvedOnly)
			if err != nil {
				sentryAlert := false
				if err != mongo.ErrNoDocuments {
					sentryAlert = true
				}
				helpers.HttpError(w, r, 400, err, false, sentryAlert, true)
				return
			}

			// generate the extension
			extension := ""
			switch pokemon.Pokemon.Version {
			case "GE":
				fallthrough
			case "GP":
				extension = "pb7"
			case "PLA":
				extension = "pa8"
			case "BD":
				fallthrough
			case "SP":
				fallthrough
			case "BDSP":
				extension = "pb8"
			case "GO":
				// Dunno how to handle GO so, let's just use pkm
				extension = "pkm"
			default:
				extension = fmt.Sprintf("pk%d", pokemon.Generation)
			}
			// generate the filename
			filename := fmt.Sprintf("%s %s (%s %d).%s", pokemon.Pokemon.Species, pokemon.Pokemon.Nickname, pokemon.Pokemon.Ot, pokemon.Pokemon.Tid, extension)

			// Increase the download metric
			helpers.IncreaseDownloads("individual", []string{strconv.Itoa(pokemon.Generation)}, []string{pokemon.Pokemon.Species}, []string{pokemon.Pokemon.Gender}, []bool{pokemon.Pokemon.IsLegal}, []bool{pokemon.Pokemon.IsShiny}, []bool{pokemon.Pokemon.IsEgg})
			// return the base64, generate and generated filename
			pt.JSON(w, r, pt.M{"pokemon": pokemon.Base64, "generation": pokemon.Generation, "filename": filename})

			helpers.MeasureDownloadTime("individual", time.Since(startTime))
		})

		r.Get("/bundle/{code}", func(w http.ResponseWriter, r *http.Request) {
			// check the download code
			codeStr := chi.URLParam(r, "code")
			if len(codeStr) != 10 {
				helpers.HttpError(w, r, 400, fmt.Errorf("Download code is invalid"), false, false, false)
				return
			}
			// the code should be an int
			_, err := strconv.Atoi(codeStr)
			if err != nil {
				helpers.HttpError(w, r, 400, fmt.Errorf("Download code is invalid"), false, false, false)
				return
			}
			// Only authed users can download non-approved pokemon
			approvedOnly := false
			if _, err := getUserInfo(r); err != nil {
				approvedOnly = true
			}

			// get the pokemon
			pokemon, err := svcGPSS.DownloadBundle(r.Context(), codeStr, approvedOnly)
			if err != nil {
				sentryAlert := false
				if err != mongo.ErrNoDocuments {
					sentryAlert = true
				}
				helpers.HttpError(w, r, 400, err, false, sentryAlert, true)
				return
			}

			// First create the slice that'll contain each pt.M object
			returnSlice := []pt.M{}
			generations := []string{}
			species := []string{}
			gender := []string{}
			legality := []bool{}
			shiny := []bool{}
			egg := []bool{}
			// First loop through each Pokemon
			for _, pk := range pokemon {
				// Generate the extension
				extension := ""
				switch pk.Pokemon.Version {
				case "GE":
					fallthrough
				case "GP":
					extension = "pb7"
				case "PLA":
					extension = "pa8"
				case "BD":
					fallthrough
				case "SP":
					fallthrough
				case "BDSP":
					extension = "pb8"
				case "GO":
					// Dunno how to handle GO so, let's just use pkm
					extension = "pkm"
				default:
					extension = fmt.Sprintf("pk%d", pk.Pokemon.Generation)
				}
				// Generate a temporary object
				tmpObj := pt.M{"download_code": pk.DownloadCode, "generation": pk.Generation, "legal": pk.Pokemon.IsLegal, "pokemon": pk.Base64, "filename": fmt.Sprintf("%s %s (%s %d).%s", pk.Pokemon.Species, pk.Pokemon.Nickname, pk.Pokemon.Ot, pk.Pokemon.Tid, extension)}
				// Now append it to the returnSlice
				returnSlice = append(returnSlice, tmpObj)
				// also append to the metric slices
				generations = append(generations, strconv.Itoa(pk.Generation))
				species = append(species, pk.Pokemon.Species)
				gender = append(gender, pk.Pokemon.Gender)
				legality = append(legality, pk.Pokemon.IsLegal)
				shiny = append(shiny, pk.Pokemon.IsShiny)
				egg = append(egg, pk.Pokemon.IsEgg)
			}

			// increment bundle downloads (only if it's not legacy because I believe PKSM individuall does it)
			if r.Header.Get("Flagbrew-V0-Legacy") == "" {
				helpers.IncreaseDownloads("bundle", generations, species, gender, legality, shiny, egg)
			}
			// With that done, let's return
			pt.JSON(w, r, pt.M{"pokemons": returnSlice, "count": len(returnSlice)})
			helpers.MeasureDownloadTime("bundle", time.Since(startTime))
			return
		})
	})
	r.Get("/view/{code}", func(w http.ResponseWriter, r *http.Request) {
		// check the download code
		codeStr := chi.URLParam(r, "code")
		if len(codeStr) != 10 {
			helpers.HttpError(w, r, 400, fmt.Errorf("Download code is invalid"), false, false, false)
			return
		}
		// the code should be an int
		_, err := strconv.Atoi(codeStr)
		if err != nil {
			helpers.HttpError(w, r, 400, fmt.Errorf("Download code is invalid"), false, false, false)
			return
		}
		// Only authed users can see non-approved pokemon
		approvedOnly := false
		if _, err := getUserInfo(r); err != nil {
			approvedOnly = true
		}
		// get the pokemon
		pokemon, err := svcGPSS.ListPokemon(r.Context(), codeStr, approvedOnly)
		if err != nil {
			sentryAlert := false
			if err != mongo.ErrNoDocuments {
				sentryAlert = true
			}
			helpers.HttpError(w, r, 400, err, false, sentryAlert, true)
			return
		}

		pt.JSON(w, r, pt.M{"pokemon": pokemon})
	})

	r.Get("/random/{amount}", func(w http.ResponseWriter, r *http.Request) {
		// Make sure the amount isn't greater than 6 but is more than 0
		amount, err := strconv.Atoi(chi.URLParam(r, "amount"))
		if err != nil || amount > 6 || amount < 1 {
			helpers.HttpError(w, r, 400, fmt.Errorf("Amount is invalid"), false, false, false)
			return
		}
		// get min gen (if it exists) or default to 1 from the url query
		minGen := 1
		if minGenStr := r.URL.Query().Get("minGen"); len(minGenStr) > 0 {
			minGen, err = strconv.Atoi(minGenStr)
			if err != nil {
				helpers.HttpError(w, r, 400, fmt.Errorf("Min gen is invalid"), false, false, false)
				return
			}
			// if it's below 1 or greater than 8 it's invalid
			if minGen < 1 || minGen > 8 {
				helpers.HttpError(w, r, 400, fmt.Errorf("Min gen is invalid"), false, false, false)
				return
			}
		}

		// do the same for max gen but default it to 8
		maxGen := 8
		if maxGenStr := r.URL.Query().Get("maxGen"); len(maxGenStr) > 0 {
			maxGen, err = strconv.Atoi(maxGenStr)
			if err != nil {
				helpers.HttpError(w, r, 400, fmt.Errorf("Max gen is invalid"), false, false, false)
				return
			}
			// if it's below 1 or greater than 8 it's invalid
			if maxGen < 1 || maxGen > 8 {
				helpers.HttpError(w, r, 400, fmt.Errorf("Max gen is invalid"), false, false, false)
				return
			}
		}

		// make sure the min gen is less than the max gen
		if minGen > maxGen {
			helpers.HttpError(w, r, 400, fmt.Errorf("Min gen can't be greater than max gen"), false, false, false)
			return
		}

		// check to see if we want LGPE inclduded
		includeLGPE := false
		if includeLGPEStr := r.URL.Query().Get("includeLGPE"); len(includeLGPEStr) > 0 {
			includeLGPE, err = strconv.ParseBool(includeLGPEStr)
			if err != nil {
				helpers.HttpError(w, r, 400, fmt.Errorf("Include LGPE is invalid"), false, false, false)
				return
			}
		}

		// Generate array from min to max as strings instead of ints
		genRange := []string{}
		for i := minGen; i <= maxGen; i++ {
			genRange = append(genRange, strconv.Itoa(i))
		}
		// if we want LGPE then append LGPE to it
		if includeLGPE {
			genRange = append(genRange, "LGPE")
		}

		// get the pokemon
		pokemon, err := svcGPSS.RandomPokemon(r.Context(), amount, genRange)
		if err != nil {
			// if it's not a mongo error, it's a bad request
			if err != mongo.ErrNoDocuments {
				helpers.HttpError(w, r, 400, err, false, true, false)
				return
			} else {
				// if it's a mongo error, it's a not found
				helpers.HttpError(w, r, 404, err, false, false, true)
				return
			}
		}

		pt.JSON(w, r, pokemon)
	})
}
