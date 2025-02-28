package main

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/FM1337/flagbrew-website-source-code/pkg/helpers"
	"github.com/FM1337/flagbrew-website-source-code/pkg/models"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httprate"
	"github.com/google/go-github/github"
	"github.com/lrstanley/pt"
)

func apiAuthRequired(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sess := session.Load(r)
		user, err := sess.GetString("user")
		helpers.PanicIfErr(err)

		if user == "" {
			w.WriteHeader(http.StatusUnauthorized)
			pt.JSON(w, r, pt.M{"authenticated": false, "error": "Authentication required."})
			return
		}

		// Check if user is in database.
		if err = svcUsers.Exists(r.Context(), user); err != nil {
			helpers.PanicIfErr(sess.Destroy(w))
			w.WriteHeader(http.StatusUnauthorized)
			pt.JSON(w, r, pt.M{"authenticated": false, "error": "Authentication required."})
			return
		}

		next.ServeHTTP(w, r)
	})
}

func registerAuthRoutes(r chi.Router) {
	r.With(httprate.LimitByIP(4, 60*time.Second)).Get("/github/redirect", func(w http.ResponseWriter, r *http.Request) {
		state := helpers.GenRandString(15)
		sess := session.Load(r)

		if id, _ := sess.GetString("user"); id != "" {
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}

		helpers.PanicIfErr(sess.PutString(w, "state", state))
		http.Redirect(w, r, oauthConfig.AuthCodeURL(state), http.StatusFound)
	})

	r.Get("/github/manage", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "https://github.com/settings/connections/applications/"+cli.Flags.Auth.Github.ClientID, http.StatusFound)
	})

	r.With(httprate.LimitByIP(4, 60*time.Second)).Get("/github/callback", func(w http.ResponseWriter, r *http.Request) {
		sess := session.Load(r)

		if !cli.Debug {
			// Only check CSRF tokens if we're out of debug mode.
			state, err := sess.GetString("state")
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				pt.JSON(w, r, pt.M{"error": "Session token not found, possible CSRF (or cookies disabled)? Please try again."})
				return
			}

			helpers.PanicIfErr(sess.Remove(w, "state"))

			if state != r.FormValue("state") {
				w.WriteHeader(http.StatusBadRequest)
				pt.JSON(w, r, pt.M{"error": "Session token not found, possible CSRF (or cookies disabled)? Please try again."})
				return
			}
		}
		helpers.PanicIfErr(sess.Remove(w, "state"))

		token, err := oauthConfig.Exchange(r.Context(), r.FormValue("code"))
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			pt.JSON(w, r, pt.M{"error": "Error getting token: " + err.Error()})
			return
		}

		client := github.NewClient(oauthConfig.Client(r.Context(), token))

		guser, _, err := client.Users.Get(r.Context(), "")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			pt.JSON(w, r, pt.M{"error": "Error obtaining user information: " + err.Error()})
			return
		}

		user := &models.User{
			GithubID:       int(guser.GetID()),
			Token:          token.AccessToken,
			AvatarURL:      guser.GetAvatarURL(),
			Username:       guser.GetLogin(),
			Name:           guser.GetName(),
			Email:          guser.GetEmail(),
			AccountCreated: guser.GetCreatedAt().Time,
			AccountUpdated: guser.GetUpdatedAt().Time,
		}

		// Check if user is a valid admin user via ENV.
		var isValid bool
		for _, admin := range cli.Flags.Auth.Github.Admins {
			if admin == user.GithubID {
				isValid = true
			}
		}
		if !isValid {
			// Check if they're in the DB,  if they are, they've already been granted access.
			if err = svcUsers.GitHubExists(r.Context(), user.GithubID); err != nil {
				helpers.HttpError(w, r, http.StatusUnauthorized, fmt.Errorf("user with GitHub ID %d is not authorized to login", user.GithubID), false, false, false)
				return
			}
		}

		if err := svcUsers.Upsert(r.Context(), user); err != nil {
			helpers.HttpError(w, r, http.StatusInternalServerError, errors.New("error writing user to database"), false, true, false)
			logger.Warnf("error writing user to database: %v", err)
			return
		}

		helpers.PanicIfErr(sess.PutString(w, "user", user.ID.Hex()))

		w.WriteHeader(http.StatusOK)
		pt.JSON(w, r, pt.M{"authenticated": true, "user": user})
	})

	r.Get("/self", func(w http.ResponseWriter, r *http.Request) {
		sess := session.Load(r)
		id, err := sess.GetString("user")
		if err != nil || id == "" {
			w.WriteHeader(http.StatusUnauthorized)
			pt.JSON(w, r, pt.M{"authenticated": false})
			return
		}
		user, err := svcUsers.Get(r.Context(), id)
		if err != nil {
			if err.Error() == "The requested data could not be found" {
				w.WriteHeader(http.StatusNotFound)
				pt.JSON(w, r, pt.M{"authenticated": false, "error": "The requested user was not found."})
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			pt.JSON(w, r, pt.M{"authenticated": false, "error": "An internal server error occurred."})
			return
		}

		pt.JSON(w, r, pt.M{"authenticated": true, "user": user})
	})

	r.Get("/logout", func(w http.ResponseWriter, r *http.Request) {
		sess := session.Load(r)
		helpers.PanicIfErr(sess.Destroy(w))

		w.WriteHeader(http.StatusOK)
		pt.JSON(w, r, pt.M{"authenticated": false})
	})
}
