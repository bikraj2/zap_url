package main

import (
	"net/http"
)

func (app *application) enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Vary", "Origin")
		w.Header().Add("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Vary", "Access-Control-Request-Method")

		origin := r.Header.Get("Origin")
		if origin != "" && len(app.cfg.cors.trustedOrigins) != 0 {
			for i := range app.cfg.cors.trustedOrigins {
				if origin == app.cfg.cors.trustedOrigins[i] {
					w.Header().Set("Access-Control-Allow-Origin", origin)
					if r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != "" {
						w.Header().Set("Access-Control-Request-Methods", "OPTIONS, PUT, PATCH, DELETE")
						w.Header().Set("Access-Control-Allow-Headers", "Authentication, Content-Type")
						w.WriteHeader(http.StatusOK)
						return
					}
				}
			}
		}
		next.ServeHTTP(w, r)
	})
}
