package config

import (
	"context"
	"database/sql"
	"fmt"
	"log"
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

type Transactions struct {
	ID                uint      `json:"id"`
	Email             string    `json:"email"`
	Amount            int       `json:"amount"`
	Payment_status    bool      `json:"payment_status"`
	Access_code       string    `json:"access_code"`
	Authorization_url string    `json:"authorization_url"`
	Reference         string    `json:"reference"`
	Payment_channel   string    `json:"payment_channel"`
	Transaction_date  time.Time `json:"transaction_date"`
	Verify_status     bool      `json:"verify_status"`
}

type GeneratePayment struct {
	Email  string `json:"email"`
	Amount int    `json:"amount"`
}

type UserBalance struct {
	Current_balance int       `json:"current_balance"`
	Last_update     time.Time `json:"last_update"`
}

type PaystackResponse struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
	Data    struct {
		AuthorizationURL string `json:"authorization_url"`
		AccessCode       string `json:"access_code"`
		Reference        string `json:"reference"`
	} `json:"data"`
}

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
		fmt.Println(err)
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
		fmt.Println(err)
		return false
	}
}

func GetUser(db *sql.DB, email string) (Accounts, error) {
	query := `SELECT id, username, email, password, fullname, gender, last_login, created_on
			  FROM accounts WHERE email = $1;`

	var (
		user Accounts
		g    interface{}
		f    interface{}
	)

	row := db.QueryRow(query, email)
	switch err := row.Scan(&user.ID, &user.Username, &user.Email,
		&user.Password, &f, &g, &user.LastLogin, &user.CreatedOn); err {
	case sql.ErrNoRows:
		return Accounts{}, err
	case nil:
		if f == nil {
			user.Fullname = ""
		} else {
			user.Fullname = fmt.Sprint(f)
		}

		if g == nil {
			user.Gender = ""
		} else {
			user.Gender = fmt.Sprint(g)
		}
		return user, nil
	default:
		fmt.Println(err)
		return Accounts{}, err
	}
}

func GetUserLogin(db *sql.DB, username string) (Accounts, error) {

	ctx := context.Background()
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return Accounts{}, err
	}

	updateLoginQuery := `UPDATE accounts SET last_login = $1 WHERE username = $2`

	_, err = tx.ExecContext(ctx, updateLoginQuery, time.Now(), username)
	if err != nil {
		_ = tx.Rollback()
		return Accounts{}, err
	}

	query := `SELECT id, username, email, password, fullname, gender, last_login, created_on
			  FROM accounts WHERE username = $1;`

	var (
		user Accounts
		g    interface{}
		f    interface{}
	)

	row := tx.QueryRowContext(ctx, query, username)
	switch err := row.Scan(&user.ID, &user.Username, &user.Email,
		&user.Password, &f, &g, &user.LastLogin, &user.CreatedOn); err {
	case sql.ErrNoRows:
		_ = tx.Rollback()
		return Accounts{}, err
	case nil:
		err = tx.Commit()
		if err != nil {
			fmt.Println(err)
		}
		if f == nil {
			user.Fullname = ""
		} else {
			user.Fullname = fmt.Sprint(f)
		}

		if g == nil {
			user.Gender = ""
		} else {
			user.Gender = fmt.Sprint(g)
		}
		return user, nil
	default:
		fmt.Println(err)
		return Accounts{}, err
	}
}

