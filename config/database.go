package config

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

func databaseURL() string {

	dBUrl, present := os.LookupEnv("DATABASE_URL")
	if present {
		return dBUrl
	}
	postgres_conn := fmt.Sprintf("host=%s port=%d user=%s "+"password=%s dbname=%s sslmode=disable",
		POSTGRES_HOST, POSTGRES_PORT, POSTGRES_USER, POSTGRES_PASSWORD,
		DB_NAME)
	return postgres_conn
}

func ConnectDatabase() *sql.DB {

	db, err := sql.Open("postgres", databaseURL())
	if err != nil {
		log.Println(err)
		panic("Failed to connect to database")
	}

	bdErr := db.Ping()
	if bdErr != nil {
		panic(bdErr)
	}
	return db
}

var Db = ConnectDatabase()
