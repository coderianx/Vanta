package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/coderianx/Vanta/internal/models"
	"github.com/coderianx/Vanta/internal/store"
	"github.com/jackc/pgx/v5"
)

// ErrPulseNotFound indicates that no pulse matches the requested ID.
var ErrPulseNotFound = errors.New("repository: pulse not found")

// CreatePulse stores a new pulse and returns its persisted representation.
func CreatePulse(
	ctx context.Context,
	userID int64,
	content string,
	tags []string,
) (models.Pulse, error) {
	var pulse models.Pulse

	if store.DB == nil {
		return pulse, errDatabaseNotInitialized
	}

	if tags == nil {
		tags = []string{}
	}

	err := store.DB.QueryRow(
		ctx,
		`
		INSERT INTO pulses (
			user_id,
			content,
			tags
		)
		VALUES ($1, $2, $3)
		RETURNING
			id,
			user_id,
			content,
			tags,
			created_at
		`,
		userID,
		content,
		tags,
	).Scan(
		&pulse.ID,
		&pulse.UserID,
		&pulse.Content,
		&pulse.Tags,
		&pulse.CreatedAt,
	)

	if err != nil {
		return models.Pulse{}, fmt.Errorf(
			"repository: create pulse: %w",
			err,
		)
	}

	return pulse, nil
}

// GetUserPulses returns all pulses belonging to a user, newest first.
func GetUserPulses(
	ctx context.Context,
	userID int64,
) ([]models.Pulse, error) {
	pulses := make([]models.Pulse, 0)

	if store.DB == nil {
		return pulses, errDatabaseNotInitialized
	}

	rows, err := store.DB.Query(
		ctx,
		`
		SELECT
			id,
			user_id,
			content,
			tags,
			created_at
		FROM pulses
		WHERE user_id = $1
		ORDER BY created_at DESC, id DESC
		`,
		userID,
	)

	if err != nil {
		return nil, fmt.Errorf(
			"repository: get user pulses: %w",
			err,
		)
	}

	defer rows.Close()

	for rows.Next() {
		var pulse models.Pulse

		err = rows.Scan(
			&pulse.ID,
			&pulse.UserID,
			&pulse.Content,
			&pulse.Tags,
			&pulse.CreatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf(
				"repository: scan user pulse: %w",
				err,
			)
		}

		pulses = append(pulses, pulse)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf(
			"repository: iterate user pulses: %w",
			err,
		)
	}

	return pulses, nil
}

// DeletePulseByOwner deletes a pulse only when it belongs to the supplied user.
// The returned boolean reports whether a pulse was deleted.
func DeletePulseByOwner(
	ctx context.Context,
	pulseID int64,
	userID int64,
) (bool, error) {
	if store.DB == nil {
		return false, errDatabaseNotInitialized
	}

	commandTag, err := store.DB.Exec(
		ctx,
		`
		DELETE FROM pulses
		WHERE id = $1
			AND user_id = $2
		`,
		pulseID,
		userID,
	)

	if err != nil {
		return false, fmt.Errorf(
			"repository: delete pulse by owner: %w",
			err,
		)
	}

	return commandTag.RowsAffected() == 1, nil
}

// GetPulseByID returns the pulse with the supplied ID.
// It returns ErrPulseNotFound when the pulse does not exist.
func GetPulseByID(
	ctx context.Context,
	pulseID int64,
) (models.Pulse, error) {
	var pulse models.Pulse

	if store.DB == nil {
		return pulse, errDatabaseNotInitialized
	}

	err := store.DB.QueryRow(
		ctx,
		`
		SELECT
			id,
			user_id,
			content,
			tags,
			created_at
		FROM pulses
		WHERE id = $1
		`,
		pulseID,
	).Scan(
		&pulse.ID,
		&pulse.UserID,
		&pulse.Content,
		&pulse.Tags,
		&pulse.CreatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return models.Pulse{}, ErrPulseNotFound
	}

	if err != nil {
		return models.Pulse{}, fmt.Errorf(
			"repository: get pulse by ID: %w",
			err,
		)
	}

	return pulse, nil
}
