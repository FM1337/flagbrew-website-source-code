package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/FM1337/flagbrew-website-source-code/pkg/helpers"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httprate"
	"go.mongodb.org/mongo-driver/bson"
)

func realIPMiddleware(next http.Handler) http.Handler {

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// get the IP
		if rip := helpers.GetIP(r, legacyKey); rip != "" {
			r.RemoteAddr = rip
		}
		next.ServeHTTP(w, r)
	})
}

func registerHTTPRoutes(r chi.Router) {
	// Because it's Vue, serve the index.html when possible.
	// r.With(banCheck, emegerencyMode).Get("/", serveIndex)
	// r.With(banCheck, emegerencyMode).NotFound(serveIndex)

	r.With(banCheck, emegerencyMode).Group(registerAPIV1Routes)
	r.With(banCheck, emegerencyMode).Group(registerAPIV2Routes)
	if cli.Debug {
		r.Group(registerTestRoutes)
	}
}

func banCheck(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// get the IP
		ip := helpers.GetIP(r, legacyKey)
		// If we can't get the IP, something is wrong, so abort
		if ip == "" {
			helpers.HttpError(w, r, http.StatusInternalServerError, fmt.Errorf("a networking error has occurred, please try again later"), false, false, false)
			return
		}

		// Check the bans collection for the user's IP
		result, _ := svcBans.ListBan(r.Context(), bson.M{"ip": ip})
		if result != nil {
			helpers.HttpError(w, r, http.StatusForbidden, fmt.Errorf("%s is banned from accessing Flagbrew's website", r.Host), false, false, false)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func emegerencyMode(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if loadedSettings["emergency_mode"].Value.(bool) {

			// get the IP
			ip := helpers.GetIP(r, legacyKey)
			// If we can't get the IP, something is wrong, so abort
			if ip == "" {
				helpers.HttpError(w, r, http.StatusInternalServerError, fmt.Errorf("a networking error has occurred, please try again later"), false, false, false)
				fmt.Println("error here meme")
				return
			}

			if ip != daemonSetting.GetOwnerIP() && ip != "::1" {
				helpers.HttpError(w, r, http.StatusServiceUnavailable, fmt.Errorf("website is currently in restricted mode, please check back later"), false, false, false)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

// this uses chi's rate limiter but bypasses it if it's a legacy request
func rateLimitByIp(requests int, limit time.Duration) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h := r.Header.Get("Flagbrew-V0-Legacy")
			if h == legacyKey {
				// okay this is a legacy request, we can continue

				next.ServeHTTP(w, r)
				return
			}
			fmt.Println("here")

			httprate.LimitByIP(requests, limit)(next).ServeHTTP(w, r)
		})
	}
}
