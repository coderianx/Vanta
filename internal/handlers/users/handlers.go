// Package users provides HTTP handlers for public user profiles.
package users

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/coderianx/Vanta/internal/repository"
	"github.com/go-chi/chi/v5"
)

// GetUserProfileHandler returns a user's public profile and follow counts.
func GetUserProfileHandler(
	w http.ResponseWriter,
	r *http.Request,
) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		w.WriteHeader(http.StatusMethodNotAllowed)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "method_not_allowed",
		})

		return
	}

	username := strings.TrimSpace(
		chi.URLParam(r, "username"),
	)

	if username == "" {
		w.WriteHeader(http.StatusBadRequest)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "invalid_request",
		})

		return
	}

	user, err := repository.GetUserByUsername(
		r.Context(),
		username,
	)

	if errors.Is(err, repository.ErrUserNotFound) {
		w.WriteHeader(http.StatusNotFound)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "user_not_found",
		})

		return
	}

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "internal_server_error",
		})

		return
	}

	followerCount, followingCount, err := repository.GetFollowCounts(
		r.Context(),
		user.ID,
	)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "internal_server_error",
		})

		return
	}

	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(map[string]any{
		"id":              user.ID,
		"name":            user.Name,
		"username":        user.Username,
		"follower_count":  followerCount,
		"following_count": followingCount,
	})
}
