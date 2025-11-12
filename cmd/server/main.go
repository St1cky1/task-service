package main

import (
	"context"
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5"
)

func main() {
	dbURL := "postgresql://user:pass@localhost:54321/tasks?sslmode=disable"

	runMigrations(dbURL)
	testConnect(dbURL)
	fmt.Println("таблицы созданы")
}

// тестовое подключение
func runMigrations(dbURL string) {
	m, err := migrate.New("file:///Users/v.petrov/task-service/migrations", dbURL)
	if err != nil {
		log.Fatal("ошибка создания миграторов", err)
	}
	defer m.Close()

	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		log.Fatal("ошибка выполнения миграций", err)
	}
	fmt.Println("все работает")
}

func testConnect(dbURL string) {
	ctx := context.Background()

	conn, err := pgx.Connect(ctx, dbURL)
	if err != nil {
		panic(err)
	}
	defer conn.Close(ctx)

	var tableCount int
	err = conn.QueryRow(context.Background(), `
		SELECT COUNT(*) 
		FROM information_schema.tables 
		WHERE table_schema = 'public' 
		AND table_name IN ('user', 'task', 'task_audit')
	`).Scan(&tableCount)

	if err != nil {
		log.Fatal("ошибка проверки таблиц", err)
	}

	fmt.Println("найдено %d таблиц", tableCount)
}
