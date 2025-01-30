package db

import (
	"context"
	"log"

	"github.com/jackc/pgx/v4/pgxpool"
)

var DB *pgxpool.Pool

func ConnectDB(dsn string) {
	var err error
	DB, err = pgxpool.Connect(context.Background(), dsn)
	if err != nil {
		log.Fatalf("Ошибка подключения к базе данных: %v", err)
	}
	log.Println("✅ Подключение к базе данных успешно")
}
