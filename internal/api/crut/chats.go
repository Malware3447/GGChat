package crut

import (
	"GGChat/internal/models/chats"
	"GGChat/internal/service/db"
	MyWS "GGChat/internal/websocket"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type ApiChats struct {
	repo             *db.DbService
	WebsocketManager *MyWS.Manager
}

func NewApiChats(repo *db.DbService, wsManager *MyWS.Manager) *ApiChats {
	return &ApiChats{
		repo:             repo,
		WebsocketManager: wsManager,
	}
}

func (a *ApiChats) NewChat(w http.ResponseWriter, r *http.Request) {
	log := logrus.New()
	log.Info("Поступил запрос на создание нового чата...")

	body := chats.NewChatRequest{}
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		log.Warn("Неверное тело запроса.")
		http.Error(w, "Invflid body request", http.StatusBadRequest)
		return
	}

	userId, ok := r.Context().Value("user_id").(int)
	if !ok {
		log.Warn("Ошибка получения ID пользователя из контекста")
		http.Error(w, "Error get chats", http.StatusBadRequest)
		return
	}

	ctx := context.Background()

	other_user_id, err := a.repo.GetUser(ctx, body.UserName)
	if err != nil {
		log.Warn("Пользователь не найден. Ошибка: ", err)
		http.Error(w, "Error creat new chat", http.StatusBadRequest)
		return
	}

	confirmation, uuid, err := a.repo.NewChat(ctx, body.ChatName, userId, other_user_id)
	if err != nil {
		log.Warn("Ошибка создания нового чата: ", err)
		http.Error(w, "Error creat new chat", http.StatusBadRequest)
		return
	}

	if confirmation == true {
		response := chats.Response{
			ChatName: body.ChatName,
			Uuid:     uuid,
			Status:   true,
		}

		w.Header().Set("Content-Type", "application-json")
		w.WriteHeader(http.StatusCreated)
		if err = json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		log.Info("Новый чат создан")
	} else {
		log.Warn("Не удалось создать новый чат")
		http.Error(w, "couldn't create a new chat", http.StatusBadRequest)
		return
	}
}

