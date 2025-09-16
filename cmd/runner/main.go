package main

import (
	"GGChat/internal/config"
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

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-ctx.Done():
	case <-quit:
		log.Info(ctx, "Завершение работы сервиса")
	}
}
