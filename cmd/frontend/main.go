package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
)

func main() {
	log := logrus.New()

	// Определяем порт для фронтенда
	port := "3000"

	// Создаем файловый сервер для обслуживания статических файлов
	fileServer := http.FileServer(http.Dir("./frontend/"))

	// Создаем маршрутизатор
	mux := http.NewServeMux()

	// Обслуживаем статические файлы
	mux.Handle("/", fileServer)

	// Создаем сервер
	server := &http.Server{
		Addr:    fmt.Sprintf(":%s", port),
		Handler: mux,
	}

	// Канал для получения сигналов завершения
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Запускаем сервер в отдельной горутине
	go func() {
		log.Infof("Фронтенд-сервер запущен на http://localhost:%s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Ошибка запуска фронтенд-сервера: %v", err)
		}
	}()

	log.Info("Фронтенд-сервер успешно запущен")

	// Ожидаем сигнала завершения
	<-quit
	log.Info("Получен сигнал завершения работы фронтенд-сервера...")

	// Создаем контекст с таймаутом для корректного завершения
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Пытаемся корректно завершить работу сервера
	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Ошибка при завершении работы фронтенд-сервера: %v", err)
	}

	log.Info("Фронтенд-сервер остановлен")
}
