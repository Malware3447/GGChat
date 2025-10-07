package db

import (
	"context"

	"github.com/google/uuid"
)

type PgRepository interface {
	UsersVerification(ctx context.Context, username, password string) (int, bool, error)
	NewUser(ctx context.Context, username, password string) (bool, int, error)
	NewChat(ctx context.Context, chatName string) (bool, uuid.UUID, error)
}
