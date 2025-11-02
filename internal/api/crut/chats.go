package crut

import (
	"GGChat/internal/models/chats"
	"GGChat/internal/service/db"
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type ApiChats struct {
	repo *db.DbService
}

func NewApiChats(repo *db.DbService) *ApiChats {
	return &ApiChats{
		repo: repo,
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

	err = a.repo.NewMessage(ctx, body.ChatId, userId, body.Content)
	if err != nil {
		log.Warn("ошибка добавления сообщения в бд: ", err)
		http.Error(w, "Error adding message in database", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err = json.NewEncoder(w).Encode(nil); err != nil {
		log.Warn("Ошибка сервера: ", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	log.Info("Сообщение добавлено")
}

func (a *ApiChats) GetMessage(w http.ResponseWriter, r *http.Request) {
	log := logrus.New()

	_, ok := r.Context().Value("user_id").(int)
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

	response, err := a.repo.GetMessage(ctx, ChatId)
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
