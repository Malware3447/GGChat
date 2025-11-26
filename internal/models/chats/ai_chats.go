package chats

import "time"

type ChatAI struct {
	Id        int
	UserID    int
	Title     string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type MessageAI struct {
	Id         int
	ChatID     int
	Content    string
	SenderType string
	SentAt     time.Time
}

type RequestChatAI struct {
	Title string `json:"title"`
}
