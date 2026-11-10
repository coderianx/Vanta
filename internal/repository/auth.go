// Package repository provides database operations for application entities.
package repository

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/coderianx/Vanta/internal/models"
	"github.com/coderianx/Vanta/internal/store"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

const uniqueViolationCode = "23505"

var (
	// ErrUsernameAlreadyExists indicates that a user with the requested username is already registered.
	ErrUsernameAlreadyExists = errors.New("repository: username already exists")

	// ErrUserNotFound indicates that no user matches the requested username.
	ErrUserNotFound = errors.New("repository: user not found")

	errDatabaseNotInitialized = errors.New("repository: database is not initialized")
)

// CreateUser inserts a user with the supplied password hash into the database.
// It returns ErrUsernameAlreadyExists when the username violates the unique constraint.
func CreateUser(ctx context.Context, name, username, passwordHash string) error {
	if store.DB == nil {
		return errDatabaseNotInitialized
	}

	_, err := store.DB.Exec(
		ctx,
		`INSERT INTO users (name, username, password) VALUES ($1, $2, $3)`,
		name,
		username,
		passwordHash,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == uniqueViolationCode {
			return ErrUsernameAlreadyExists
		}

		return fmt.Errorf("repository: create user: %w", err)
	}

	return nil
}

// GetUserByUsername returns the user and password hash for the supplied username.
// It returns ErrUserNotFound when the username is not registered.
func GetUserByUsername(
	ctx context.Context,
	username string,
) (models.User, error) {
	var user models.User

	if store.DB == nil {
		return user, errDatabaseNotInitialized
	}

	err := store.DB.QueryRow(
		ctx,
		`
		SELECT
			id,
			name,
			username,
			password
		FROM users
		WHERE username = $1
		`,
		username,
	).Scan(
		&user.ID,
		&user.Name,
		&user.Username,
		&user.PasswordHash,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return models.User{}, ErrUserNotFound
	}

	if err != nil {
		return models.User{}, fmt.Errorf(
			"repository: get user by username: %w",
			err,
		)
	}

	return user, nil
}

// CreateRefreshToken hashes a refresh token with SHA-256 and stores it for the user.
func CreateRefreshToken(
	ctx context.Context,
	userID int64,
	token string,
	expiresAt time.Time,
) error {
	if store.DB == nil {
		return errDatabaseNotInitialized
	}

	tokenHash := sha256.Sum256(
		[]byte(token),
	)

	_, err := store.DB.Exec(
		ctx,
		`
		INSERT INTO refresh_tokens (
			user_id,
			token_hash,
			expires_at
		)
		VALUES ($1, $2, $3)
		`,
		userID,
		hex.EncodeToString(tokenHash[:]),
		expiresAt,
	)

	if err != nil {
		return fmt.Errorf(
			"repository: create refresh token: %w",
			err,
		)
	}

	return nil
}
