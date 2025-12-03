package main

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
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

type OpenRouterProvider struct {
	APIKey string
	Model  string
}

func (p OpenRouterProvider) GetEndpoint() string {
	return "https://openrouter.ai/api/v1/chat/completions"
}

func (p OpenRouterProvider) GetModel() string {
	if p.Model == "" {
		return "anthropic/claude-3-haiku"
	}
	return p.Model
}

func main() {
	// Тестовое сообщение
	userMessage := "Привет, как дела?"

	// Используем существующую функцию для генерации ответа
	provider := OpenRouterProvider{
		APIKey: "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",
		Model:  "openai/gpt-oss-20b:free",
	}

	conversation := []Message{
		{Role: "system", Content: "Вы полезный ассистент для чата."},
		{Role: "user", Content: userMessage},
	}

	// Преобразуем запрос в JSON
	jsonReq := fmt.Sprintf(`{
		"model": "%s",
		"messages": [
			{"role": "system", "content": "%s"},
			{"role": "user", "content": "%s"}
		],
		"stream": false
	}`, provider.GetModel(), conversation[0].Content, conversation[1].Content)

	httpReq, err := http.NewRequest("POST", provider.GetEndpoint(), strings.NewReader(jsonReq))
	if err != nil {
		fmt.Printf("Ошибка создания запроса: %v\n", err)
		return
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+provider.APIKey)
	httpReq.Header.Set("HTTP-Referer", "https://github.com")
	httpReq.Header.Set("X-Title", "AI Chat Assistant")

	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	resp, err := client.Do(httpReq)
	if err != nil {
		fmt.Printf("Ошибка отправки запроса: %v\n", err)
		return
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Ошибка чтения ответа: %v\n", err)
		return
	}

	fmt.Printf("Ответ от ИИ: %s\n", string(bodyBytes))
}
