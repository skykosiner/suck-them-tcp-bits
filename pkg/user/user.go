package user

import (
	"context"
	"database/sql"
	"fmt"

	sq "github.com/Masterminds/squirrel"
)

type User struct {
	Username string `json:"username" sql:"username"`
}

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

func DeleteUserFromDb(username string, db *sql.DB) error {
	query, args, err := sq.Delete("users").Where(sq.Eq{"username": username}).ToSql()
	if err != nil {
		return err
	}

	if _, err := db.Exec(query, args...); err != nil {
		return err
	}

	return nil
}

func GetUsers(ctx context.Context, db *sql.DB) ([]User, error) {
	var users []User
	rows, err := sq.Select("*").From("users").RunWith(db).Query()
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		var user User
		if err := rows.Scan(&user.Username); err != nil {
			return nil, err
		}

		users = append(users, user)
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
		return users, nil
	}
}
