package db

import (
	"GGChat/internal/db"
	"GGChat/internal/models/chats"
	"context"

	"github.com/google/uuid"
)

type DbService struct {
	repo db.PgRepository
}

func NewDbService(repo db.PgRepository) *DbService {
	return &DbService{repo: repo}
}

func (ds *DbService) UsersVerification(ctx context.Context, username, password string) (int, bool, error) {
	return ds.repo.UsersVerification(ctx, username, password)
}

func (ds *DbService) NewUser(ctx context.Context, username, password string) (bool, int, error) {
	return ds.repo.NewUser(ctx, username, password)
}

func (ds *DbService) NewChat(ctx context.Context, chatName string, UserId int, other_user_id int) (bool, uuid.UUID, error) {
	return ds.repo.NewChat(ctx, chatName, UserId, other_user_id)
}

func (ds *DbService) DeleteChat(ctx context.Context, uuid uuid.UUID) error {
	return ds.repo.DeleteChat(ctx, uuid)
}

func (ds *DbService) GetAllChats(ctx context.Context, UserId int) ([]chats.Chat, error) {
	return ds.repo.GetAllChats(ctx, UserId)
}

func (ds *DbService) GetUser(ctx context.Context, username string) (int, error) {
	return ds.repo.GetUser(ctx, username)
}

func (ds *DbService) NewMessage(ctx context.Context, chatId uuid.UUID, senderId int, encryptedContent string, encryptedKeys map[int]string) (int, string, error) {
	return ds.repo.NewMessage(ctx, chatId, senderId, encryptedContent, encryptedKeys)
}

func (ds *DbService) GetMessage(ctx context.Context, chatId uuid.UUID, currentUserId int) ([]chats.Message, error) {
	return ds.repo.GetMessage(ctx, chatId, currentUserId)
}

func (ds *DbService) UpdateMessageStatus(ctx context.Context, messageId int, status string) error {
	return ds.repo.UpdateMessageStatus(ctx, messageId, status)
}

func (ds *DbService) AddPublicKey(ctx context.Context, userId int, publicKey string) error {
	return ds.repo.AddPublicKey(ctx, userId, publicKey)
}

func (ds *DbService) GetPublicKeysForChat(ctx context.Context, chatId uuid.UUID, senderId int) (map[int]string, error) {
	return ds.repo.GetPublicKeysForChat(ctx, chatId, senderId)
}
