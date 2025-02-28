package main

import (
	"net/http"

	"github.com/FM1337/flagbrew-website-source-code/pkg/helpers"
	"github.com/FM1337/flagbrew-website-source-code/pkg/models"
	"github.com/go-chi/chi/v5"
	"github.com/lrstanley/pt"
)

func registerUsersRoutes(r chi.Router) {
	r.Use(apiAuthRequired)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		users, err := svcUsers.List(r.Context())
		if helpers.HttpError(w, r, http.StatusInternalServerError, err, false, true, false) {
			return
		}

		pt.JSON(w, r, pt.M{"users": users})
	})

	r.Delete("/{id}", func(w http.ResponseWriter, r *http.Request) {
		err := svcUsers.Delete(r.Context(), chi.URLParam(r, "id"))
		if helpers.HttpError(w, r, http.StatusInternalServerError, err, false, true, false) {
			return
		}

		w.WriteHeader(http.StatusOK)
	})

	r.Post("/", func(w http.ResponseWriter, r *http.Request) {
		var user models.User
		if !helpers.FDecode(w, r, &user) {
			return
		}

		if helpers.HttpError(w, r, http.StatusInternalServerError, svcUsers.Upsert(r.Context(), &user), false, true, false) {
			return
		}

		w.WriteHeader(http.StatusOK)
	})
}

func registerUserRoutes(r chi.Router) {
	r.Use(apiAuthRequired)

	r.Get("/{id}", func(w http.ResponseWriter, r *http.Request) {
		user, err := svcUsers.Get(r.Context(), chi.URLParam(r, "id"))
		if helpers.HttpError(w, r, http.StatusInternalServerError, err, false, true, false) {
			return
		}

		pt.JSON(w, r, user)
	})
}
