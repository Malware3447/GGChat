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

	ctx := context.Background()

	confirmation, uuid, err := a.repo.NewChat(ctx, body.ChatName)
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
	log.Info("Обновляем информацию о чатах...")

	ctx := context.Background()

	response, err := a.repo.GetAllChats(ctx)
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
