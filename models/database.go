package models

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

const (
	POSTGRES_USER     = "golang"
	POSTGRES_PASSWORD = "googleGo"
	DB_NAME           = "wallet"
	POSTGRES_HOST     = "localhost"
	POSTGRES_PORT     = 5432
)

func ConnectDatabase() *sql.DB {
	postgres_conn := fmt.Sprintf("host=%s port=%d user=%s "+"password=%s dbname=%s sslmode=disable",
		POSTGRES_HOST, POSTGRES_PORT, POSTGRES_USER, POSTGRES_PASSWORD,
		DB_NAME)

	db, err := sql.Open("postgres", postgres_conn)
	if err != nil {
		fmt.Println(err)
		panic("Failed to connect to database")
	}
	fmt.Println("Connected")

	return db
}
