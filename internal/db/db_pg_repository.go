package db

import "context"

type PgRepository interface {
	UsersVerification(ctx context.Context, username, password string) (int, error)
	NewUser(ctx context.Context, username, password string) (bool, error)
}
