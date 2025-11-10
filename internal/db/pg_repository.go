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
	tx, err := repo.db.Begin(ctx)
	if err != nil {
		return false, -1, fmt.Errorf("не удалось начать транзакцию: %w", err)
	}
	defer tx.Rollback(ctx)

	const q = `
	INSERT INTO users (username, password)
	VALUES ($1, $2)
	`

	_, err = tx.Exec(ctx, q, username, password)

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

	err = tx.QueryRow(ctx, b, username).Scan(&id)

	if errors.Is(err, pgx.ErrNoRows) {
		return false, -1, nil
	}

	if err != nil {
		return false, -1, fmt.Errorf("ошибка при поиске данных в БД: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return false, -1, fmt.Errorf("не удалось закоммитить транзакцию: %w", err)
	}

	return true, id, nil
}

func (repo *RepositoryPg) NewChat(ctx context.Context, chatName string, UserId int, other_user_id int) (bool, uuid.UUID, error) {
	tx, err := repo.db.Begin(ctx)
	if err != nil {
		return false, uuid.UUID{}, fmt.Errorf("не удалось начать транзакцию: %w", err)
	}
	defer tx.Rollback(ctx)

	var ChatId uuid.UUID
	const q = `
		INSERT INTO chats (name)
		VALUES ($1)
		returning uuid
	`

	err = tx.QueryRow(ctx, q, chatName).Scan(&ChatId)
	if err != nil {
		return false, ChatId, fmt.Errorf("не удалось создать чат: %w", err)
	}

	const p = `
		INSERT INTO chat_nembers (chat_id, user_id)
		VALUES ($1, $2)
	`

	_, err = tx.Exec(ctx, p, ChatId, UserId)
	if err != nil {
		return false, ChatId, fmt.Errorf("не удалось создать чат: %w", err)
	}

	const b = `
		INSERT INTO chat_nembers (chat_id, user_id)
		VALUES ($1, $2)
	`

	_, err = tx.Exec(ctx, b, ChatId, other_user_id)
	if err != nil {
		return false, ChatId, fmt.Errorf("не удалось создать чат: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return false, ChatId, fmt.Errorf("не удалось закоммитить транзакцию: %w", err)
	}

	return true, ChatId, nil
}

func (repo *RepositoryPg) DeleteChat(ctx context.Context, uuid uuid.UUID) error {
	tx, err := repo.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("не удалось начать транзакцию: %w", err)
	}
	defer tx.Rollback(ctx)

	const r = `
		SELECT uuid FROM chats
		WHERE uuid = $1
	`

	result, err := tx.Exec(ctx, r, uuid)
	if err != nil {
		return err
	}
	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("чат не найден или был удален ранее")
	}

	const q = `
		DELETE FROM message
		WHERE chat_id = $1
	`

	_, err = tx.Exec(ctx, q, uuid)
	if err != nil {
		return err
	}

	const w = `
		DELETE FROM chat_nembers
		WHERE chat_id = $1
	`

	_, err = tx.Exec(ctx, w, uuid)
	if err != nil {
		return err
	}

	const e = `
		DELETE FROM chats
		WHERE uuid = $1
	`

	result, err = tx.Exec(ctx, e, uuid)
	if err != nil {
		return err
	}
	rowsAffected = result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("чат не был удален")
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("не удалось закоммитить транзакцию: %w", err)
	}
	return nil
}

