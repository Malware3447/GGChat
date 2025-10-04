package main

import (
	"GGChat/internal/api"
	"GGChat/internal/api/crut"
	"GGChat/internal/config"
	database "GGChat/internal/db"
	"GGChat/internal/service/db"
	"context"
	"fmt"
	"github.com/Malware3447/configo"
	"github.com/Malware3447/spg"
	"github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"syscall"
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

	crutApi := crut.NewCrut(pgService)

	router := api.NewApi(crutApi)

	router.Init()

	log.Info("Сервис успешно запущен")

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-ctx.Done():
	case <-quit:
		log.Info(ctx, "Завершение работы сервиса")
	}
}
