package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
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
	http.HandleFunc("/healthcheck", HealthCheck)
	if err := http.ListenAndServe(GetServerAddress(), nil); err != http.ErrServerClosed {
		fmt.Println(err)
	}
}
