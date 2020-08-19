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

		respId := models.AddRecordToAccounts(Db, user)
		jsonResp, err := json.Marshal(respId)
		if err != nil {
			fmt.Println(err)
		}
		w.WriteHeader(http.StatusCreated)
		fmt.Fprint(w, string(jsonResp))

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprint(w, `{"Message" : "Method not allowed"}`)
	}

}
