package config

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

func ConnectDatabase() *sql.DB {
	postgres_conn := fmt.Sprintf("host=%s port=%d user=%s "+"password=%s dbname=%s sslmode=disable",
		POSTGRES_HOST, POSTGRES_PORT, POSTGRES_USER, POSTGRES_PASSWORD,
		DB_NAME)

	db, err := sql.Open("postgres", postgres_conn) //os.get(db url) for production
	if err != nil {
		log.Println(err)
		panic("Failed to connect to database")
	}

	bdErr := db.Ping()
	if bdErr != nil {
		panic(bdErr)
	}

	fmt.Println("Connected") // Remember to remove this in production

	return db
}

var Db = ConnectDatabase()
