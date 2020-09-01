package transaction

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Uchencho/wallet/config"
)

type verifyPaystackResponse struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
	Data    struct {
		Amount          int         `json:"amount"`
		Currency        string      `json:"currency"`
		TransactionDate time.Time   `json:"transaction_date"`
		Status          string      `json:"status"`
		Reference       string      `json:"reference"`
		Domain          string      `json:"domain"`
		Metadata        string      `json:"metadata"`
		GatewayResponse string      `json:"gateway_response"`
		Message         interface{} `json:"message"`
		Channel         string      `json:"channel"`
		IPAddress       string      `json:"ip_address"`
		Log             struct {
			TimeSpent      int           `json:"time_spent"`
			Attempts       int           `json:"attempts"`
			Authentication interface{}   `json:"authentication"`
			Errors         int           `json:"errors"`
			Success        bool          `json:"success"`
			Mobile         bool          `json:"mobile"`
			Input          []interface{} `json:"input"`
			Channel        interface{}   `json:"channel"`
			History        []struct {
				Type    string `json:"type"`
				Message string `json:"message"`
				Time    int    `json:"time"`
			} `json:"history"`
		} `json:"log"`
		Fees          interface{} `json:"fees"`
		Authorization struct {
			AuthorizationCode string `json:"authorization_code"`
			CardType          string `json:"card_type"`
			Last4             string `json:"last4"`
			ExpMonth          string `json:"exp_month"`
			ExpYear           string `json:"exp_year"`
			Bin               string `json:"bin"`
			Bank              string `json:"bank"`
			Channel           string `json:"channel"`
			Signature         string `json:"signature"`
			Reusable          bool   `json:"reusable"`
			CountryCode       string `json:"country_code"`
		} `json:"authorization"`
		Customer struct {
			ID           int    `json:"id"`
			CustomerCode string `json:"customer_code"`
			FirstName    string `json:"first_name"`
			LastName     string `json:"last_name"`
			Email        string `json:"email"`
		} `json:"customer"`
		Plan            string `json:"plan"`
		RequestedAmount int    `json:"requested_amount"`
	} `json:"data"`
}

func getPaystacKey() string {
	pKey, present := os.LookupEnv("paystack_key")
	if present {
		return pKey
	}
	return config.Paystack_key
}

type GeneratePayment struct {
	Email  string `json:"email"`
	Amount int    `json:"amount"`
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

type UserBalance struct {
	Current_balance int       `json:"current_balance"`
	Last_update     time.Time `json:"last_update"`
}

type recipient struct {
	Email  string `json:"email"`
	Amount int    `json:"amount"`
}

func HitPaystack(email string, amount int) (r PaystackResponse, err error) {
	p := GeneratePayment{
		Email:  email,
		Amount: amount,
	}

	const paylink = "https://api.paystack.co/transaction/initialize"

	reqBody, _ := json.Marshal(p)
	req, err := http.NewRequest("POST", paylink, bytes.NewBuffer(reqBody))
	if err != nil {
		log.Println(err)
		return PaystackResponse{}, err
	}
	value := "Bearer " + getPaystacKey()
	req.Header.Set("Content-Type", "application/json")
	req.Header.Add("Authorization", value)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Error making a request to Paystack ", err)
		return PaystackResponse{}, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("Error making a request to Paystack")
		return PaystackResponse{}, err
	}

	err = json.Unmarshal(body, &r)
	if err != nil {
		log.Println("Error making a request to Paystack")
		return PaystackResponse{}, err
	}
	return r, nil
}

