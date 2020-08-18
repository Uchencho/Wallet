package models

import (
	"database/sql"
	"fmt"
	"time"
)

type User struct {
	ID        uint      `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Password  string    `json:"password"`
	Fullname  string    `json:"fullname"`
	Gender    string    `json:"gender"`
	Activated bool      `json:"activated"`
	CreatedOn time.Time `json:"created_on"`
	LastLogin time.Time `json:"last_login"`
}

// query = "DROP TABLE IF EXISTS test; CREATE TABLE test(name VARCHAR(50), designation VARCHAR(100), age VARCHAR(50));"

func CreateTable(db *sql.DB) {
	query := `CREATE TABLE IF NOT EXISTS accounts (
		id serial PRIMARY KEY,
		username VARCHAR ( 50 ) UNIQUE NOT NULL,
		email VARCHAR ( 50 ) UNIQUE NOT NULL,
		password VARCHAR ( 50 ) NOT NULL,
		fullname VARCHAR ( 200 ),
		gender VARCHAR ( 200 ),
		activated VARCHAR ( 10 ),
		created_on TIMESTAMP,
		last_login TIMESTAMP
		);`
	_, err := db.Exec(query)
	if err != nil {
		panic(err)
	}

	fmt.Println("\n\n Created the table successfully")
}
