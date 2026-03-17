package server

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/lynx-go/lynx-clean-template/internal/pkg/config"
)

// CORSMiddleware creates a CORS middleware based on the provided configuration
func CORSMiddleware(config *config.Http_Cors) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			// Set Allow-Origin header
			if len(config.GetAllowOrigins()) > 0 {
				allowed := false
				for _, allowedOrigin := range config.GetAllowOrigins() {
					if allowedOrigin == "*" || allowedOrigin == origin {
						allowed = true
						w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
						break
					}
				}
				if !allowed && origin != "" {
					// Origin not in whitelist
					http.Error(w, "Origin not allowed", http.StatusForbidden)
					return
				}
			}

			// Set Allow-Methods header
			if len(config.GetAllowMethods()) > 0 {
				w.Header().Set("Access-Control-Allow-Methods", strings.Join(config.GetAllowMethods(), ", "))
			}

			// Set Allow-Headers header
			if len(config.GetAllowHeaders()) > 0 {
				w.Header().Set("Access-Control-Allow-Headers", strings.Join(config.GetAllowHeaders(), ", "))
			}

			// Set Expose-Headers header
			if len(config.GetExposeHeaders()) > 0 {
				w.Header().Set("Access-Control-Expose-Headers", strings.Join(config.GetExposeHeaders(), ", "))
			}

			// Set Allow-Credentials header
			if config.GetAllowCredentials() {
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}

			// Set Max-Age header
			if config.GetMaxAge() > 0 {
				w.Header().Set("Access-Control-Max-Age", strconv.Itoa(int(config.GetMaxAge())))
			}

			// Handle preflight requests
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
