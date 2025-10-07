package api

import (
	"GGChat/internal/api/crut"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Api struct {
	router     *chi.Mux
	apiService *crut.ApiVerifications
	apiChat    *crut.ApiChats
}

func NewApi(apiService *crut.ApiVerifications, apiChat *crut.ApiChats) *Api {
	return &Api{
		router:     nil,
		apiService: apiService,
		apiChat:    apiChat,
	}
}

func (a *Api) Init() {
	a.router = chi.NewRouter()

	// CORS middleware
	a.router.Use(middleware.Logger)
	a.router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	})

	a.router.Route("/api/v1/users", func(router chi.Router) {
		router.Post("/verifications", a.apiService.UsersVerifications)
		router.Post("/register", a.apiService.UsersRegistrations)
	})

	a.router.Route("/api/v1/chats", func(router chi.Router) {
		router.Post("/new_chat", a.apiChat.NewChat)
	})

	go func() {
		if err := http.ListenAndServe(fmt.Sprintf(":%v", 8080), a.router); err != nil {
			panic(fmt.Sprintf("%v: %v", nil, err))
		}
	}()
}
