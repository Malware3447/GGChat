package api

import (
	"GGChat/internal/api/crut"
	"fmt"
	"github.com/go-chi/chi/v5"
	"net/http"
)

type Api struct {
	router     *chi.Mux
	apiService *crut.ApiVerifications
}

func NewApi(apiService *crut.ApiVerifications) *Api {
	return &Api{
		router:     nil,
		apiService: apiService,
	}
}

func (a *Api) Init() {
	a.router = chi.NewRouter()

	a.router.Route("/api/v1/users", func(router chi.Router) {
		router.Get("/verifications/", a.apiService.UsersVerifications)
		router.Post("/register", a.apiService.UsersRegistrations)
	})

	go func() {
		if err := http.ListenAndServe(fmt.Sprintf(":%v", 8080), a.router); err != nil {
			panic(fmt.Sprintf("%v: %v", nil, err))
		}
	}()
}
