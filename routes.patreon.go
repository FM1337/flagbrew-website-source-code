package main

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/FM1337/flagbrew-website-source-code/pkg/helpers"
	"github.com/getsentry/sentry-go"
	"github.com/go-chi/chi/v5"
	"github.com/lrstanley/pt"
	qrcode "github.com/skip2/go-qrcode"
)

func patreonOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := helpers.GetIP(r, legacyKey)
		if ip == "" {
			helpers.HttpError(w, r, http.StatusInternalServerError, fmt.Errorf("There's a networking issue going on, please try again later. Internal Error Code: MISSING_NETWORK_DATA"), false, true, false)
			return
		}
		isPatreon := false
		// Check both the header and the URL param
		patreonCode := r.Header.Get("patreon")
		if patreonCode == "" {
			// Okay it's not in the header, check the route
			patreonCode = chi.URLParam(r, "patreon")
		}

		// If patreon code is still empty, then we don't need to do anything
		if patreonCode != "" {
			isPatreon = svcPatron.IsPatron(r.Context(), patreonCode)
		}

		if !isPatreon {
			// Stop em here dead in their tracks
			// Set up Context
			success, context := helpers.GenerateSentryEventLogContext([]string{"ip", "path", "patron_code", "headers"}, []interface{}{ip, r.URL.Path, patreonCode, r.Header})
			// log with context
			if success {
				helpers.LogToSentryWithContext(sentry.LevelWarning, "Non-Patron attempted to access Patron-Only route", context)
			}

			// return the error
			helpers.HttpError(w, r, http.StatusForbidden, fmt.Errorf("this route is for Flagbrew Patreon supporters only"), false, false, false)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func registerPatreonRoutes(r chi.Router) {
	r.With(patreonOnly).Get("/update-check/{app}", func(w http.ResponseWriter, r *http.Request) {
		app := chi.URLParam(r, "app")
		if app == "" {
			helpers.HttpError(w, r, http.StatusBadRequest, fmt.Errorf("Missing required route param app"), false, true, false)
			return
		}

		hash, err := svcFile.GetLatestAppHash(r.Context(), app)
		if err != nil {
			helpers.HttpError(w, r, http.StatusInternalServerError, err, false, true, true)
			return
		}

		// Users can provide their hash (short or long) to see if they're on the latest
		providedHash := r.Header.Get("hash")
		if providedHash != "" {
			// Check the length,
			if len(providedHash) > 40 {
				helpers.HttpError(w, r, http.StatusBadRequest, fmt.Errorf("Hash length of %d too long", len(providedHash)), false, true, true)
				return
			}
			// Okay let's check
			if hash == providedHash || strings.Contains(providedHash, hash.(string)) || strings.Contains(hash.(string), providedHash) {
				pt.JSON(w, r, pt.M{"hash": hash, "on_latest": true})
			} else {
				pt.JSON(w, r, pt.M{"hash": hash, "on_latest": false})
			}
			return
		}
		// If no hash provided, then don't show the on_latest
		pt.JSON(w, r, pt.M{"hash": hash})
	})

	r.With(patreonOnly).Get("/update/{patreon}/{app}/{hash}/{type}", func(w http.ResponseWriter, r *http.Request) {
		missingParams, params := helpers.GetRequiredData(w, r, "route", []string{"patreon", "app", "type", "hash"}, true)

		if missingParams {
			return
		}

		filesize, file, err := svcFile.DownloadPatreonBuild(r.Context(), params["app"], params["hash"], params["type"])
		if err != nil {
			helpers.HttpError(w, r, http.StatusInternalServerError, err, false, true, true)
		}

		w.Header().Set("Content-Length", strconv.Itoa(int(filesize)))
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s-%s.%s", params["app"], params["hash"], params["type"]))
		w.Header().Set("Content-Type", "application/octet-stream")

		_, err = io.Copy(w, &file)
		if err != nil {
			helpers.HttpError(w, r, http.StatusInternalServerError, err, false, true, true)
			return
		}
	})

	r.With(patreonOnly).Get("/update-qr/{app}/{hash}/{type}", func(w http.ResponseWriter, r *http.Request) {
		missingParams, params := helpers.GetRequiredData(w, r, "route", []string{"app", "type", "hash"}, true)
		if missingParams {
			return
		}

		patreon := r.Header.Get("patreon")

		if !svcFile.PatreonBuildExists(r.Context(), params["app"], params["hash"], params["type"]) {
			helpers.HttpError(w, r, http.StatusBadRequest, fmt.Errorf("app/app commit doesn't exist"), false, false, false)
			return
		}

		image, err := qrcode.Encode(fmt.Sprintf("%s/api/v2/patreon/update/%s/%s/%s/%s", cli.Flags.SiteURL, patreon, params["app"], params["hash"], params["type"]), qrcode.Medium, 512)
		if err != nil {
			helpers.HttpError(w, r, http.StatusInternalServerError, err, false, true, true)
			return
		}
		w.Header().Set("Content-Type", "image/png")
		w.Write(image)
	})
}
