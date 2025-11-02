package chats

import (
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
	Uuid uuid.UUID `json:"uuid"`
	Name string    `json:"name"`
}

type Message struct {
	UserId  int    `json:"user_id"`
	Content string `json:"content"`
}
