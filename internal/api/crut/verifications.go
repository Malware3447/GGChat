package crut

import (
	modelV "GGChat/internal/models/crut/verifications"
	"GGChat/internal/service/db"
	"context"
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"net/http"
)

type ApiVerifications struct {
	repo *db.DbService
}

func NewCrut(repo *db.DbService) *ApiVerifications {
	return &ApiVerifications{
		repo: repo,
	}
}

func (v *ApiVerifications) UsersVerifications(w http.ResponseWriter, r *http.Request) {
	log := logrus.New()
	log.Info("Пришел запрос на верификацию...")
	body := modelV.Request{}
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		log.Warn(fmt.Printf("Неправильное тело запроса: %v", err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	ctx := context.Background()

	id, confirmation, err := v.repo.UsersVerification(ctx, body.Username, body.Password)
	if err != nil {
		log.Error(fmt.Printf("Неправильный запрос в базу данных: %v", err))
		http.Error(w, "Invalid request in database", http.StatusBadRequest)
		return
	}

	if confirmation == true {
		response := modelV.Response{
			Id:           id,
			Confirmation: confirmation,
		}

		w.Header().Set("Content-Type", "application-json")
		w.WriteHeader(http.StatusCreated)
		if err = json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		log.Info("Верификация пройдена.")
	} else {
		log.Warn("Ошибка верификации")
		http.Error(w, "Error verifications.", http.StatusBadRequest)
		return
	}
}

func (v *ApiVerifications) UsersRegistrations(w http.ResponseWriter, r *http.Request) {
	log := logrus.New()
	log.Info("Пришел запрос на регистрацию...")
	body := modelV.Request{}
	err := json.NewDecoder(r.Body).Decode(&body)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	ctx := context.Background()

	confirmation, id, err := v.repo.NewUser(ctx, body.Username, body.Password)
	if err != nil {
		http.Error(w, "Invalid request in database", http.StatusBadRequest)
		return
	}

	response := modelV.Response{
		Id:           id,
		Confirmation: confirmation,
		Massage:      "Пользователь успешно зарегистрирован",
	}

	if confirmation == true {
		w.Header().Set("Content-Type", "application-json")
		w.WriteHeader(http.StatusCreated)
		if err = json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, "Internal server error", http.StatusBadRequest)
			return
		}

		log.Info("Пользователь успешно зарегистрирован")
	} else {
		http.Error(w, "Invalid request database", http.StatusBadRequest)
		return
	}

}
