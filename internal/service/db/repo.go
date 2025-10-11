package db

import (
	"GGChat/internal/db"
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

func (ds *DbService) NewChat(ctx context.Context, chatName string) (bool, uuid.UUID, error) {
	return ds.repo.NewChat(ctx, chatName)
}

func (ds *DbService) DeleteChat(ctx context.Context, uuid uuid.UUID) error {
	return ds.repo.DeleteChat(ctx, uuid)
}
