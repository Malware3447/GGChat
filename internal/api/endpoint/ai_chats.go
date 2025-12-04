package endpoint

import (
	"GGChat/internal/models/chats"
	"GGChat/internal/service/db"
	MyWS "GGChat/internal/websocket"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/sirupsen/logrus"
)

type AIApiChats struct {
	repo             *db.DbService
	gm               *gRPCMethod
	WebsocketManager *MyWS.Manager
}

func NewAIApiChats(repo *db.DbService, wsManager *MyWS.Manager) *AIApiChats {
	return &AIApiChats{
		repo:             repo,
		WebsocketManager: wsManager,
	}
}

func (a *AIApiChats) GetAllChatsAI(w http.ResponseWriter, r *http.Request) {
	UserID, ok := r.Context().Value("user_id").(int)
	if !ok {
		logrus.Warn("Ошибка получения ID пользователя из контекста")
		http.Error(w, "Error get chats", http.StatusBadRequest)
		return
	}

	response, err := a.repo.GetAllChatsAI(context.Background(), UserID)
	if err != nil {
		logrus.Warn("ошибка получения списка чатов: ", err)
		http.Error(w, "Error get list chats", http.StatusBadRequest)
	}

	w.Header().Set("Content-Type", "application-json")
	w.WriteHeader(http.StatusOK)
	if err = json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func (a *AIApiChats) DeleteChatAI(w http.ResponseWriter, r *http.Request) {
	IDStr := chi.URLParam(r, "id")
	ID, err := strconv.Atoi(IDStr)
	if err != nil {
		logrus.Warn("Ошибка конвертации ID: ", err)
		http.Error(w, "Error converting ID", http.StatusBadRequest)
		return
	}

	err = a.repo.DeleteChatAI(context.Background(), ID)
	if err != nil {
		logrus.Warn("ошибка удаления чата: ", err)
		http.Error(w, "Error ", http.StatusBadRequest)
	}
	w.Header().Set("Content-Type", "application-json")
	w.WriteHeader(http.StatusOK)
	if err = json.NewEncoder(w).Encode(nil); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func (a *AIApiChats) CreateChat(w http.ResponseWriter, r *http.Request) {
	UserID, ok := r.Context().Value("user_id").(int)
	if !ok {
		logrus.Warn("Ошибка получения ID пользователя из контекста")
		http.Error(w, "Error get user ID", http.StatusBadRequest)
		return
	}

	body := chats.RequestChatAI{}
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		logrus.Warn("Неверное тело запроса.")
		http.Error(w, "Invalid body request", http.StatusBadRequest)
		return
	}

	chat, err := a.repo.CreateChatAI(context.Background(), UserID, body.Title)
	if err != nil {
		logrus.Warn("ошибка создания чата: ", err)
		http.Error(w, "ошибка создания чата", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application-json")
	w.WriteHeader(http.StatusOK)
	if err = json.NewEncoder(w).Encode(chat); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func (a *AIApiChats) NewMessage(w http.ResponseWriter, r *http.Request) {
	UserID, ok := r.Context().Value("user_id").(int)
	if !ok {
		logrus.Warn("Ошибка получения ID пользователя из контекста")
		http.Error(w, "Error get chats", http.StatusBadRequest)
		return
	}

	body := chats.MessageAI{}
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		logrus.Warn("Неверное тело запроса.")
		http.Error(w, "Invalid body request", http.StatusBadRequest)
		return
	}
	body.SenderType = "user"

	// Создание чата если его нет
	if body.ChatID == 0 {
		firstMessage := body.Content
		if len(firstMessage) > 20 {
			firstMessage = firstMessage[:20] + "..."
		}
		chat, err := a.repo.CreateChatAI(context.Background(), UserID, firstMessage)
		if err != nil {
			logrus.Warn("ошибка создания чата")
			http.Error(w, "ошибка создания чата", http.StatusBadRequest)
			return
		}
		body.ChatID = chat.Id
	}

	// Сохраняем сообщение пользователя
	response, err := a.repo.CreateMessage(context.Background(), body.ChatID, body.SenderType, body.Content)
	if err != nil {
		logrus.Warn("ошибка отправки сообщения: ", err)
		http.Error(w, "ошибка отправки сообщения", http.StatusBadRequest)
		return
	}

	// Генерация ответа от AI
	go a.GenerateAIResponse(body.ChatID)

	w.Header().Set("Content-Type", "application-json")
	w.WriteHeader(http.StatusOK)
	if err = json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func (a *AIApiChats) GetMessages(w http.ResponseWriter, r *http.Request) {
	chatID, err := strconv.Atoi(chi.URLParam(r, "chat_id"))
	if err != nil {
		logrus.Warn("Ошибка конвертации ID чата: ", err)
		http.Error(w, "Error converting chat ID", http.StatusBadRequest)
		return
	}

	messages, err := a.repo.GetMessageInChat(context.Background(), chatID, 100, nil)
	if err != nil {
		logrus.Warn("ошибка получения сообщений: ", err)
		http.Error(w, "Error get messages", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application-json")
	w.WriteHeader(http.StatusOK)
	if err = json.NewEncoder(w).Encode(messages); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func (a *AIApiChats) GenerateAIResponse(chatID int) {
	messages, err := a.repo.GetMessageInChat(context.Background(), chatID, 100, nil)
	if err != nil {
		logrus.Warn("ошибка получения истории сообщений для контекста: ", err)
		messages = []chats.MessageAI{}
	}

	var history []Message

	for _, msg := range messages {
		role := "user"

		if msg.SenderType == "ai" {
			role = "assistant"
		}

		history = append(history, Message{
			Role:    role,
			Content: msg.Content,
		})
	}

	aiResponse, docName, match, err := GenerateAIResponse(history)
	if err != nil {
		logrus.Warn("Ошибка генерации AI ответа: ", err)
		aiResponse = "Извините, произошла ошибка при обработке запроса."
	}

	if docName != "" {
		doneDocUrl, err := a.gm.DocGenerator(match, docName)
		if err != nil {
			return
		}

		_, err = a.gm.ConPDF(doneDocUrl)
		if err != nil {
			return
		}

		_, err = a.repo.CreateMessage(context.Background(), chatID, "ai", "/api/v1/ai_chats/download_doc/"+docName)
	}

	_, err = a.repo.CreateMessage(context.Background(), chatID, "ai", aiResponse)
	if err != nil {
		logrus.Warn("ошибка создания сообщения от AI: ", err)
		return
	}

	fmt.Println("AI ответил:", aiResponse)
}

var documentFields map[string]string = map[string]string{
	"Иск":       "1",
	"Претензия": "2",
}

func GenerateAIResponse(history []Message) (string, string, string, error) {
	var documents []string
	for doc := range documentFields {
		documents = append(documents, doc)
	}

	docNum, err := matchDocument(history, documents)
	if err != nil {
		logrus.Warn("Match document error:", err)
	}

	docTags, docName, err := collectingTags(docNum)
	if err != nil {
		logrus.Warn("Collecting tags error:", err)
	}

	provider := getProvider()

	cont := fmt.Sprintf("You are collecting information for the following fields: %s. Think through everything step by step before asking questions. Ask the user some related questions to fill in the missing information. No more than 4 questions at a time. If you consider some user formulations unsuitable for a legal document, then when forming JSON, change them to more suitable ones from a legal point of view. When asking the user questions, do not send them fields in JSON format. DON'T GUESS and DON't ASSUME the meaning. Answer briefly and concisely. When all the fields are collected, respond ONLY using the JSON object containing the fields. Do not add any other text. If any field remains empty, fill it in as \"нет\".", docTags)
	logrus.Info(cont)

	systemPrompt := Message{
		Role:    "system",
		Content: cont,
	}

	conversation := []Message{systemPrompt}
	conversation = append(conversation, history...)

	req := Request{
		Model:    provider.GetModel(),
		Messages: conversation,
		Stream:   false,
	}

	jsonReq, err := provider.PrepareRequest(req)
	if err != nil {
		return "", "", "", err
	}

	httpReq, err := http.NewRequest("POST", provider.GetEndpoint(), strings.NewReader(string(jsonReq)))
	if err != nil {
		return "", "", "", err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	if orProvider, ok := provider.(OpenRouterProvider); ok {
		httpReq.Header.Set("Authorization", "Bearer "+orProvider.APIKey)
	}

	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		fmt.Println("Error calling provider:", err)
		return "", "", "", err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", "", err
	}

	// Try to parse as OpenRouter format
	var openRouterResp struct {
		Choices []struct {
			Message Message `json:"message"`
		} `json:"choices"`
	}

	var content string
	if err := json.Unmarshal(bodyBytes, &openRouterResp); err == nil && len(openRouterResp.Choices) > 0 {
		content = openRouterResp.Choices[0].Message.Content
	} else {
		var response Response
		if err := json.Unmarshal(bodyBytes, &response); err != nil {
			return "", "", "", fmt.Errorf("failed to parse response for document %v", err)
		}
		content = response.Message.Content
	}

	// Try to extract and parse JSON
	re := regexp.MustCompile(`(?s)\{.*\}`)
	match := re.FindString(content)
	if match != "" {
		var result map[string]interface{}
		if json.Unmarshal([]byte(match), &result) == nil {
			// Check if all fields are present and non-empty
			allPresent := true
			for _, field := range docTags {
				if val, ok := result[field]; !ok || val == "" {
					allPresent = false
					break
				}
			}
			if allPresent {
				fmt.Println("собранная информация:", match)
				return "", docName, match, nil
			}
		}
	}

	return content, "", "", nil
}

func (a *AIApiChats) DownloadDocument(w http.ResponseWriter, r *http.Request) {
	docID := chi.URLParam(r, "doc")

	// Используем filepath.Join для кроссплатформенной совместимости
	basePath := "C:/Users/salam/QuattroProject/containers/temporary_storage"
	fullPath := filepath.Join(basePath, "done_"+docID+".pdf")

	// Проверяем существует ли файл
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		logrus.Errorf("Файл не найден на диске по пути: %s", fullPath)
		http.Error(w, "Файл не найден на сервере", http.StatusNotFound)
		return
	} else if err != nil {
		logrus.Errorf("Ошибка при проверке файла: %v", err)
		http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
		return
	}

	// Проверяем можно ли прочитать файл
	file, err := os.Open(fullPath)
	if err != nil {
		logrus.Errorf("Ошибка открытия файла: %v", err)
		http.Error(w, "Невозможно открыть файл", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	// Получаем информацию о файле
	fileInfo, err := file.Stat()
	if err != nil {
		logrus.Errorf("Ошибка получения информации о файле: %v", err)
		http.Error(w, "Ошибка получения информации о файле", http.StatusInternalServerError)
		return
	}

	// Проверяем что файл не пустой
	if fileInfo.Size() == 0 {
		logrus.Warnf("Файл пустой: %s", fullPath)
		http.Error(w, "Файл пустой", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", docID+".pdf"))
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", fileInfo.Size()))

	http.ServeFile(w, r, fullPath)
}
