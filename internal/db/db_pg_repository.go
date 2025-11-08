package db

import (
	"context"

	"GGChat/internal/models/chats"

	"github.com/google/uuid"
)

type PgRepository interface {
	UsersVerification(ctx context.Context, username, password string) (int, bool, error)
	NewUser(ctx context.Context, username, password string) (bool, int, error)
	NewChat(ctx context.Context, chatName string, UserId int, other_user_id int) (bool, uuid.UUID, error)
	DeleteChat(ctx context.Context, uuid uuid.UUID) error
	GetAllChats(ctx context.Context, UserId int) ([]chats.Chat, error)
	GetUser(ctx context.Context, username string) (int, error)
	NewMessage(ctx context.Context, chatId uuid.UUID, senderId int, content string) (int, string, error)
	GetMessage(ctx context.Context, chatId uuid.UUID, currentUserId int) ([]chats.Message, error)
	UpdateMessageStatus(ctx context.Context, messageId int, status string) error
}
