package app

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/Uchencho/wallet/config"
)

func FundAccount(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	authorized, email, err := checkAuth(req)
	if !authorized {
		unAuthorizedResponse(w, err)
		return
	}

	switch req.Method {
	case http.MethodPost:

		var wg sync.WaitGroup
		wg.Add(1)
		// Get email and amount from post request, make sure amount is string
		var pl config.GeneratePayment
		_ = json.NewDecoder(req.Body).Decode(&pl)
		if pl.Amount == 0 {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, `{"Message" : "Amount is needed"}`)
			return
		}
		pl.Email = fmt.Sprint(email)
		// Hit paystack to return link
		result, err := hitPaystack(pl.Email, pl.Amount*100)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, `{"Message" : "Something went wrong"}`)
		}

		// Send details to db
		go func() {
			if !config.AddTransaction(Db, pl, result) {
				log.Println("Could not add transaction to database")
			}
			wg.Done()
		}()

		// Send the link to the user
		jsonresp, _ := json.Marshal(result.Data)
		fmt.Fprint(w, string(jsonresp))
		wg.Wait()
		return

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprint(w, `{"Message" : "Method not allowed"}`)
	}
}

func TransactionHistory(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	authorized, email, err := checkAuth(req)
	if !authorized {
		unAuthorizedResponse(w, err)
		return
	}

	switch req.Method {
	case http.MethodGet:
		transactions := config.GetTransactions(Db, fmt.Sprint(email))
		if transactions == nil {
			fmt.Fprint(w, `[]`)
			return
		}
		jsonresp, _ := json.Marshal(transactions)
		fmt.Fprint(w, string(jsonresp))
		return
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprint(w, `{"Message" : "Method not allowed"}`)
	}
}

func VerifyTransaction(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	authorized, email, err := checkAuth(req)
	if !authorized {
		unAuthorizedResponse(w, err)
		return
	}

	switch req.Method {
	case http.MethodPost:
		var tranx config.Transactions

		_ = json.NewDecoder(req.Body).Decode(&tranx)
		if tranx.Reference == "" {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, `{"Message" : "Incomplete Payload"}`)
			return
		}

		result, attempted := paystackVerify(tranx.Reference)
		result.Email = fmt.Sprint(email)

		if !attempted {
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, `{"Message" : "Transaction was abandoned"}`)
		}

		// Write to db, update transaction table and balance table
		_, alreadyverified := config.UpdateTransaction(Db, result, true) // Only credits
		if alreadyverified {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, `{"Message" : "Transaction has already been verified"}`)
			return
		}

		w.WriteHeader(http.StatusOK)
		switch {
		case result.Payment_status:
			fmt.Fprint(w, `{"Message" : "Transaction was successful and has been updated accordingly"}`)
		default:
			fmt.Fprint(w, `{"Message" : "Transaction failed"}`)
		}
		return

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprint(w, `{"Message" : "Method not allowed"}`)
	}
}
