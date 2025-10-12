package db

import (
	"GGChat/internal/models/chats"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
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

func (repo *RepositoryPg) NewChat(ctx context.Context, chatName string) (bool, uuid.UUID, error) {
	var ChatId uuid.UUID
	const q = `
		INSERT INTO chats (name)
		VALUES ($1)
		returning uuid
	`

	err := repo.db.QueryRow(ctx, q, chatName).Scan(&ChatId)

	if err != nil {
		return false, ChatId, fmt.Errorf("не удалось создать чат: %w", err)
	}

	return true, ChatId, nil
}

func (repo *RepositoryPg) DeleteChat(ctx context.Context, uuid uuid.UUID) error {
	const r = `
		SELECT uuid FROM chats
		WHERE uuid = $1
	`

	result, err := repo.db.Exec(ctx, r, uuid)
	if err != nil {
		return err
	}
	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("Чат не найден или был удален ранее")
	}

	const q = `
		DELETE FROM message
		WHERE chat_id = $1
	`

	_, err = repo.db.Exec(ctx, q, uuid)
	if err != nil {
		return err
	}

	const w = `
		DELETE FROM chat_nembers
		WHERE chat_id = $1
	`

	_, err = repo.db.Exec(ctx, w, uuid)
	if err != nil {
		return err
	}

	const e = `
		DELETE FROM chats
		WHERE uuid = $1
	`

	result, err = repo.db.Exec(ctx, e, uuid)
	if err != nil {
		return err
	}
	rowsAffected = result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("чат не был удален")
	}
	return nil
}

func (repo *RepositoryPg) GetAllChats(ctx context.Context) ([]chats.Chat, error) {
	const q = `
		SELECT uuid, name FROM chats
	`

	rows, err := repo.db.Query(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("не удалось получить список чатов: %w", err)
	}
	defer rows.Close()

	var Chats []chats.Chat
	for rows.Next() {
		var chat chats.Chat
		err := rows.Scan(&chat.Uuid, &chat.Name)
		if err != nil {
			return nil, fmt.Errorf("ошибка при сканировании данных чата: %w", err)
		}
		Chats = append(Chats, chat)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ошибка при итерации по результатам: %w", err)
	}

	return Chats, nil
}
