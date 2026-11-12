package store

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

var DB *pgxpool.Pool

func ConnectDatabase() error {
	var err error

	DB_URL := os.Getenv("DB_URL")

	if len(DB_URL) == 0 {
		return errors.New("store:DB_URL is not set in .env")
	}

	DB, err = pgxpool.New(
		context.Background(),
		DB_URL,
	)

	if err != nil {
		return err
	}

	return nil
}

func CreateTables() error {

	_, err := DB.Exec(
		context.Background(),
		`
		CREATE TABLE IF NOT EXISTS users(
            id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
			name TEXT NOT NULL,
            username TEXT NOT NULL UNIQUE,
            password TEXT NOT NULL,
            created_at TIMESTAMPZ NOT NULL DEFAULT NOW()
        );
		`,
	)

	if err != nil {
		return fmt.Errorf("store: create 'users' table error: %w", err)
	}

	_, err = DB.Exec(
		context.Background(),
		`
		CREATE TABLE IF NOT EXISTS refresh_tokens(
			id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
			user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			token_hash CHAR(64) NOT NULL UNIQUE,
			expires_at TIMESTAMPTZ NOT NULL,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);
		`,
	)

	if err != nil {
		return fmt.Errorf("store: create 'refresh_tokens' table error: %w", err)
	}

	_, err = DB.Exec(
		context.Background(),
		`
		CREATE TABLE IF NOT EXISTS follows(
			follower_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			followed_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
			PRIMARY KEY (follower_id, followed_id),
			CONSTRAINT follows_no_self_follow
				CHECK (follower_id <> followed_id)
		);
		`,
	)

	if err != nil {
		return fmt.Errorf("store: create 'follows' table error: %w", err)
	}

	_, err = DB.Exec(
		context.Background(),
		`
		CREATE INDEX IF NOT EXISTS idx_follows_followed_id
			ON follows (followed_id);
		`,
	)

	if err != nil {
		return fmt.Errorf("store: create 'follows' index error: %w", err)
	}

	_, err = DB.Exec(
		context.Background(),
		`
		CREATE TABLE IF NOT EXISTS pulses (
		    id BIGSERIAL PRIMARY KEY,
			user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
		    content TEXT NOT NULL,
		    tags TEXT[] NOT NULL DEFAULT '{}',
			created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);
		`,
	)

	if err != nil {
		return fmt.Errorf(
			"store: create 'pulses' table error: %w", err,
		)
	}

	_, err = DB.Exec(
		context.Background(),
		`
		CREATE INDEX IF NOT EXISTS idx_pulses_user_id_created_at
			ON pulses (user_id, created_at DESC);
		`,
	)

	if err != nil {
		return fmt.Errorf(
			"store: create 'pulses' index error: %w", err,
		)
	}

	return nil
}
