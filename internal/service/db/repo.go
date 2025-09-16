package db

import (
	"GGChat/internal/db"
	"context"
)

type DbService struct {
	repo db.PgRepository
}

func NewDbService(repo db.PgRepository) *DbService {
	return &DbService{repo: repo}
}

func (ds *DbService) UsersVerification(ctx context.Context, username, password string) (int, error) {
	return ds.repo.UsersVerification(ctx, username, password)
}

func (ds *DbService) NewUser(ctx context.Context, username, password string) (bool, error) {
	return ds.repo.NewUser(ctx, username, password)
}
