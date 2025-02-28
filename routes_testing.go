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

func registerTestRoutes(r chi.Router) {
	r.Post("/test/pokemonInfo", func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, 40000)
		file, _, err := r.FormFile("pkmn")
		if err != nil {
			// do something
		}
		pkmn := bytes.Buffer{}
		io.Copy(&pkmn, file)
		data, err := apiGPSS.GetPokemonInfo(pkmn.Bytes(), nil)
		if err != nil {
			fmt.Println(err.Error())
		}
		pt.JSON(w, r, pt.M{"pokemon": data})
	})

	r.Get("/test/latest-hash/{app}", func(w http.ResponseWriter, r *http.Request) {
		hash, err := svcFile.GetLatestAppHash(r.Context(), chi.URLParam(r, "app"))
		if err != nil {
			helpers.HttpError(w, r, 500, err, false, false, false)
			return
		}

		// Users can provide their hash (short or long) to see if they're on the latest
		providedHash := r.Header.Get("hash")
		if providedHash != "" {
			// Check the length,
			if len(providedHash) > 40 {
				helpers.HttpError(w, r, http.StatusBadRequest, fmt.Errorf("Hash length of %d too long", len(providedHash)), false, true, false)
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
}
