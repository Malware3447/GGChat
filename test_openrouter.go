package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Request struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Stream   bool      `json:"stream"`
}

func main() {
	// API endpoint
	url := "https://openrouter.ai/api/v1/chat/completions"

	// Тестовый запрос
	reqBody := Request{
		Model: "anthropic/claude-3-haiku",
		Messages: []Message{
			{Role: "user", Content: "Hello, this is a test request"},
		},
		Stream: false,
	}

	// Конвертируем в JSON
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		fmt.Printf("Ошибка при конвертации в JSON: %v\n", err)
		return
	}

	// Создаем HTTP запрос
	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("Ошибка при создании HTTP запроса: %v\n", err)
		return
	}

	// Устанавливаем заголовки
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer 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")
	httpReq.Header.Set("HTTP-Referer", "https://github.com")
	httpReq.Header.Set("X-Title", "Test Request")

	// Выполняем запрос
	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		fmt.Printf("Ошибка при выполнении запроса: %v\n", err)
		return
	}
	defer resp.Body.Close()

	// Выводим статус ответа
	fmt.Printf("Статус ответа: %s\n", resp.Status)
	fmt.Printf("Код статуса: %d\n", resp.StatusCode)
}
