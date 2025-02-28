package main

import (
	"fmt"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/FM1337/flagbrew-website-source-code/pkg/helpers"
	"github.com/FM1337/flagbrew-website-source-code/pkg/models"
	"github.com/go-chi/chi/v5"
	"github.com/lrstanley/pt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func registerModerationRoutes(r chi.Router) {
	r.Get("/words", func(w http.ResponseWriter, r *http.Request) {
		page := 1    // default to 1
		amount := 30 // default to 30

		// If we have a page number provided and it is an int
		if number, err := strconv.Atoi(r.URL.Query().Get("page")); err == nil {
			// set it to the page number
			page = number
		}

		// If we have an amount number provided and it is an int
		if number, err := strconv.Atoi(r.URL.Query().Get("amount")); err == nil {
			// make sure it isn't greater than...200 and that it is greater than or equal to 0
			if number <= 200 && number >= 1 {
				amount = number
			}
		}

		sortBy := r.URL.Query().Get("sort")
		sortDesc := r.URL.Query().Get("sortDesc")

		sort := bson.M{}

		if sortBy != "" {
			if sortDesc == "yes" {
				sort = bson.M{sortBy: -1}
			} else {
				sort = bson.M{sortBy: 1}
			}
		}

		query := bson.M{}

		words, pages, total, err := svcFilter.ListWords(r.Context(), query, page, amount, sort)
		if err != nil {
			helpers.HttpError(w, r, 500, err, false, true, false)
			return
		}
		pt.JSON(w, r, pt.M{"words": words, "pages": pages, "total": total})
	})
	r.Post("/words", func(w http.ResponseWriter, r *http.Request) {
		user, err := getUserInfo(r)
		if err != nil {
			helpers.HttpError(w, r, 500, err, false, true, false)
			return
		}

		missingData, returnData := helpers.GetRequiredData(w, r, "post", []string{"word", "strict", "case_sensitive"}, true)
		if missingData {
			return
		}

		strict := false
		if returnData["strict"] == "true" {
			strict = true
		}
		caseInsensitive := true
		if returnData["case_sensitive"] == "true" {
			caseInsensitive = false
		}

		// Now insert the word
		err = svcFilter.AddWord(r.Context(), returnData["word"], strict, caseInsensitive, user.Username)
		if err != nil {
			errCode := http.StatusBadRequest
			if err.Error() != "word exists already" {
				helpers.LogToSentry(err)
				errCode = http.StatusInternalServerError
			}
			helpers.HttpError(w, r, errCode, err, false, false, false)
			return
		}

		pt.JSON(w, r, pt.M{"message": fmt.Sprintf("%s successfully added to word filter list", returnData["word"])})

	})
	r.Delete("/words/{word}", func(w http.ResponseWriter, r *http.Request) {
		user, err := getUserInfo(r)
		if err != nil {
			helpers.HttpError(w, r, 500, err, false, true, false)
			return
		}

		word := chi.URLParam(r, "word")
		sensitive := r.PostFormValue("case_sensitive")
		if sensitive == "" {
			helpers.HttpError(w, r, 400, fmt.Errorf("Word Deletion: case_sensitive is missing"), false, true, false)
			return
		}
		caseInsensitive := true
		if sensitive == "true" {
			caseInsensitive = false
		}

		err = svcFilter.RemoveWord(r.Context(), word, caseInsensitive)
		if err != nil {
			errCode := http.StatusBadRequest
			if err.Error() != "word doesn't exist" {
				helpers.LogToSentry(err)
				errCode = http.StatusInternalServerError
			}
			helpers.HttpError(w, r, errCode, err, false, false, false)
			return
		}
		// Delete is successful, insert an word deletion log TODO
		var log interface{}
		log = helpers.GenerateWordDeleteLog(user.Username, word)
		if helpers.HttpError(w, r, http.StatusInternalServerError, svcLogs.UpsertLog(r.Context(), &log), false, true, false) {
			return
		}

		// Done
		pt.JSON(w, r, pt.M{"message": fmt.Sprintf("word %s successfully deleted", word)})
	})
	r.Get("/logs", func(w http.ResponseWriter, r *http.Request) {
		page := 1    // default to 1
		amount := 30 // default to 30

		// If we have a page number provided and it is an int
		if number, err := strconv.Atoi(r.URL.Query().Get("page")); err == nil {
			// set it to the page number
			page = number
		}

		// If we have an amount number provided and it is an int
		if number, err := strconv.Atoi(r.URL.Query().Get("amount")); err == nil {
			// make sure it isn't greater than...200 and that it is greater than or equal to 0
			if number <= 200 && number >= 1 {
				amount = number
			}
		}

		logType := r.URL.Query().Get("type")
		if logType == "" {
			helpers.HttpError(w, r, 400, fmt.Errorf("Log Type missing from request"), false, true, false)
			return
		}

		sortBy := r.URL.Query().Get("sort")
		sortDesc := r.URL.Query().Get("sortDesc")

		sort := bson.M{}

		if sortBy != "" {
			if sortDesc == "yes" {
				sort = bson.M{sortBy: -1}
			} else {
				sort = bson.M{sortBy: 1}
			}
		}

		// If we want pending approvals only, add to the filter
		query := bson.M{"log_type": logType}
		if r.URL.Query().Get("pending") == "yes" && logType == "gpss_upload" {
			query["approved"] = false
			query["rejected"] = false
		}

		var logs, pages, total interface{}
		var err error
		switch logType {
		case "banned":
			logs, pages, total, err = svcBans.ListBans(r.Context(), bson.M{}, page, amount, sort)
			break
		case "restrictions":
			logs, pages, total, err = svcRestrict.ListRestricted(r.Context(), bson.M{}, page, amount, sort)
			break
		default:
			logs, pages, total, err = svcLogs.ListLogs(r.Context(), query, page, amount, sort)
			break
		}
		if err != nil {
			helpers.HttpError(w, r, 500, err, false, true, false)
			return
		}

		pt.JSON(w, r, pt.M{"logs": logs, "pages": pages, "total": total})
	})

	r.Get("/log/{type}", func(w http.ResponseWriter, r *http.Request) {
		// Get the route info
		logType := chi.URLParam(r, "type")
		if !models.ValidLogTypes[logType] {
			helpers.HttpError(w, r, 400, fmt.Errorf("Moderation Log Fetch: Log type is invalid"), false, true, false)
			return
		}
		// Get the query field and query value
		queryField := r.URL.Query().Get("query_field")
		if queryField == "" {
			helpers.HttpError(w, r, 400, fmt.Errorf("Moderation Log Fetch: Query Field is required"), false, true, false)
			return
		}
		queryValue := r.URL.Query().Get("query_value")
		if queryValue == "" {
			helpers.HttpError(w, r, 400, fmt.Errorf("Moderation Log Fetch: Query Value is required"), false, true, false)
			return
		}

		// construct the query
		query := bson.M{queryField: queryValue, "log_type": logType}

		log, err := svcLogs.GetLog(r.Context(), query)
		if err != nil {
			helpers.HttpError(w, r, 500, err, false, true, false)
			return
		}
		pt.JSON(w, r, pt.M{"log": log})
	})

	r.Delete("/delete/gpss/{type}/{code}", func(w http.ResponseWriter, r *http.Request) {

		user, err := getUserInfo(r)
		if err != nil {
			helpers.HttpError(w, r, 500, err, false, true, false)
			return
		}

		code := chi.URLParam(r, "code")
		if len(code) != 10 {
			helpers.HttpError(w, r, 400, fmt.Errorf("GPSS Deletion: Download code is invalid"), false, true, false)
			return
		}

		entityType := chi.URLParam(r, "type")
		if !models.ValidGPSSEntities[entityType] {
			helpers.HttpError(w, r, 400, fmt.Errorf("GPSS Deletion: Entity type is invalid"), false, true, false)
			return
		}

		reason := r.PostFormValue("reason")
		if reason == "" {
			helpers.HttpError(w, r, 400, fmt.Errorf("GPSS Deletion: Deletion reason is missing"), false, true, false)
			return
		}

		// Call the deletion first
		if entityType == "Pokemon" {
			err = svcGPSS.RemovePokemon(r.Context(), code, false, true)
		} else {
			err = svcGPSS.RemoveBundle(r.Context(), code)
		}

		if err != nil {
			helpers.HttpError(w, r, 500, err, false, true, false)
			return
		}

		// If no errors, then insert a new deletion log
		var log interface{}
		log = helpers.GenerateDeletionLog(user.Name, reason, entityType, code)
		if helpers.HttpError(w, r, http.StatusInternalServerError, svcLogs.UpsertLog(r.Context(), &log), false, true, false) {
			return
		}

		// Now update the upload log to reflect deletion
		if helpers.HttpError(w, r, http.StatusInternalServerError, svcLogs.UpdateLog(r.Context(), bson.M{"log_type": "gpss_upload", "download_code": code}, bson.M{"$set": bson.M{"deleted": true}}), false, true, false) {
			return
		}

		// Deletion successful
		pt.JSON(w, r, pt.M{"message": fmt.Sprintf("%s of download code %s has been successfully deleted", entityType, code)})
	})

	r.Post("/ban", func(w http.ResponseWriter, r *http.Request) {
		user, err := getUserInfo(r)
		if err != nil {
			helpers.HttpError(w, r, 500, err, false, true, false)
			return
		}

		ip := r.PostFormValue("ip")
		if net.ParseIP(ip) == nil {
			helpers.HttpError(w, r, 400, fmt.Errorf("IP Ban Error: IP is invalid"), false, true, false)
			return
		}

		reason := r.PostFormValue("reason")
		if reason == "" {
			helpers.HttpError(w, r, 400, fmt.Errorf("IP Ban Error: Ban reason is missing"), false, true, false)
			return
		}

		if helpers.HttpError(w, r, http.StatusInternalServerError, svcBans.Ban(r.Context(), helpers.GenerateBan(ip, reason, user.Name)), false, true, false) {
			return
		}

		pt.JSON(w, r, pt.M{"message": fmt.Sprintf("%s banned successfully", ip)})
	})

	r.Delete("/ban/{ip}", func(w http.ResponseWriter, r *http.Request) {
		user, err := getUserInfo(r)
		if err != nil {
			helpers.HttpError(w, r, 500, err, false, true, false)
			return
		}

		ip := chi.URLParam(r, "ip")
		if net.ParseIP(ip) == nil {
			helpers.HttpError(w, r, 400, fmt.Errorf("IP Unban Error: IP is invalid"), false, true, false)
			return
		}

		// Get the ban first
		ban, err := svcBans.ListBan(r.Context(), bson.M{"ip": ip})
		if helpers.HttpError(w, r, http.StatusInternalServerError, err, false, true, false) {
			return
		}

		// Attempt the delete
		if helpers.HttpError(w, r, http.StatusInternalServerError, svcBans.Unban(r.Context(), ip), false, true, false) {
			return
		}

		// Delete is successful, insert an unban log
		var log interface{}
		log = helpers.GenerateUnbanLog(user.Name, ban)
		if helpers.HttpError(w, r, http.StatusInternalServerError, svcLogs.UpsertLog(r.Context(), &log), false, true, false) {
			return
		}

		// Done
		pt.JSON(w, r, pt.M{"message": fmt.Sprintf("%s unbanned successfully", ip)})
	})

	r.Get("/ban/{ip}", func(w http.ResponseWriter, r *http.Request) {
		ip := chi.URLParam(r, "ip")
		if net.ParseIP(ip) == nil {
			helpers.HttpError(w, r, 400, fmt.Errorf("IP Ban Check Error: IP is invalid"), false, true, false)
			return
		}

		ban, err := svcBans.ListBan(r.Context(), bson.M{"ip": ip})
		if err != nil {
			if err != mongo.ErrNoDocuments {
				helpers.HttpError(w, r, http.StatusInternalServerError, err, false, false, false)
				return
			}
			pt.JSON(w, r, pt.M{"ban": false})
			return
		}

		pt.JSON(w, r, pt.M{"ban": ban})
	})

	r.Get("/settings", func(w http.ResponseWriter, r *http.Request) {

		pt.JSON(w, r, pt.M{"settings": loadedSettings})
	})

	r.Put("/settings/{setting}", func(w http.ResponseWriter, r *http.Request) {
		key := chi.URLParam(r, "setting")
		if key == "" {
			helpers.HttpError(w, r, http.StatusBadRequest, fmt.Errorf("Missing Setting Key from URL"), false, true, false)
			return
		}

		if _, ok := loadedSettings[key]; !ok {
			helpers.HttpError(w, r, http.StatusBadRequest, fmt.Errorf("Setting %s doesn't exist", key), false, true, false)
			return
		}

		value := r.PostFormValue("value")
		if value == "" && !loadedSettings[key].CanBeEmpty {
			helpers.HttpError(w, r, http.StatusBadRequest, fmt.Errorf("Missing value for updating %s setting", key), false, true, false)
			return
		}
		// get user info

		user, err := getUserInfo(r)
		if err != nil {
			helpers.HttpError(w, r, http.StatusBadRequest, fmt.Errorf("Could not get user info, error details: %s", err.Error()), false, true, false)
			return
		}
		// keep originalSetting until we've updated the database
		originalSetting := *loadedSettings[key]
		var result interface{}

		// Try to conver the value to whatever it wants
		switch loadedSettings[key].Type {
		case "bool":
			result, err = strconv.ParseBool(value)
			if err != nil {
				helpers.HttpError(w, r, http.StatusBadRequest, fmt.Errorf("Setting %s wanted bool, did not get bool", key), false, true, false)
				return
			}
		case "int":
			result, err = strconv.Atoi(value)
			if err != nil {
				helpers.HttpError(w, r, http.StatusBadRequest, fmt.Errorf("setting %s wanted int, did not get int", key), false, true, false)
				return
			}
		default:
			helpers.HttpError(w, r, http.StatusInternalServerError, fmt.Errorf("setting %s wanted %s, but we can't handle that yet", key, loadedSettings[key].Type), false, true, false)
			return
		}
		// Update the setting in memory
		setting := loadedSettings[key]
		setting.Value = result
		setting.ModifiedDate = time.Now()
		loadedSettings[key] = setting

		// Now update the database
		err = svcSettings.UpdateSetting(r.Context(), key, result)
		if err != nil {
			// If we hit an error restore the original setting in memory
			loadedSettings[key] = &originalSetting
			// then tell the user something went wrong
			helpers.HttpError(w, r, http.StatusInternalServerError, fmt.Errorf("could not update setting in database, original setting in memory restored, error details: %s", err.Error()), false, true, false)
			return
		}

		// Now log the change
		var log interface{}
		log = helpers.GenerateSettingChangeLog(key, user.Name, originalSetting.Value, setting.Value)
		err = svcLogs.UpsertLog(r.Context(), &log)
		if err != nil {
			// If we hit an error restore the original setting in memory
			loadedSettings[key] = &originalSetting
			// then tell the user something went wrong
			helpers.HttpError(w, r, http.StatusInternalServerError, fmt.Errorf("could not log setting change in database, original setting in memory restored, error details: %s", err.Error()), false, true, false)
			return
		}

		pt.JSON(w, r, pt.M{"message": fmt.Sprintf("%s has been successfully updated with the value %v", key, result)})
	})

	r.Post("/approve", func(w http.ResponseWriter, r *http.Request) {
		user, err := getUserInfo(r)
		if err != nil {
			helpers.HttpError(w, r, http.StatusInternalServerError, err, false, true, false)
			return
		}

		// check for data
		missingData, returnData := helpers.GetRequiredData(w, r, "post", []string{"code"}, true)
		if missingData {
			return
		}

		// Now approve
		err = svcApproval.Approve(r.Context(), returnData["code"], user.Name) // TODO should I switch to user.Username?
		if err != nil {
			helpers.HttpError(w, r, http.StatusInternalServerError, err, false, true, false)
			return
		}

		pt.JSON(w, r, pt.M{"message": fmt.Sprintf("Pokemon %s has been approved successfully", returnData["code"])})
	})

	r.Post("/reject", func(w http.ResponseWriter, r *http.Request) {
		user, err := getUserInfo(r)
		if err != nil {
			helpers.HttpError(w, r, http.StatusInternalServerError, err, false, true, false)
			return
		}

		// check for data
		missingData, returnData := helpers.GetRequiredData(w, r, "post", []string{"code", "reason"}, true)
		if missingData {
			return
		}

		// Now approve
		err = svcApproval.Reject(r.Context(), returnData["code"], returnData["reason"], user.Name) // TODO should I switch to user.Username?
		if err != nil {
			helpers.HttpError(w, r, http.StatusInternalServerError, err, false, true, false)
			return
		}

		pt.JSON(w, r, pt.M{"message": fmt.Sprintf("Pokemon %s has been rejected successfully", returnData["code"])})
	})

	r.Post("/restrict", func(w http.ResponseWriter, r *http.Request) {
		user, err := getUserInfo(r)
		if err != nil {
			helpers.HttpError(w, r, http.StatusInternalServerError, err, false, true, false)
			return
		}

		// check for data
		missingData, returnData := helpers.GetRequiredData(w, r, "post", []string{"ip", "reason"}, true)
		if missingData {
			return
		}

		// Now restrict them
		err = svcRestrict.RestrictUploader(r.Context(), returnData["ip"], returnData["reason"], user.Name)
		if err != nil {
			helpers.HttpError(w, r, http.StatusInternalServerError, err, false, true, false)
			return
		}

		pt.JSON(w, r, pt.M{"message": fmt.Sprintf("%s now requires GPSS uploads to be approved", returnData["ip"])})
	})

	r.Get("/restrict/{ip}", func(w http.ResponseWriter, r *http.Request) {
		// check for data
		missingData, returnData := helpers.GetRequiredData(w, r, "route", []string{"ip"}, true)
		if missingData {
			return
		}

		// Check to see if they're restricted
		restricted, reason := svcRestrict.IsUploaderRestricted(r.Context(), returnData["ip"])
		pt.JSON(w, r, pt.M{"restricted": restricted, "reason": reason})
	})

	r.Delete("/restrict/{ip}", func(w http.ResponseWriter, r *http.Request) {
		user, err := getUserInfo(r)
		if err != nil {
			helpers.HttpError(w, r, http.StatusInternalServerError, err, false, true, false)
			return
		}

		// check for data
		missingData, returnData := helpers.GetRequiredData(w, r, "route", []string{"ip"}, true)
		if missingData {
			return
		}

		// Now unrestrict them
		err = svcRestrict.UnrestrictUploader(r.Context(), returnData["ip"], user.Name)
		if err != nil {
			helpers.HttpError(w, r, http.StatusInternalServerError, err, false, true, false)
			return
		}

		pt.JSON(w, r, pt.M{"message": fmt.Sprintf("%s no longer requires GPSS uploads to be approved", returnData["ip"])})
	})

	// 	r.Post("/migrate", func(w http.ResponseWriter, r *http.Request) {
	// 		// check for data
	// 		oldPkmns, _, err := r.FormFile("pokemons")
	// 		if err != nil {
	// 			helpers.HttpError(w, r, http.StatusBadRequest, err, false, true, false)
	// 			return
	// 		}
	// 		oldLogs, _, err := r.FormFile("logs")
	// 		if err != nil {
	// 			helpers.HttpError(w, r, http.StatusBadRequest, err, false, true, false)
	// 			return
	// 		}

	// 		// read into bytes

	// 		oldPkmnsBytes, err := ioutil.ReadAll(oldPkmns)
	// 		if err != nil {
	// 			helpers.HttpError(w, r, http.StatusInternalServerError, err, false, true, false)
	// 			return
	// 		}

	// 		oldLogBytes, err := ioutil.ReadAll(oldLogs)
	// 		if err != nil {
	// 			helpers.HttpError(w, r, http.StatusInternalServerError, err, false, true, false)
	// 			return
	// 		}

	// 		// call GPSS's Migrate
	// 		go func() {
	// 			err = svcGPSS.Migrate(oldLogBytes, oldPkmnsBytes, &r.Header)
	// 			if err != nil {
	// 				helpers.LogToSentry(err)
	// 			}
	// 		}()

	//		pt.JSON(w, r, pt.M{"message": "Migrating has begun, if any errors occurr they will be logged to sentry!"})
	//	})
}

func getUserInfo(r *http.Request) (user *models.User, err error) {
	// Get the current user information
	sess := session.Load(r)
	userID, err := sess.GetString("user")
	if err != nil {
		return user, err
	}

	user, err = svcUsers.Get(r.Context(), userID)
	if err != nil {
		return user, err
	}

	return user, err
}
