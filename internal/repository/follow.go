package repository

import (
	"context"
	"fmt"

	"github.com/coderianx/Vanta/internal/store"
)

// FollowUser creates a follow relationship if it does not already exist.
// The returned boolean reports whether a new relationship was created.
func FollowUser(
	ctx context.Context,
	followerID int64,
	followedID int64,
) (bool, error) {
	if store.DB == nil {
		return false, errDatabaseNotInitialized
	}

	commandTag, err := store.DB.Exec(
		ctx,
		`
		INSERT INTO follows (
			follower_id,
			followed_id
		)
		VALUES ($1, $2)
		ON CONFLICT (follower_id, followed_id) DO NOTHING
		`,
		followerID,
		followedID,
	)

	if err != nil {
		return false, fmt.Errorf(
			"repository: follow user: %w",
			err,
		)
	}

	return commandTag.RowsAffected() == 1, nil
}

// UnfollowUser removes a follow relationship if it exists.
func UnfollowUser(
	ctx context.Context,
	followerID int64,
	followedID int64,
) error {
	if store.DB == nil {
		return errDatabaseNotInitialized
	}

	_, err := store.DB.Exec(
		ctx,
		`
		DELETE FROM follows
		WHERE follower_id = $1
			AND followed_id = $2
		`,
		followerID,
		followedID,
	)

	if err != nil {
		return fmt.Errorf(
			"repository: unfollow user: %w",
			err,
		)
	}

	return nil
}

// GetFollowStatus reports the follow relationships in both directions.
func GetFollowStatus(
	ctx context.Context,
	userID int64,
	targetUserID int64,
) (bool, bool, error) {
	if store.DB == nil {
		return false, false, errDatabaseNotInitialized
	}

	var following bool
	var followedBy bool

	err := store.DB.QueryRow(
		ctx,
		`
		SELECT
			EXISTS (
				SELECT 1
				FROM follows
				WHERE follower_id = $1
					AND followed_id = $2
			),
			EXISTS (
				SELECT 1
				FROM follows
				WHERE follower_id = $2
					AND followed_id = $1
			)
		`,
		userID,
		targetUserID,
	).Scan(
		&following,
		&followedBy,
	)

	if err != nil {
		return false, false, fmt.Errorf(
			"repository: get follow status: %w",
			err,
		)
	}

	return following, followedBy, nil
}

// GetFollowCounts returns the user's follower and following counts.
func GetFollowCounts(
	ctx context.Context,
	userID int64,
) (int64, int64, error) {
	if store.DB == nil {
		return 0, 0, errDatabaseNotInitialized
	}

	var followerCount int64
	var followingCount int64

	err := store.DB.QueryRow(
		ctx,
		`
		SELECT
			(
				SELECT COUNT(*)
				FROM follows
				WHERE followed_id = $1
			) AS follower_count,
			(
				SELECT COUNT(*)
				FROM follows
				WHERE follower_id = $1
			) AS following_count
		`,
		userID,
	).Scan(
		&followerCount,
		&followingCount,
	)

	if err != nil {
		return 0, 0, fmt.Errorf(
			"repository: get follow counts: %w",
			err,
		)
	}

	return followerCount, followingCount, nil
}