func PaystackVerify(reference string) (res Transactions, attempted bool) {

	var verifyLink = "https://api.paystack.co/transaction/verify/" + reference

	req, err := http.NewRequest("GET", verifyLink, bytes.NewBuffer([]byte{}))
	if err != nil {
		log.Println(err)
		return Transactions{}, false
	}
	value := "Bearer " + getPaystacKey()
	req.Header.Set("Content-Type", "application/json")
	req.Header.Add("Authorization", value)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Error making request to paystack, ", err)
		return Transactions{}, false
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("Error reading the body, ", err)
		return Transactions{}, false
	}

	var verifyPayload verifyPaystackResponse
	err = json.Unmarshal(body, &verifyPayload)
	if err != nil {
		log.Println("Error Unmarshalling json, ", err)
		return Transactions{}, false
	}

	res.Reference = reference
	res.Payment_channel = verifyPayload.Data.Channel
	res.Amount = verifyPayload.Data.Amount
	if verifyPayload.Data.Status == "failed" || verifyPayload.Data.Status == "success" {
		res.Verify_status = true
		if verifyPayload.Data.Status == "failed" {
			res.Payment_status = false
		} else {
			res.Payment_status = true
		}
		return res, true
	}
	return res, false
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
		log.Println("Error in adding transaction to database", err)
		return false
	}
	return true
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
func UpdateTransaction(db *sql.DB, tnx Transactions) (bool, bool) {

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

		_, err := dbTx.ExecContext(ctx, increaseBalance, tnx.Amount/100, time.Now(), tnx.Email)
		if err != nil {
			log.Println(err)
			_ = dbTx.Rollback()
			return false, false
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

func sendMoney(db *sql.DB, r recipient, email string) (bool, bool) {

	ctx := context.Background()
	dbTx, err := db.BeginTx(ctx, nil)
	if err != nil {
		log.Println(err)
		return false, false
	}

	// Check sender's current balance
	var senderBal int
	query := `SELECT current_balance FROM balance WHERE email = $1;`
	row := dbTx.QueryRowContext(ctx, query, email)
	switch err := row.Scan(&senderBal); err {
	case sql.ErrNoRows:
		log.Println("User does not have a balance, very big issue ", err)
		_ = dbTx.Rollback()
		return false, false
	case nil:
		if senderBal < r.Amount {
			return false, true
		}
	default:
		log.Println(err)
		return false, false
	}

	// Check if recipient has a balance record
	row = dbTx.QueryRowContext(ctx, query, r.Email)
	switch err := row.Scan(&senderBal); err {
	case sql.ErrNoRows:
		log.Println("Recipient does not have a balance, very big issue ", err)
		_ = dbTx.Rollback()
		return false, false
	}

	// Debit the account (sender)
	query = `UPDATE balance SET 
				current_balance = current_balance - $1,
				last_update = $2
			WHERE email = $3;`
	_, err = dbTx.ExecContext(ctx, query, r.Amount, time.Now(), email)
	if err != nil {
		log.Println(err)
		_ = dbTx.Rollback()
		return false, false
	}

	// Credit the account (recipient)
	query = `UPDATE balance SET 
				current_balance = current_balance + $1,
				last_update = $2
			WHERE email = $3;`
	_, err = dbTx.ExecContext(ctx, query, r.Amount, time.Now(), r.Email)
	if err != nil {
		log.Println(err)
		_ = dbTx.Rollback()
		return false, false
	}

	// Add a transaction record (sender)
	query = `INSERT INTO transactions (
		email, amount, payment_status, access_code, authorization_url, 
		reference, payment_channel, transaction_date, verify_status
	) VALUES ( 
		$1, $2, $3, $4, $5, $6, $7, $8, $9
	) RETURNING id`

	_, err = dbTx.ExecContext(ctx, query, email, -r.Amount, true, "wallet",
		"wallet", "wallet-to-wallet", "walletPay", time.Now(), true)
	if err != nil {
		log.Println(err)
		_ = dbTx.Rollback()
		return false, false
	}

	// Add a transaction record (recepient)
	query = `INSERT INTO transactions (
		email, amount, payment_status, access_code, authorization_url, 
		reference, payment_channel, transaction_date, verify_status
	) VALUES ( 
		$1, $2, $3, $4, $5, $6, $7, $8, $9
	) RETURNING id`

	_, err = dbTx.ExecContext(ctx, query, r.Email, r.Amount, true, "wallet",
		"wallet", "wallet-to-wallet", "walletPay", time.Now(), true)
	if err != nil {
		log.Println(err)
		_ = dbTx.Rollback()
		return false, false
	}

	// Commit the transactions
	err = dbTx.Commit()
	if err != nil {
		log.Println(err)
		return false, false
	}

	return true, false
}

func deductBalance(db *sql.DB, amount int, email string) (bool, bool) {

	ctx := context.Background()
	dbTx, err := db.BeginTx(ctx, nil)
	if err != nil {
		log.Println(err)
		return false, false
	}

	// Check sender's current balance
	var senderBal int
	query := `SELECT current_balance FROM balance WHERE email = $1;`
	row := dbTx.QueryRowContext(ctx, query, email)
	switch err := row.Scan(&senderBal); err {
	case sql.ErrNoRows:
		log.Println("User does not have a balance, very big issue ", err)
		_ = dbTx.Rollback()
		return false, false
	case nil:
		if senderBal < amount {
			return false, true
		}
	default:
		log.Println(err)
		return false, false
	}

	// Deduct the user's wallet
	query = `UPDATE balance SET
				current_balance = current_balance - $1,
				last_update = $2
				WHERE email = $3;`
	_, err = dbTx.ExecContext(ctx, query, amount, time.Now(), email)
	if err != nil {
		log.Println(err)
		_ = dbTx.Rollback()
		return false, false
	}

	// Add a transaction record
	query = `INSERT INTO transactions (
		email, amount, payment_status, access_code, authorization_url, 
		reference, payment_channel, transaction_date, verify_status
	) VALUES ( 
		$1, $2, $3, $4, $5, $6, $7, $8, $9
	) RETURNING id`

	_, err = dbTx.ExecContext(ctx, query, email, -amount, true, "wallet",
		"payment", "payment", "walletPay", time.Now(), true)
	if err != nil {
		log.Println(err)
		_ = dbTx.Rollback()
		return false, false
	}

	// Commit the transactions
	err = dbTx.Commit()
	if err != nil {
		log.Println(err)
		return false, false
	}

	return true, false
}