// pg_repository.go
func (repo *RepositoryPg) GetAllChats(ctx context.Context, UserId int) ([]chats.Chat, error) {
	const q = `
		SELECT 
			c.uuid, 
			c.name,
			(
				SELECT COUNT(*)
				FROM message_status ms
				JOIN message m ON ms.message_id = m.id
				WHERE m.chat_id = c.uuid      
				AND ms.user_id = $1           
				AND ms.status != 'read'       
				AND m.sender_id != $1         
			) AS unread_count,
			-- Используем LATERAL JOIN для получения последнего сообщения и его ключа
			lm.content AS last_message,
			lmk.encrypted_key AS last_message_key
		FROM chats c
		JOIN chat_nembers cn ON c.uuid = cn.chat_id
		
		-- Находим последнее сообщение (lm)
		LEFT JOIN LATERAL (
			SELECT m.id, m.content, m.sent_at
			FROM message m
			WHERE m.chat_id = c.uuid
			ORDER BY m.sent_at DESC
			LIMIT 1
		) lm ON true
		
		-- Находим ключ (lmk) ДЛЯ ЭТОГО пользователя и ДЛЯ ЭТОГО сообщения
		LEFT JOIN LATERAL (
			SELECT mk.encrypted_key
			FROM message_keys mk
			WHERE mk.message_id = lm.id AND mk.user_id = $1
			LIMIT 1
		) lmk ON true
		
		WHERE cn.user_id = $1
		-- Сортируем по дате последнего сообщения
		ORDER BY lm.sent_at DESC NULLS LAST;
	`

	rows, err := repo.db.Query(ctx, q, UserId)
	if err != nil {
		return nil, fmt.Errorf("не удалось получить список чатов пользователя: %w", err)
	}
	defer rows.Close()

	var Chats []chats.Chat
	for rows.Next() {
		var chat chats.Chat
		// Теперь сканируем 5 полей, включая новое LastMessageKey
		err := rows.Scan(&chat.Uuid, &chat.Name, &chat.UnreadCount, &chat.LastMessage, &chat.LastMessageKey)
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

func (repo *RepositoryPg) GetUser(ctx context.Context, username string) (int, error) {
	var user_id int
	const q = `
		SELECT id FROM users
		WHERE username = $1
	`

	err := repo.db.QueryRow(ctx, q, username).Scan(&user_id)
	if err != nil {
		return -1, fmt.Errorf("ошибка при получение списка всех пользователей: %v", err)
	}

	return user_id, nil
}

func (repo *RepositoryPg) NewMessage(ctx context.Context, chatId uuid.UUID, senderId int, encryptedContent string, encryptedKeys map[int]string) (int, string, error) {
	tx, err := repo.db.Begin(ctx)
	if err != nil {
		return -1, "", fmt.Errorf("не удалось начать транзакцию: %w", err)
	}
	defer tx.Rollback(ctx)

	const q = `
		INSERT INTO message (chat_id, sender_id, content, sent_at)
		VALUES ($1, $2, $3, NOW())
		RETURNING id
	`
	var messageId int
	err = tx.QueryRow(ctx, q, chatId, senderId, encryptedContent).Scan(&messageId)
	if err != nil {
		return -1, "", fmt.Errorf("ошибка при отправке сообщения: %v", err)
	}

	const qKey = `
        INSERT INTO message_keys (message_id, user_id, encrypted_key)
        VALUES ($1, $2, $3)
    `

	for userId, encKey := range encryptedKeys {
		if _, err := tx.Exec(ctx, qKey, messageId, userId, encKey); err != nil {
			return -1, "", err
		}
	}

	const qMembers = `
		SELECT user_id FROM chat_nembers
		WHERE chat_id = $1
	`
	rows, err := tx.Query(ctx, qMembers, chatId)
	if err != nil {
		return -1, "", fmt.Errorf("ошибка при получении участников чата: %v", err)
	}

	var memberIds []int
	for rows.Next() {
		var userId int
		if err := rows.Scan(&userId); err != nil {
			rows.Close()
			return -1, "", fmt.Errorf("ошибка сканирования user_id: %v", err)
		}
		memberIds = append(memberIds, userId)
	}

	rows.Close()

	if err := rows.Err(); err != nil {
		return -1, "", fmt.Errorf("ошибка при итерации по участникам: %v", err)
	}

	const qStatus = `
		INSERT INTO message_status (message_id, user_id, status)
		VALUES ($1, $2, $3)
	`

	senderStatus := "read"

	for _, userId := range memberIds {
		var status string
		if userId == senderId {
			status = senderStatus
		} else {
			status = "delivered"
		}

		if _, err := tx.Exec(ctx, qStatus, messageId, userId, status); err != nil {
			return -1, "", fmt.Errorf("ошибка при установке статуса сообщения: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return -1, "", fmt.Errorf("не удалось закоммитить транзакцию: %w", err)
	}

	return messageId, senderStatus, nil
}
func (repo *RepositoryPg) GetMessage(ctx context.Context, chatId uuid.UUID, currentUserId int) ([]chats.Message, error) {
	tx, err := repo.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("не удалось начать транзакцию: %w", err)
	}
	defer tx.Rollback(ctx)

	const q = `
        SELECT m.id, m.sender_id, m.content, ms.status, m.sent_at,
               COALESCE(mk.encrypted_key, '') as encrypted_key
        FROM message m
        JOIN message_status ms ON m.id = ms.message_id
        LEFT JOIN message_keys mk ON m.id = mk.message_id AND mk.user_id = $2
        WHERE m.chat_id = $1 AND ms.user_id = $2
        ORDER BY m.id ASC
    `

	rows, err := tx.Query(ctx, q, chatId, currentUserId)
	if err != nil {
		return nil, fmt.Errorf("не получилось получить список сообщений: %v", err)
	}
	defer rows.Close()

	var result []chats.Message
	for rows.Next() {
		var message chats.Message
		err := rows.Scan(
			&message.MessageId,
			&message.UserId,
			&message.Content,
			&message.Status,
			&message.Time,
			&message.EncryptedKey,
		)
		if err != nil {
			return nil, fmt.Errorf("ошибка при сканировании: %v", err)
		}
		result = append(result, message)
	}

	const w = `
        UPDATE message_status
        SET status = 'read'
        WHERE message_id IN (
            SELECT id FROM message
            WHERE chat_id = $1 AND sender_id != $2
        )
        AND status != 'read'
    `

	_, err = tx.Exec(ctx, w, chatId, currentUserId)
	if err != nil {
		return nil, fmt.Errorf("ошибка при установке статуса 'read': %v", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("не удалось закоммитить транзакцию: %w", err)
	}

	return result, nil
}

func (repo *RepositoryPg) UpdateMessageStatus(ctx context.Context, messageId int, status string) error {
	tx, err := repo.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("не удалось начать транзакцию: %w", err)
	}
	defer tx.Rollback(ctx)

	const q = `
		UPDATE message_status
		SET status = $1
		WHERE message_id = $2 AND status != 'read'
	`
	_, err = tx.Exec(ctx, q, status, messageId)
	if err != nil {
		return fmt.Errorf("ошибка при обновлении статуса сообщения: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("не удалось закоммитить транзакцию: %w", err)
	}
	return nil
}

func (repo *RepositoryPg) AddPublicKey(ctx context.Context, userId int, publicKey string) error {
	const q = `UPDATE users SET public_key = $1 WHERE id = $2`
	_, err := repo.db.Exec(ctx, q, publicKey, userId)
	return err
}

func (repo *RepositoryPg) GetPublicKeysForChat(ctx context.Context, chatId uuid.UUID, senderId int) (map[int]string, error) {
	const q = `
        SELECT u.id, u.public_key 
        FROM users u
        JOIN chat_nembers cn ON u.id = cn.user_id
        WHERE cn.chat_id = $1 AND u.id != $2
    `
	rows, err := repo.db.Query(ctx, q, chatId, senderId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	keys := make(map[int]string)
	for rows.Next() {
		var userId int
		var publicKey *string
		if err := rows.Scan(&userId, &publicKey); err != nil {
			return nil, err
		}
		if publicKey != nil {
			keys[userId] = *publicKey
		}
	}
	return keys, nil
}
