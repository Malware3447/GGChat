package crut

import (
	"GGChat/internal/models/chats"
	"GGChat/internal/service/db"
	MyWS "GGChat/internal/websocket"
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/sirupsen/logrus"
)

type AIChats struct {
	Repo             *db.DbService
	WebsocketManager *MyWS.Manager
}

func NewAIChats(repo *db.DbService, wsManager *MyWS.Manager) *AIChats {
	return &AIChats{
		Repo:             repo,
		WebsocketManager: wsManager,
	}
}

func (a *AIChats) CreateChatAI(w http.ResponseWriter, r *http.Request) {
	UserID, ok := r.Context().Value("user_id").(int)
	if !ok {
		logrus.Warn("Ошибка получения ID пользователя из контекста")
		http.Error(w, "Error get chats", http.StatusBadRequest)
		return
	}

	body := chats.RequestChatAI{}
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		logrus.Warn("Неверное тело запроса.")
		http.Error(w, "Invflid body request", http.StatusBadRequest)
		return
	}

	response, err := a.Repo.CreateChatAI(context.Background(), UserID, body.Title)
	if err != nil {
		logrus.Warn("ошибка создания чата: ", err)
		http.Error(w, "Error create chat", http.StatusBadRequest)
	}

	w.Header().Set("Content-Type", "application-json")
	w.WriteHeader(http.StatusOK)
	if err = json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func (a *AIChats) GetAllChatsAI(w http.ResponseWriter, r *http.Request) {
	UserID, ok := r.Context().Value("user_id").(int)
	if !ok {
		logrus.Warn("Ошибка получения ID пользователя из контекста")
		http.Error(w, "Error get chats", http.StatusBadRequest)
		return
	}

	response, err := a.Repo.GetAllChatsAI(context.Background(), UserID)
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

func (a *AIChats) DeleteChatAI(w http.ResponseWriter, r *http.Request) {
	IDStr := chi.URLParam(r, "uuid")
	ID, err := strconv.Atoi(IDStr)
	if err != nil {
		logrus.Warn("Ошибка конвертации ID: ", err)
		http.Error(w, "Error converting ID", http.StatusBadRequest)
		return
	}

	err = a.Repo.DeleteChatAI(context.Background(), ID)
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
