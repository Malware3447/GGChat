package db

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RepositoryPg struct {
	db *pgxpool.Pool
}

func NewRepositoryPg(db *pgxpool.Pool) PgRepository {
	return &RepositoryPg{db: db}
}

func (repo *RepositoryPg) UsersVerification(ctx context.Context, username, password string) (int, bool, error) {
	const q = `
	SELECT id, password FROM users
	WHERE username = $1
	`

	var storedPassword string
	var id int
	err := repo.db.QueryRow(ctx, q, username).Scan(&id, &storedPassword)

	if errors.Is(err, pgx.ErrNoRows) {
		return -1, false, nil
	}

	if err != nil {
		return -1, false, fmt.Errorf("ошибка при поиске данных в БД: %w", err)
	}

	if storedPassword != password {
		return -1, false, nil
	}

	return id, true, nil
}

func (repo *RepositoryPg) NewUser(ctx context.Context, username, password string) (bool, int, error) {
	const q = `
	INSERT INTO users (username, password)
	VALUES ($1, $2)
	`

	_, err := repo.db.Exec(ctx, q, username, password)

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return false, -1, fmt.Errorf("пользователь с таким именем уже существует: %w", err)
	}

	if err != nil {
		return false, -1, fmt.Errorf("не удалось вставить нового пользователя: %w", err)
	}

	const b = `
	SELECT id FROM users
	WHERE username = $1
	`

	var id int

	err = repo.db.QueryRow(ctx, b, username).Scan(&id)

	if errors.Is(err, pgx.ErrNoRows) {
		return false, -1, nil
	}

	if err != nil {
		return false, -1, fmt.Errorf("ошибка при поиске данных в БД: %w", err)
	}

	return true, id, nil
}
