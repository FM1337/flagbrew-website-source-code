package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/FM1337/flagbrew-website-source-code/pkg/helpers"
	"github.com/getsentry/sentry-go"
	"github.com/go-chi/chi/v5"
	"github.com/lrstanley/pt"
)

func registerPKSMRoutes(r chi.Router) {
	r.With(rateLimitByIp(5, 10*time.Second)).Post("/legality", func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, 400000)

		if !loadedSettings["pksm_legality_analysis_enabled"].Value.(bool) {
			helpers.HttpError(w, r, http.StatusServiceUnavailable, fmt.Errorf("legality analysis is currently disabled, try again later"), false, false, false)
			return
		}

		missingHeaders, headers := helpers.GetRequiredData(w, r, "header", []string{"generation"}, true)
		if missingHeaders {
			return
		}

		if headers["generation"] == "6" {
			headers["generation"] = "Gen6"
		} else if headers["generation"] == "7" {
			headers["generation"] = "Gen7"
		} else if headers["generation"] == "8" {
			headers["generation"] = "Gen8"
		} else if headers["generation"] == "BDSP" {
			headers["generation"] = "Gen8b"
		} else {
			delete(headers, "generation")
		}

		// Set the process size
		if err := r.ParseMultipartForm(400000); err != nil {
			// Set up Context
			success, context := helpers.GenerateSentryEventLogContext([]string{"ip", "path", "headers", "error"}, []interface{}{helpers.GetIP(r, legacyKey), r.URL.Path, r.Header, err})
			// log with context
			if success {
				helpers.LogToSentryWithContext(sentry.LevelWarning, "An error was encountered when checking the legality of a pokemon", context)
			}
			helpers.HttpError(w, r, http.StatusInternalServerError, err, false, false, true)
			return
		}
		// Get the pokemon
		file, _, err := r.FormFile("pkmn")
		if err != nil {
			// Set up Context
			success, context := helpers.GenerateSentryEventLogContext([]string{"ip", "path", "headers", "error"}, []interface{}{helpers.GetIP(r, legacyKey), r.URL.Path, r.Header, err})
			// log with context
			if success {
				helpers.LogToSentryWithContext(sentry.LevelWarning, "An error was encountered when reading the pkmn from r.FormFile for legality checking", context)
			}
			helpers.HttpError(w, r, http.StatusInternalServerError, err, false, false, true)
			return
		}
		// Copy to the buffer after creating it
		var buf bytes.Buffer
		_, err = io.Copy(&buf, file)
		if err != nil {
			helpers.HttpError(w, r, http.StatusInternalServerError, err, false, true, true)
			return
		}

		// Now let's boogie
		legality, err := apiGPSS.GetLegalityInfo(buf.Bytes(), headers)
		if err != nil {
			helpers.HttpError(w, r, http.StatusInternalServerError, err, false, true, true)
			return
		}
		pt.JSON(w, r, legality)
	})
	r.With(rateLimitByIp(5, 10*time.Second)).Post("/legalize", func(w http.ResponseWriter, r *http.Request) {
		if !loadedSettings["pksm_auto_legality_enabled"].Value.(bool) {
			helpers.HttpError(w, r, http.StatusServiceUnavailable, fmt.Errorf("auto legality is currently disabled, try again later"), false, false, false)
			return
		}

		r.Body = http.MaxBytesReader(w, r.Body, 400000)

		missingHeaders, headers := helpers.GetRequiredData(w, r, "header", []string{"generation"}, true)
		if missingHeaders {
			return
		}

		if headers["generation"] == "6" {
			headers["generation"] = "Gen6"
		} else if headers["generation"] == "7" {
			headers["generation"] = "Gen7"
		} else if headers["generation"] == "8" {
			headers["generation"] = "Gen8"
		} else if headers["generation"] == "BDSP" {
			headers["generation"] = "Gen8b"
		} else {
			delete(headers, "generation")
		}

		// Set the process size
		if err := r.ParseMultipartForm(400000); err != nil {
			helpers.HttpError(w, r, http.StatusInternalServerError, err, false, true, true)
			return
		}
		// Get the pokemon
		file, _, err := r.FormFile("pkmn")
		if err != nil {
			helpers.HttpError(w, r, http.StatusInternalServerError, err, false, true, true)
			return
		}
		// Copy to the buffer after creating it
		var buf bytes.Buffer
		_, err = io.Copy(&buf, file)
		if err != nil {
			helpers.HttpError(w, r, http.StatusInternalServerError, err, false, true, true)
			return
		}

		// Now let's boogie
		legalize, err := apiGPSS.AutoLegalize(buf.Bytes(), headers)
		if err != nil {
			helpers.HttpError(w, r, http.StatusInternalServerError, err, false, true, true)
			return
		}
		pt.JSON(w, r, pt.M{
			"pokemon": legalize.Pokemon,
			"success": legalize.Legal,
			"ran":     legalize.Ran,
			"report":  legalize.Report,
		})
	})
}
