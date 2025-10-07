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