func (a *ApiChats) DeleteChat(w http.ResponseWriter, r *http.Request) {
	log := logrus.New()
	log.Info("Пришел запрос на удаление чата...")
	uuidStr := chi.URLParam(r, "uuid")
	uuid, err := uuid.Parse(uuidStr)
	if err != nil {
		log.Warn("Ошибка парсинга uuid: ", err)
		http.Error(w, "Invalid response", http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	err = a.repo.DeleteChat(ctx, uuid)
	if err != nil {
		log.Warn("Ошибка удаления чата: ", err)
		http.Error(w, "Error request in database", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err = json.NewEncoder(w).Encode(nil); err != nil {
		log.Warn("Ошибка сервера: ", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	log.Info("Чат успешно удален")
}

func (a *ApiChats) GetAllChats(w http.ResponseWriter, r *http.Request) {
	log := logrus.New()

	userId, ok := r.Context().Value("user_id").(int)
	if !ok {
		log.Warn("Ошибка получения ID пользователя из контекста")
		http.Error(w, "Error get chats", http.StatusBadRequest)
		return
	}

	log.Info("Запрос списка чатов от пользователя №", userId, "...")

	ctx := context.Background()

	response, err := a.repo.GetAllChats(ctx, userId)
	if err != nil {
		log.Warn("Ошибка в запросе к БД: ", err)
		http.Error(w, "Error request database", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err = json.NewEncoder(w).Encode(response); err != nil {
		log.Warn("Ошибка сервера: ", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	log.Info("Информация о чатах обновлена.")
}

func (a *ApiChats) NewMessage(w http.ResponseWriter, r *http.Request) {
	log := logrus.New()

	userId, ok := r.Context().Value("user_id").(int)
	if !ok {
		log.Warn("Ошибка получения ID пользователя из контекста")
		http.Error(w, "Error get chats", http.StatusBadRequest)
		return
	}

	body := chats.NewMessageRequest{}
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		log.Warn("Неверное тело запроса.")
		http.Error(w, "Invflid body request", http.StatusBadRequest)
		return
	}

	log.Info("Запрос на отправку нового сообщения...")

	ctx := context.Background()

	_, status, err := a.repo.NewMessage(ctx, body.ChatId, userId, body.Content)
	if err != nil {
		log.Warn("ошибка добавления сообщения в бд: ", err)
		http.Error(w, "Error adding message in database", http.StatusBadRequest)
		return
	}

	response := chats.Message{
		Status: status,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err = json.NewEncoder(w).Encode(response); err != nil {
		log.Warn("Ошибка сервера: ", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	log.Info("Сообщение добавлено")
}

func (a *ApiChats) GetMessage(w http.ResponseWriter, r *http.Request) {
	log := logrus.New()

	userId, ok := r.Context().Value("user_id").(int)
	if !ok {
		log.Warn("Ошибка получения ID пользователя из контекста")
		http.Error(w, "Error get chats", http.StatusBadRequest)
		return
	}

	ChatIdStr := chi.URLParam(r, "chat_id")
	ChatId, err := uuid.Parse(ChatIdStr)
	if err != nil {
		log.Warn("Ошибка парсинга UUID: ", err)
		http.Error(w, "Error parsing UUID", http.StatusBadRequest)
		return
	}

	ctx := context.Background()

	response, err := a.repo.GetMessage(ctx, ChatId, userId)
	if err != nil {
		log.Warn("ошибка получения списка сообщений: ", err)
		http.Error(w, "error get list message", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err = json.NewEncoder(w).Encode(response); err != nil {
		log.Warn("Ошибка сервера: ", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func (a *ApiChats) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	logrus.Info("MEOW")
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logrus.Error("ошибка апгрейда до WebSocket: ", err)
		return
	}

	chatIdStr := chi.URLParam(r, "chat_id")

	userId, ok := r.Context().Value("user_id").(int)
	if !ok {
		logrus.Error("ошибка получения Id пользователя из контекста")
		conn.Close()
		return
	}

	client := &MyWS.Client{
		Id:     fmt.Sprintf("%d-%s", userId, chatIdStr),
		UserId: userId,
		ChatId: chatIdStr,
		Conn:   conn,
		Send:   make(chan []byte, 256),
	}

	a.WebsocketManager.Register <- client

	go a.handleClientMessages(client)
	go a.writeMessagesToClient(client)
}

func (a *ApiChats) handleClientMessages(client *MyWS.Client) {
	defer func() {
		a.WebsocketManager.Undergister <- client
		client.Conn.Close()
	}()

	for {
		var msg MyWS.Message
		err := client.Conn.ReadJSON(&msg)

		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure, websocket.CloseNoStatusReceived) {
				logrus.Error("Ошибка чтения/закрытия соединения:", err)
			}
			break
		}

		switch msg.Type {
		case "new_message":
			chatId, err := uuid.Parse(client.ChatId)
			if err != nil {
				logrus.Error("Ошибка парсинга ChatId:", err)
				continue
			}

			messageId, status, err := a.repo.NewMessage(context.Background(), chatId, client.UserId, msg.Content)
			if err != nil {
				logrus.Error("Ошибка сохранения сообщения:", err)
				client.Conn.Close()
				break
			}

			message := MyWS.Message{
				Id:        messageId,
				Type:      "new_message",
				Content:   msg.Content,
				ChatId:    client.ChatId,
				UserId:    client.UserId,
				Status:    status,
				Timestamp: time.Now(),
			}
			a.WebsocketManager.SendMessage(message)

		case "read_receipt":
			err := a.repo.UpdateMessageStatus(context.Background(), msg.MessageId, "read")
			if err != nil {
				logrus.Error("Ошибка обновления статуса на 'read':", err)
				continue
			}

			updateMessage := MyWS.Message{
				Type:   "status_update",
				Id:     msg.MessageId,
				ChatId: client.ChatId,
				Status: "read",
			}
			a.WebsocketManager.SendMessage(updateMessage)

		default:
			logrus.Warnf("Получено сообщение с неизвестным типом: %s", msg.Type)
		}
	}
}

func (a *ApiChats) writeMessagesToClient(client *MyWS.Client) {
	defer client.Conn.Close()

	for message := range client.Send {

		w, err := client.Conn.NextWriter(websocket.TextMessage)
		if err != nil {
			return
		}

		w.Write(message)

		n := len(client.Send)
		for i := 0; i < n; i++ {
			w.Write(<-client.Send)
		}

		if err := w.Close(); err != nil {
			return
		}
	}

	client.Conn.WriteMessage(websocket.CloseMessage, []byte{})
}
