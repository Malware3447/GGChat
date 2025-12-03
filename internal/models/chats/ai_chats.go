package chats

import "time"

type ChatAI struct {
	Id        int       `json:"id"`
	UserID    int       `json:"user_id"`
	Title     string    `json:"title"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type MessageAI struct {
	Id         int       `json:"id"`
	ChatID     int       `json:"chat_id"`
	Content    string    `json:"content"`
	SenderType string    `json:"sender_type"`
	SentAt     time.Time `json:"sent_at"`
}

type RequestChatAI struct {
	Title string `json:"title"`
}
