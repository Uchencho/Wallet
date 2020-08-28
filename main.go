package main

import (
	"fmt"
	"net/http"

	"github.com/Uchencho/wallet/app/account"
	"github.com/Uchencho/wallet/app/transaction"
	"github.com/Uchencho/wallet/app/utils"
	"github.com/Uchencho/wallet/config"
)

func main() {

	defer config.Db.Close()

	config.CreateAccountTable(config.Db)
	config.CreateTransactionTable(config.Db)
	config.CreateBalanceTable(config.Db)

	http.HandleFunc("/healthcheck", account.HealthCheck)
	http.HandleFunc("/register", account.RegisterUser)
	http.HandleFunc("/login", account.LoginUser)
	http.HandleFunc("/profile", account.UserProfile)
	http.HandleFunc("/refresh", account.RefreshToken)
	http.HandleFunc("/logout", account.Logout)
	http.HandleFunc("/fund", transaction.FundAccount)
	http.HandleFunc("/transactions", transaction.TransactionHistory)
	http.HandleFunc("/verify", transaction.VerifyTransaction)
	http.HandleFunc("/current_balance", transaction.GetBalance)
	if err := http.ListenAndServe(utils.GetServerAddress(), nil); err != http.ErrServerClosed {
		fmt.Println(err)
	}
}
