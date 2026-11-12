package pulse

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/coderianx/Vanta/internal/middlewares"
	"github.com/coderianx/Vanta/internal/models"
	"github.com/coderianx/Vanta/internal/repository"
	"github.com/go-chi/chi/v5"
)

func CreateNewPulse(
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

	var req models.CreateNewPulseRequest

	err := json.NewDecoder(r.Body).Decode(&req)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "invalid_request",
		})

		return
	}

	req.Content = strings.TrimSpace(req.Content)

	if req.Content == "" {
		w.WriteHeader(http.StatusBadRequest)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "content_is_required",
		})

		return
	}

	newPulse, err := repository.CreatePulse(
		r.Context(),
		claims.UserID,
		req.Content,
		req.Tags,
	)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "internal_server_error",
		})

		return
	}

	w.WriteHeader(http.StatusCreated)

	json.NewEncoder(w).Encode(newPulse)
}

// GetUserPulses returns the pulses published by the requested username.
func GetUserPulses(
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

	pulses, err := repository.GetUserPulses(
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

	json.NewEncoder(w).Encode(pulses)
}

// GetPulseByID returns a pulse using the ID supplied in the URL.
func GetPulseByID(
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

	idParam := strings.TrimSpace(
		chi.URLParam(r, "id"),
	)

	pulseID, err := strconv.ParseInt(
		idParam,
		10,
		64,
	)

	if err != nil || pulseID <= 0 {
		w.WriteHeader(http.StatusBadRequest)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "invalid_pulse_id",
		})

		return
	}

	foundPulse, err := repository.GetPulseByID(
		r.Context(),
		pulseID,
	)

	if errors.Is(err, repository.ErrPulseNotFound) {
		w.WriteHeader(http.StatusNotFound)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "pulse_not_found",
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

	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(foundPulse)
}

// DeleteMyPulse deletes an owned pulse using the ID supplied in the URL.
func DeleteMyPulse(
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

	idParam := strings.TrimSpace(
		chi.URLParam(r, "id"),
	)

	pulseID, err := strconv.ParseInt(
		idParam,
		10,
		64,
	)

	if err != nil || pulseID <= 0 {
		w.WriteHeader(http.StatusBadRequest)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "invalid_pulse_id",
		})

		return
	}

	deleted, err := repository.DeletePulseByOwner(
		r.Context(),
		pulseID,
		claims.UserID,
	)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "internal_server_error",
		})

		return
	}

	if !deleted {
		w.WriteHeader(http.StatusNotFound)

		json.NewEncoder(w).Encode(map[string]any{
			"error": "pulse_not_found",
		})

		return
	}

	w.WriteHeader(http.StatusNoContent)
}
