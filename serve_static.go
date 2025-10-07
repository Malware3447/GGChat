package main

import (
	"log"
	"net/http"
)

func main() {
	// Создаем файловый сервер для статических файлов
	fs := http.FileServer(http.Dir("."))
	
	// Настраиваем маршруты
	http.Handle("/", fs)
	
	log.Println("Сервер статических файлов запущен на http://localhost:3000")
	log.Println("Откройте http://localhost:3000/test_browser.html для тестирования")
	
	// Запускаем сервер
	if err := http.ListenAndServe(":3000", nil); err != nil {
		log.Fatal("Ошибка запуска сервера:", err)
	}
}
