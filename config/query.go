package config

import (
	"database/sql"
	"fmt"
	"log"
)

func CreateAccountTable(db *sql.DB) {

	query := `CREATE TABLE IF NOT EXISTS accounts (
		id serial PRIMARY KEY,
		username VARCHAR ( 50 ) UNIQUE NOT NULL,
		email VARCHAR ( 50 ) UNIQUE NOT NULL,
		password VARCHAR ( 200 ) NOT NULL,
		fullname VARCHAR ( 200 ),
		gender VARCHAR ( 200 ),
		activated BOOL,
		created_on TIMESTAMP,
		last_login TIMESTAMP
		);`

	_, err := db.Exec(query)
	if err != nil {
		fmt.Println(err)
	}
}

// Droptable which is used to drop the accounts table is currently not in use
func DropTable(db *sql.DB) {
	query := `DROP TABLE IF EXISTS accounts`

	_, err := db.Exec(query)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("\n\n Table dropped successfully")
}

func CreateTransactionTable(db *sql.DB) {

	query := `CREATE TABLE IF NOT EXISTS transactions (
		id serial PRIMARY KEY,
		email VARCHAR ( 50 ) NOT NULL,
		amount INT NOT NULL,
		payment_status BOOL,
		access_code VARCHAR ( 200 ),
		authorization_url VARCHAR ( 200 ),
		reference VARCHAR ( 200 ),
		payment_channel VARCHAR ( 200 ),
		transaction_date TIMESTAMP,
		verify_status BOOL,
		FOREIGN KEY(email)
			REFERENCES accounts(email)
			ON DELETE CASCADE
		);`

	_, err := db.Exec(query)
	if err != nil {
		log.Println(err)
	}
}

func CreateBalanceTable(db *sql.DB) {

	query := `CREATE TABLE IF NOT EXISTS balance (
		id serial PRIMARY KEY,
		email VARCHAR ( 50 ) UNIQUE NOT NULL,
		current_balance INT NOT NULL,
		last_update TIMESTAMP,
		FOREIGN KEY(email)
			REFERENCES accounts(email)
			ON DELETE CASCADE
	);`

	_, err := db.Exec(query)
	if err != nil {
		log.Println(err)
	}
}
