package chats

import (
	"github.com/google/uuid"
)

type NewChatRequest struct {
	ChatName string `json:"chat_name"`
}

type Response struct {
	ChatName string    `json:"chat_name"`
	Uuid     uuid.UUID `json:"uuid"`
	Status   bool      `json:"status"`
}

type Chat struct {
	Uuid uuid.UUID `json:"uuid"`
	Name string    `json:"name"`
}

type GetAllChatsResponse struct {
	Chats  []Chat `json:"chats"`
	Status bool   `json:"status"`
}
