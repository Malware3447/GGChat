package db

import "context"

type PgRepository interface {
	UsersVerification(ctx context.Context, username, password string) (int, bool, error)
	NewUser(ctx context.Context, username, password string) (bool, error)
}
