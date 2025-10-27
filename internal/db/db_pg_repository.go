package db

import (
	"context"

	"GGChat/internal/models/chats"

	"github.com/google/uuid"
)

type PgRepository interface {
	UsersVerification(ctx context.Context, username, password string) (int, bool, error)
	NewUser(ctx context.Context, username, password string) (bool, int, error)
	NewChat(ctx context.Context, chatName string, UserId int) (bool, uuid.UUID, error)
	DeleteChat(ctx context.Context, uuid uuid.UUID) error
	GetAllChats(ctx context.Context, UserId int) ([]chats.Chat, error)
}
