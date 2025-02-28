package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/FM1337/flagbrew-website-source-code/pkg/helpers"
	"github.com/go-chi/chi/v5"
	"github.com/lrstanley/pt"
)

func secretKeyRequired(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := helpers.GetIP(r, legacyKey)
		if ip == "" {
			helpers.HttpError(w, r, http.StatusInternalServerError, fmt.Errorf("there's a networking issue going on, please try again later. Internal Error Code: MISSING_NETWORK_DATA"), false, true, false)
			return
		}
		secret := r.Header.Get("Upload-Secret")
		if secret != cli.Flags.Secrets.Github.UploadSecret {
			helpers.HttpError(w, r, http.StatusForbidden, fmt.Errorf("%s You are not authorized to upload files", ip), false, true, false)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// TODO Logging for uploads and failed downloads
func registerFileRoutes(r chi.Router) {
	r.Route("/upload", func(r chi.Router) {
		r.With(apiAuthRequired).Post("/mystery-gift", func(w http.ResponseWriter, r *http.Request) {
			r.ParseMultipartForm(2 << 20)
			var buf bytes.Buffer
			// get the file
			file, header, err := r.FormFile("file")
			if err != nil {
				helpers.HttpError(w, r, http.StatusInternalServerError, err, false, true, false)
				return
			}
			defer file.Close()
			// copy the file
			_, err = io.Copy(&buf, file)
			if err != nil {
				helpers.HttpError(w, r, http.StatusInternalServerError, err, false, true, false)
				return
			}
			// Now insert
			size, err := svcFile.UploadMysteryGift(r.Context(), header.Filename, buf.Bytes())
			if err != nil {
				helpers.HttpError(w, r, http.StatusInternalServerError, err, false, true, false)
				return
			}
			pt.JSON(w, r, pt.M{"message": fmt.Sprintf("Successfully uploaded file %s with a size of %d", header.Filename, size)})
		})
		r.With(secretKeyRequired).Post("/patron-build", func(w http.ResponseWriter, r *http.Request) {
			// First get the required headers
			// TODO move header checks to helper function to reduce code
			missingHeaders, headers := helpers.GetRequiredData(w, r, "header", []string{"app", "hash", "extensions"}, true)
			if missingHeaders {
				return
			}

			app := headers["app"]
			hash := headers["hash"]
			extensions := headers["extensions"]

			// Extensions should be a comma seperated list, it'll tell us what files we want to process
			// We'll always have at-least two files, one that can be installed and the zip file so if it's less than 2, then we have a problem
			// Extensions should always match the field name that the file is attached to.
			extensionList := strings.Split(extensions, ",")
			if len(extensionList) < 2 {
				helpers.HttpError(w, r, http.StatusBadRequest, fmt.Errorf("expected a extensions list of at-least 2 items, got %d items instead", len(extensionList)), false, true, false)
				return
			}

			// Set the process size
			if err := r.ParseMultipartForm(32 << 20); err != nil {
				helpers.HttpError(w, r, http.StatusInternalServerError, err, false, true, false)
				return
			}
			// Create the sizes map
			sizes := make(map[string]int)
			// Loop through the extensions
			for _, ext := range extensionList {
				// Create the buffer
				var buf bytes.Buffer
				// get the file
				file, header, err := r.FormFile(ext)
				if err != nil {
					helpers.HttpError(w, r, http.StatusInternalServerError, err, false, true, false)
					return
				}
				defer file.Close()
				// copy the file
				_, err = io.Copy(&buf, file)
				if err != nil {
					helpers.HttpError(w, r, http.StatusInternalServerError, err, false, true, false)
					return
				}
				// Now insert
				size, err := svcFile.UploadPatreonBuild(r.Context(), header.Filename, buf.Bytes(), hash, app, ext)
				if err != nil {
					helpers.HttpError(w, r, http.StatusInternalServerError, err, false, true, false)
					return
				}
				// Create the map entry
				sizes[ext] = size
			}

			// Provided there are no errors, we should be good to return now
			pt.JSON(w, r, pt.M{"message": fmt.Sprintf("Successfully uploaded patreon builds of %s", app), "sizes": sizes})
		})
	})
	// r.With(rateLimitByIp(18, 1*time.Minute)).Route("/download", func(r chi.Router) {
	// 	r.Get("/mystery-gift/{filename}", func(w http.ResponseWriter, r *http.Request) {
	// 		filename := chi.URLParam(r, "filename")
	// 		if filename == "" {
	// 			helpers.HttpError(w, r, http.StatusInternalServerError, fmt.Errorf("missing filename for mystery gift download"), false, true, false)
	// 		}

	// 		filesize, file, err := svcFile.DownloadMysteryGift(r.Context(), filename)
	// 		if err != nil {
	// 			// Set up Context
	// 			success, context := helpers.GenerateSentryEventLogContext([]string{"ip", "path", "headers", "error"}, []interface{}{helpers.GetIP(r, legacyKey), r.URL.Path, r.Header, err})
	// 			// log with context
	// 			if success {
	// 				helpers.LogToSentryWithContext(sentry.LevelWarning, "An error was encountered when fetching data from the Mystery gift database", context)
	// 			}
	// 			helpers.HttpError(w, r, http.StatusInternalServerError, err, false, false, true)
	// 		}

	// 		w.Header().Set("Content-Length", strconv.Itoa(int(filesize)))
	// 		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	// 		w.Header().Set("Content-Type", "application/octet-stream")

	// 		_, err = io.Copy(w, &file)
	// 		if err != nil {
	// 			// Set up Context
	// 			success, context := helpers.GenerateSentryEventLogContext([]string{"ip", "path", "headers", "error"}, []interface{}{helpers.GetIP(r, legacyKey), r.URL.Path, r.Header, err})
	// 			// log with context
	// 			if success {
	// 				helpers.LogToSentryWithContext(sentry.LevelWarning, "An error was encountered when io copying data from the file to the write buffer", context)
	// 			}
	// 			helpers.HttpError(w, r, http.StatusInternalServerError, err, false, false, true)
	// 			return
	// 		}
	// 	})
	// })
}
