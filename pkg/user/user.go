package user

import (
	"context"
	"database/sql"
	"fmt"

	sq "github.com/Masterminds/squirrel"
)

func UserExists(username string, db *sql.DB, ctx context.Context) (bool, error) {
	// Check if the user already exists
	query, args, err := sq.Select("1").
		From("users").
		Where(sq.Eq{"username": username}).
		Limit(1).
		ToSql()

	if err != nil {
		return false, err
	}

	var result int
	err = db.QueryRowContext(ctx, query, args...).Scan(&result)
	if err != nil {
		if err == sql.ErrNoRows {
			// User doesn't exist
			return false, nil
		}

		return false, err
	}

	// Check if the context has been cancelled or timed out
	select {
	case <-ctx.Done():
		return false, fmt.Errorf("It's so over: %w", ctx.Err())
	default:
		return true, nil
	}
}
