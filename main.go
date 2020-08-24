package main

import (
	"fmt"
	"net/http"

	"github.com/Uchencho/wallet/app"
	"github.com/Uchencho/wallet/config"
)

func main() {

	defer app.Db.Close()

	config.CreateAccountTable(app.Db)

	http.HandleFunc("/healthcheck", app.HealthCheck)
	http.HandleFunc("/register", app.RegisterUser)
	http.HandleFunc("/login", app.LoginUser)
	http.HandleFunc("/profile", app.UserProfile)
	http.HandleFunc("/refresh", app.RefreshToken)
	http.HandleFunc("/logout", app.Logout)
	if err := http.ListenAndServe(app.GetServerAddress(), nil); err != http.ErrServerClosed {
		fmt.Println(err)
	}
}
