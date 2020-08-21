package models

import (
	"database/sql"
	"fmt"
	"time"
)

type Accounts struct {
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

// type NullTime struct {
// 	pq.NullTime
// }

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
		panic(err)
	}
}

// Droptable which is used to drop the accounts table is currently not in use
func DropTable(db *sql.DB) {
	query := `DROP TABLE IF EXISTS accounts`

	_, err := db.Exec(query)
	if err != nil {
		panic(err)
	}

	fmt.Println("\n\n Table dropped successfully")
}

func AddRecordToAccounts(db *sql.DB, user Accounts) bool {

	if userExists := checkUser(db, user); userExists {
		return false
	}

	query := `INSERT INTO accounts (
		username, email, password, created_on, last_login
	) VALUES (
		$1, $2, $3, $4, $5
	) RETURNING id`

	_, err := db.Exec(query, user.Username, user.Email, user.Password,
		user.CreatedOn, user.LastLogin)
	if err != nil {
		panic(err)
	}
	return true
}

func checkUser(db *sql.DB, user Accounts) bool {
	query := `SELECT email FROM accounts WHERE email = $1`

	var email string
	row := db.QueryRow(query, user.Email)
	switch err := row.Scan(&email); err {
	case sql.ErrNoRows:
		return false
	case nil:
		return true
	default:
		panic(err)
	}
}

func GetUser(db *sql.DB, username string) (Accounts, error) {
	query := `SELECT id, username, email, password, last_login, created_on
			  FROM accounts WHERE username = $1`

	var user Accounts

	row := db.QueryRow(query, username)
	switch err := row.Scan(&user.ID, &user.Username, &user.Email,
		&user.Password, &user.LastLogin, &user.CreatedOn); err {
	case sql.ErrNoRows:
		return Accounts{}, err
	case nil:
		return user, nil
	default:
		panic(err)
	}
}
