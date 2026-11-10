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

	// ErrUserNotFound indicates that no user matches the requested identifier.
	ErrUserNotFound = errors.New("repository: user not found")

	// ErrInvalidRefreshToken indicates that a refresh token is unknown or expired.
	ErrInvalidRefreshToken = errors.New("repository: invalid refresh token")

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

// GetUserByID returns public account data for the supplied user ID.
// It returns ErrUserNotFound when the user does not exist.
func GetUserByID(
	ctx context.Context,
	userID int64,
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
			username
		FROM users
		WHERE id = $1
		`,
		userID,
	).Scan(
		&user.ID,
		&user.Name,
		&user.Username,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return models.User{}, ErrUserNotFound
	}

	if err != nil {
		return models.User{}, fmt.Errorf(
			"repository: get user by ID: %w",
			err,
		)
	}

	return user, nil
}

// RotateRefreshToken replaces a valid refresh token with a new one atomically.
// It returns ErrInvalidRefreshToken when the old token is unknown or expired.
func RotateRefreshToken(
	ctx context.Context,
	oldToken string,
	newToken string,
	newTokenExpiresAt time.Time,
) (models.User, error) {
	var user models.User

	if store.DB == nil {
		return user, errDatabaseNotInitialized
	}

	oldTokenHash := sha256.Sum256(
		[]byte(oldToken),
	)

	newTokenHash := sha256.Sum256(
		[]byte(newToken),
	)

	tx, err := store.DB.Begin(ctx)

	if err != nil {
		return models.User{}, fmt.Errorf(
			"repository: begin refresh token rotation: %w",
			err,
		)
	}

	defer func() {
		_ = tx.Rollback(ctx)
	}()

	var refreshTokenID int64

	err = tx.QueryRow(
		ctx,
		`
		SELECT
			rt.id,
			u.id,
			u.name,
			u.username
		FROM refresh_tokens AS rt
		INNER JOIN users AS u ON u.id = rt.user_id
		WHERE rt.token_hash = $1
			AND rt.expires_at > NOW()
		FOR UPDATE OF rt
		`,
		hex.EncodeToString(oldTokenHash[:]),
	).Scan(
		&refreshTokenID,
		&user.ID,
		&user.Name,
		&user.Username,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return models.User{}, ErrInvalidRefreshToken
	}

	if err != nil {
		return models.User{}, fmt.Errorf(
			"repository: get refresh token for rotation: %w",
			err,
		)
	}

	_, err = tx.Exec(
		ctx,
		`DELETE FROM refresh_tokens WHERE id = $1`,
		refreshTokenID,
	)

	if err != nil {
		return models.User{}, fmt.Errorf(
			"repository: delete old refresh token: %w",
			err,
		)
	}

	_, err = tx.Exec(
		ctx,
		`
		INSERT INTO refresh_tokens (
			user_id,
			token_hash,
			expires_at
		)
		VALUES ($1, $2, $3)
		`,
		user.ID,
		hex.EncodeToString(newTokenHash[:]),
		newTokenExpiresAt,
	)

	if err != nil {
		return models.User{}, fmt.Errorf(
			"repository: store rotated refresh token: %w",
			err,
		)
	}

	if err = tx.Commit(ctx); err != nil {
		return models.User{}, fmt.Errorf(
			"repository: commit refresh token rotation: %w",
			err,
		)
	}

	return user, nil
}
