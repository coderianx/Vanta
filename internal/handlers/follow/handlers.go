package follow

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/coderianx/Vanta/internal/middlewares"
	"github.com/coderianx/Vanta/internal/repository"
	"github.com/go-chi/chi/v5"
)

// FollowHandler creates a follow relationship with the requested user.
func FollowHandler(
	w http.ResponseWriter,
	r *http.Request,
) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		w.WriteHeader(http.StatusMethodNotAllowed)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "method_not_allowed",
		})

		return
	}

	claims, ok := middlewares.GetClaims(r)

	if !ok {
		w.WriteHeader(http.StatusUnauthorized)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "invalid_token",
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

	targetUser, err := repository.GetUserByUsername(
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

	if claims.UserID == targetUser.ID {
		w.WriteHeader(http.StatusBadRequest)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "cannot_follow_self",
		})

		return
	}

	created, err := repository.FollowUser(
		r.Context(),
		claims.UserID,
		targetUser.ID,
	)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "internal_server_error",
		})

		return
	}

	status := http.StatusOK

	if created {
		status = http.StatusCreated
	}

	w.WriteHeader(status)

	json.NewEncoder(w).Encode(map[string]any{
		"following": true,
		"username":  targetUser.Username,
	})
}

// UnfollowHandler removes the authenticated user's follow relationship.
func UnfollowHandler(
	w http.ResponseWriter,
	r *http.Request,
) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodDelete {
		w.Header().Set("Allow", http.MethodDelete)
		w.WriteHeader(http.StatusMethodNotAllowed)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "method_not_allowed",
		})

		return
	}

	claims, ok := middlewares.GetClaims(r)

	if !ok {
		w.WriteHeader(http.StatusUnauthorized)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "invalid_token",
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

	targetUser, err := repository.GetUserByUsername(
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

	if claims.UserID == targetUser.ID {
		w.WriteHeader(http.StatusBadRequest)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "cannot_unfollow_self",
		})

		return
	}

	err = repository.UnfollowUser(
		r.Context(),
		claims.UserID,
		targetUser.ID,
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
		"following": false,
		"username":  targetUser.Username,
	})
}

// GetFollowStatusHandler returns both follow directions between two users.
func GetFollowStatusHandler(
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

	claims, ok := middlewares.GetClaims(r)

	if !ok {
		w.WriteHeader(http.StatusUnauthorized)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "invalid_token",
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

	targetUser, err := repository.GetUserByUsername(
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

	following, followedBy, err := repository.GetFollowStatus(
		r.Context(),
		claims.UserID,
		targetUser.ID,
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
		"following":   following,
		"followed_by": followedBy,
		"mutual":      following && followedBy,
		"username":    targetUser.Username,
	})
}
