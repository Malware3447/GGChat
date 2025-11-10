package chats

import (
	"time"

	"github.com/google/uuid"
)

type NewChatRequest struct {
	ChatName string `json:"chat_name"`
	UserName string `json:"user_name"`
}

type Response struct {
	ChatName string    `json:"chat_name"`
	Uuid     uuid.UUID `json:"uuid"`
	Status   bool      `json:"status"`
}

type NewMessageRequest struct {
	ChatId  uuid.UUID `json:"chat_id"`
	Content string    `json:"content"`
}

type GetMessageRequest struct {
	ChatId int `json:"chat_id"`
}

type Chat struct {
	Uuid           uuid.UUID `json:"uuid"`
	Name           string    `json:"name"`
	LastMessage    *string   `json:"last_message"`
	LastMessageKey *string   `json:"last_message_key"`
	UnreadCount    int       `json:"unread_count"`
}

type Message struct {
	MessageId    int       `json:"message_id"`
	UserId       int       `json:"user_id"`
	Content      string    `json:"content"`
	EncryptedKey string    `json:"encrypted_key"`
	Status       string    `json:"status"`
	Time         time.Time `json:"time"`
}
