package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	connString := "postgres://demo:demo@localhost:5432/ggchat"
	pool, err := pgxpool.New(context.Background(), connString)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	defer pool.Close()

	ctx := context.Background()

	migration, err := os.ReadFile("migrations/init.sql")
	if err != nil {
		log.Fatalf("Failed to read migration file: %v", err)
	}

	queries := strings.Split(string(migration), ";")

	for _, query := range queries {
		query = strings.TrimSpace(query)
		if query == "" {
			continue
		}
		_, err := pool.Exec(ctx, query)
		if err != nil {
			log.Fatalf("Failed to execute query: %s, error: %v", query, err)
		}
	}

	fmt.Println("Миграция успешно выполнена")
}
