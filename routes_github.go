package main

import (
	"net/http"

	"github.com/FM1337/flagbrew-website-source-code/pkg/helpers"
	"github.com/getsentry/sentry-go"
	"github.com/go-chi/chi/v5"
	"github.com/lrstanley/pt"
)

func registerGitHubRoutes(r chi.Router) {
	r.Get("/members", func(w http.ResponseWriter, r *http.Request) {
		members, err := svcGitHub.ListMembers(r.Context())
		if helpers.HttpError(w, r, http.StatusInternalServerError, err, false, false, true) {
			// Set up Context
			success, context := helpers.GenerateSentryEventLogContext([]string{"ip", "path", "headers", "error"}, []interface{}{helpers.GetIP(r, legacyKey), r.URL.Path, r.Header, err})
			// log with context
			if success {
				helpers.LogToSentryWithContext(sentry.LevelWarning, "An error was encountered when listing github members from the database", context)
			}
			return
		}

		pt.JSON(w, r, pt.M{"members": members})
	})

	r.Get("/repos", func(w http.ResponseWriter, r *http.Request) {
		repos, err := svcGitHub.ListRepos(r.Context())
		if helpers.HttpError(w, r, http.StatusInternalServerError, err, false, false, true) {
			// Set up Context
			success, context := helpers.GenerateSentryEventLogContext([]string{"ip", "path", "headers", "error"}, []interface{}{helpers.GetIP(r, legacyKey), r.URL.Path, r.Header, err})
			// log with context
			if success {
				helpers.LogToSentryWithContext(sentry.LevelWarning, "An error was encountered when listing github repos from the database", context)
			}
			return
		}

		pt.JSON(w, r, pt.M{"repos": repos})
	})

	r.Get("/repo/{repo}", func(w http.ResponseWriter, r *http.Request) {
		repo, err := svcGitHub.ListRepo(r.Context(), chi.URLParam(r, "repo"))
		if helpers.HttpError(w, r, http.StatusInternalServerError, err, false, false, true) {
			// Set up Context
			success, context := helpers.GenerateSentryEventLogContext([]string{"ip", "path", "headers", "repo", "error"}, []interface{}{helpers.GetIP(r, legacyKey), r.URL.Path, r.Header, chi.URLParam(r, "repo"), err})
			// log with context
			if success {
				helpers.LogToSentryWithContext(sentry.LevelWarning, "An error was encountered when listing a github repo from the database", context)
			}
			return
		}

		pt.JSON(w, r, pt.M{"repo": repo})
	})
}
