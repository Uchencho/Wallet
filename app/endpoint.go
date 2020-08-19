package app

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Uchencho/wallet/models"
)

var Db = models.ConnectDatabase()

func RegisterUser(w http.ResponseWriter, req *http.Request) {

	w.Header().Set("Content-Type", "application/json")
	switch req.Method {
	case http.MethodPost:
		var user models.Accounts

		err := json.NewDecoder(req.Body).Decode(&user)
		if err != nil {
			fmt.Println(err)
			panic(err)
		}
		user.CreatedOn = time.Now()

		if created := models.AddRecordToAccounts(Db, user); created {
			w.WriteHeader(http.StatusCreated)
			fmt.Fprint(w, `{"Message" : "Successfully Created"}`)
		} else {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, `{"Message" : "User already exists, please login"}`)
		}

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprint(w, `{"Message" : "Method not allowed"}`)
	}

}
