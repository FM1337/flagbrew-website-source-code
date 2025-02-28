package main

import (
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/httprate"
)

// V1 routes are legacy that is needed to support old versions of PKSM, in the future this will probably be removed
func registerAPIV1Routes(r chi.Router) {
	// r.Route("/pksm", registerPKSMLegacyRoutes)
	// r.Route("/gpss", registerGPSSLegacyRoutes)
	// r.Route("/api/v1", registerAPIV1LegacyRoutes)
	// r.Route("/static/other/gifts", registerMysteryGiftLegacyRoute)
}

// API V2 Routes, additional API version should have their functions tagged with the version in the name
// e.g: registerV3GPSSRoutes and so on.
func registerAPIV2Routes(r chi.Router) {
	r.Route("/api/v2", func(r chi.Router) {
		r.Route("/gpss", registerGPSSRoutes)
		r.Route("/github", registerGitHubRoutes)
		r.Route("/auth", registerAuthRoutes)
		r.Route("/users", registerUsersRoutes)
		r.Route("/user", registerUserRoutes)
		r.Route("/files", registerFileRoutes)
		r.Route("/patreon", registerPatreonRoutes)
		r.With(httprate.LimitByIP(10, 30*time.Second)).Route("/pksm", registerPKSMRoutes)
		r.With(apiAuthRequired).Route("/moderation", registerModerationRoutes)
		r.Route("/metrics", registerMetricRoutes)
	})
}
