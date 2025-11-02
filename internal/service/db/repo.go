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

func (ds *DbService) NewMessage(ctx context.Context, chatId uuid.UUID, senderId int, content string) error {
	return ds.repo.NewMessage(ctx, chatId, senderId, content)
}

func (ds *DbService) GetMessage(ctx context.Context, chatId uuid.UUID) ([]chats.Message, error) {
	return ds.repo.GetMessage(ctx, chatId)
}
