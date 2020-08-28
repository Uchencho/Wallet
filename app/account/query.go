package account

import (
	"context"
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
