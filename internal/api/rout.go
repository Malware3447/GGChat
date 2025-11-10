package api

import (
	"GGChat/internal/api/crut"
	"GGChat/internal/config"
	"bufio"
	"fmt"
	"net"
	"net/http"

	MyMDL "GGChat/internal/api/middleware"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/sirupsen/logrus"
)

type Api struct {
	router     *chi.Mux
	apiService *crut.ApiVerifications
	apiChat    *crut.ApiChats
	cfg        *config.Config
}

type responseWrapper struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWrapper) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWrapper) Flush() {
	if fl, ok := rw.ResponseWriter.(http.Flusher); ok {
		fl.Flush()
	}
}

func (rw *responseWrapper) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hj, ok := rw.ResponseWriter.(http.Hijacker); ok {
		return hj.Hijack()
	}
	return nil, nil, fmt.Errorf("responseWrapper: ResponseWriter не реализует http.Hijacker")
}

func NewApi(apiService *crut.ApiVerifications, apiChat *crut.ApiChats, cfg *config.Config) *Api {
	return &Api{
		router:     nil,
		apiService: apiService,
		apiChat:    apiChat,
		cfg:        cfg,
	}
}

func (a *Api) Init() {

	a.router = chi.NewRouter()

	a.router.Use(middleware.Logger)
	a.router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clientOrigin := r.Header.Get("Origin")
			w.Header().Set("Access-Control-Allow-Origin", clientOrigin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
			w.Header().Set("Access-Control-Allow-Credentials", "true")

			logrus.Info("Request URL: ", r.URL.Path)
			logrus.Info("Request Origin: ", clientOrigin)
			logrus.Info("Request Headers: ", r.Header)

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			wrapped := &responseWrapper{ResponseWriter: w, statusCode: http.StatusOK}
			next.ServeHTTP(wrapped, r)

			logrus.Info("Response Status: ", wrapped.statusCode)
			logrus.Info("Response Headers: ", wrapped.Header())
		})
	})

	a.router.Route("/api/v1/users", func(router chi.Router) {
		router.Post("/verifications", a.apiService.UsersVerifications)
		router.Post("/register", a.apiService.UsersRegistrations)

		router.Group(func(r chi.Router) {
			r.Use(MyMDL.JWTMiddleware(a.cfg.Jwt.SecretToken))

			r.Post("/public_key", a.apiChat.SetPublicKey)
		})
	})

	a.router.Route("/api/v1/chats", func(router chi.Router) {
		router.Use(MyMDL.JWTMiddleware(a.cfg.Jwt.SecretToken))

		router.Post("/new_chat", a.apiChat.NewChat)
		router.Delete("/delete_chat/{uuid}", a.apiChat.DeleteChat)
		router.Get("/all_chats", a.apiChat.GetAllChats)
		router.Get("/get_message/{chat_id}", a.apiChat.GetMessage)
		router.Get("/ws/{chat_id}", a.apiChat.HandleWebSocket)

		router.Get("/public_keys/{chat_id}", a.apiChat.GetChatPublicKeys)
	})
}

func (a *Api) GetRouter() http.Handler {
	return a.router
}
