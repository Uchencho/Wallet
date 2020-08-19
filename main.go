package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/Uchencho/wallet/app"
	"github.com/Uchencho/wallet/models"
)

type healthJSON struct {
	Name   string
	Active bool
}

func HealthCheck(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch req.Method {
	case http.MethodGet:
		w.WriteHeader(http.StatusOK)
		resp := &healthJSON{
			Name:   "REST based wallet api, up and running",
			Active: true,
		}
		jsonResp, _ := json.Marshal(resp)
		fmt.Fprint(w, string(jsonResp))
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprint(w, `{"message" : "Method Not Allowed"}`)
	}

}

func GetServerAddress() string {
	const defaultServerAddress = "127.0.0.1:8000"
	serverAddress, present := os.LookupEnv("SERVER_ADDRESS")
	if present {
		return serverAddress
	}
	return defaultServerAddress
}

func main() {

	defer app.Db.Close()

	models.DropTable(app.Db)
	models.CreateAccountTable(app.Db)

	http.HandleFunc("/healthcheck", HealthCheck)
	http.HandleFunc("/register", app.RegisterUser)
	if err := http.ListenAndServe(GetServerAddress(), nil); err != http.ErrServerClosed {
		fmt.Println(err)
	}
}
