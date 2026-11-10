// Package middlewares provides HTTP middleware for Vanta routes.
package middlewares

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/coderianx/Vanta/internal/models"
	"github.com/golang-jwt/jwt/v5"
)

type claimsContextKey struct{}

var claimsKey claimsContextKey

// AuthMiddleware validates a Bearer JWT and stores its claims in the request context.
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			authorization := r.Header.Get("Authorization")
			parts := strings.Fields(authorization)

			if len(parts) != 2 ||
				!strings.EqualFold(parts[0], "Bearer") ||
				parts[1] == "" {

				w.Header().Set(
					"WWW-Authenticate",
					`Bearer realm="vanta"`,
				)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)

				json.NewEncoder(w).Encode(map[string]any{
					"error": "invalid_token",
				})

				return
			}

			jwtSecret := os.Getenv("JWT_SECRET")

			if jwtSecret == "" {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)

				json.NewEncoder(w).Encode(map[string]any{
					"error": "internal_server_error",
				})

				return
			}

			claims := &models.Claims{}

			token, err := jwt.ParseWithClaims(
				parts[1],
				claims,
				func(token *jwt.Token) (any, error) {
					if token.Method != jwt.SigningMethodHS256 {
						return nil, fmt.Errorf(
							"middlewares: unexpected signing method: %s",
							token.Header["alg"],
						)
					}

					return []byte(jwtSecret), nil
				},
				jwt.WithValidMethods(
					[]string{jwt.SigningMethodHS256.Alg()},
				),
				jwt.WithIssuer("vanta"),
				jwt.WithExpirationRequired(),
			)

			if err != nil || !token.Valid {
				w.Header().Set(
					"WWW-Authenticate",
					`Bearer realm="vanta", error="invalid_token"`,
				)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)

				json.NewEncoder(w).Encode(map[string]any{
					"error": "invalid_token",
				})

				return
			}

			ctx := context.WithValue(
				r.Context(),
				claimsKey,
				claims,
			)

			next.ServeHTTP(
				w,
				r.WithContext(ctx),
			)
		},
	)
}

// GetClaims returns the JWT claims stored by AuthMiddleware.
// The boolean result reports whether authenticated claims were present.
func GetClaims(r *http.Request) (*models.Claims, bool) {
	claims, ok := r.Context().Value(claimsKey).(*models.Claims)

	return claims, ok
}
