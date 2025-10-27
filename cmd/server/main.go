package main

import (
	"GGChat/internal/api"
	"GGChat/internal/api/crut"
	"GGChat/internal/config"
	database "GGChat/internal/db"
	"GGChat/internal/service/db"
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Malware3447/configo"
	"github.com/Malware3447/spg"
	"github.com/sirupsen/logrus"
)

func main() {
	cfg, _ := configo.MustLoad[config.Config]()
	log := logrus.New()
	ctx := context.Background()

	poolPg, err := spg.NewClient(ctx, &cfg.DatabasePg)
	if err != nil {
		log.Error(fmt.Errorf("ошибка при запуске Postgres: %s", err))
		panic(err)
	}
	log.Info("Postgres успешно запущен")

	dbPg := database.NewRepositoryPg(poolPg)

	pgService := db.NewDbService(dbPg)

	jwt := api.NewJwt(cfg)

	crutApi := crut.NewCrut(pgService, jwt)
	chat := crut.NewApiChats(pgService)

	router := api.NewApi(crutApi, chat, cfg)

	router.Init()

	fileServer := http.FileServer(http.Dir("./frontend/"))

	mux := http.NewServeMux()

	mux.Handle("/api/v1/", router.GetRouter())

	mux.Handle("/", fileServer)

	port := 8080

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Infof("Сервер запущен на http://localhost:%d", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Ошибка запуска сервера: %v", err)
		}
	}()

	log.Info("Сервер успешно запущен")

	<-quit
	log.Info("Получен сигнал завершения работы сервера...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Ошибка при завершении работы сервера: %v", err)
	}

	log.Info("Сервер остановлен")
}