func EditUser(db *sql.DB, user *Accounts) error {
	query := `UPDATE accounts SET fullname = $1, gender = $2 WHERE email = $3`

	_, err := db.Exec(query, user.Fullname, user.Gender, user.Email)
	if err != nil {
		return err
	}
	return nil
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

func AddTransaction(db *sql.DB, p GeneratePayment, res PaystackResponse) bool {

	query := `INSERT INTO transactions (
		email, amount, payment_status, access_code, authorization_url, 
		reference, payment_channel, transaction_date, verify_status
	) VALUES ( 
		$1, $2, $3, $4, $5, $6, $7, $8, $9
	) RETURNING id`

	_, err := db.Exec(query, p.Email, p.Amount, false, res.Data.AccessCode,
		res.Data.AuthorizationURL, res.Data.Reference, "", time.Now(), false)

	if err != nil {
		log.Println(err)
		return false
	}
	return true
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

func GetTransactions(db *sql.DB, email string) (tnx []Transactions) {

	query := `SELECT id, email, amount, payment_status,
			transaction_date, reference, payment_channel, authorization_url, verify_status
			FROM transactions WHERE email = $1;`

	row, err := db.Query(query, email)
	if err != nil {
		log.Println(err)
		return []Transactions{}
	}
	defer row.Close()

	var temp Transactions

	for row.Next() {
		err := row.Scan(&temp.ID, &temp.Email, &temp.Amount, &temp.Payment_status,
			&temp.Transaction_date, &temp.Reference, &temp.Payment_channel,
			&temp.Authorization_url, &temp.Verify_status)
		if err != nil {
			log.Println(err)
			return []Transactions{}
		}
		tnx = append(tnx, temp)
	}

	err = row.Err()
	if err != nil {
		log.Println(err)
		return []Transactions{}
	}
	return tnx
}

func InitializeBalance(db *sql.DB, email string) bool {

	query := `INSERT INTO balance (
		email, current_balance, last_update
	) VALUES (
		$1, $2, $3
	) RETURNING id`

	_, err := db.Exec(query, email, 0, time.Now())
	if err != nil {
		log.Println(err)
		return false
	}
	return true
}

// Returns status of query and if transaction has already been verified
func UpdateTransaction(db *sql.DB, tnx Transactions, credit bool) (bool, bool) {

	ctx := context.Background()
	dbTx, err := db.BeginTx(ctx, nil)
	if err != nil {
		log.Println(err)
	}

	var verify_status bool

	// Check status of transaction
	q := `SELECT verify_status FROM transactions WHERE email = $1 AND reference = $2;`
	row := dbTx.QueryRowContext(ctx, q, tnx.Email, tnx.Reference)
	switch err := row.Scan(&verify_status); err {
	case sql.ErrNoRows:
		log.Println(err)
		_ = dbTx.Rollback()
		return false, false
	case nil:
		if verify_status {
			return true, true // Transaction has already been verified
		}
	default:
		log.Println("default error ", err)
	}

	if tnx.Payment_status {
		// Transaction was successful and it has not been verified before, update balance

		increaseBalance := `UPDATE balance SET 
							current_balance = current_balance + $1,
							last_update = $2
					  WHERE email = $3;`

		decreaseBalance := `UPDATE balance SET 
							current_balance = current_balance - $1,
							last_update = $2
						WHERE email = $3;`
		if credit {
			_, err := dbTx.ExecContext(ctx, increaseBalance, tnx.Amount/100, time.Now(), tnx.Email)
			if err != nil {
				log.Println(err)
				_ = dbTx.Rollback()
				return false, false
			}
		} else {
			_, err := dbTx.ExecContext(ctx, decreaseBalance, tnx.Amount/100, time.Now(), tnx.Email)
			if err != nil {
				log.Println(err)
				_ = dbTx.Rollback()
				return false, false
			}
		}
	}

	updatetranx := `UPDATE transactions SET 
						verify_status = true,
						payment_status = $1,
						payment_channel = $2 
					WHERE email = $3;`

	_, err = dbTx.ExecContext(ctx, updatetranx, tnx.Payment_status,
		tnx.Payment_channel, tnx.Email)
	if err != nil {
		log.Println(err)
		_ = dbTx.Rollback()
		return false, false
	}
	err = dbTx.Commit()
	if err != nil {
		log.Println(err)
		return false, false
	}
	return true, false
}

func GetCurrentBalance(db *sql.DB, email string) (bal UserBalance, err error) {

	query := `SELECT current_balance, last_update FROM balance WHERE email = $1;`

	row := db.QueryRow(query, email)
	switch err := row.Scan(&bal.Current_balance, &bal.Last_update); err {
	case sql.ErrNoRows:
		log.Println("This is bad, user has no balance record")
		return UserBalance{}, err
	case nil:
		return bal, nil
	default:
		return UserBalance{}, err
	}

}
