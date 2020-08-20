package app

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Uchencho/wallet/models"
)

type loginInfo struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type loginResponse struct {
	Token   string          `json:"token"`
	UserDet models.Accounts `json:"UserDet"`
}

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
		user.Password, _ = hashPassword(user.Password)

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

func LoginUser(w http.ResponseWriter, req *http.Request) {

	w.Header().Set("Content-Type", "application/json")
	switch req.Method {
	case http.MethodPost:
		var loginDet loginInfo

		err := json.NewDecoder(req.Body).Decode(&loginDet)
		if err != nil {
			// Return a 400 bad request here
			fmt.Println(err)
		}

		user, err := models.GetUser(Db, loginDet.Username)
		if err != nil {
			fmt.Println(err)
			panic(err)
		}

		if checkPasswordHash(loginDet.Password, user.Password) {
			jwtString, err := generateJWT(user)
			if err != nil {
				fmt.Println(err)
			}
			b := loginResponse{
				Token:   jwtString,
				UserDet: user,
			}
			jsonResp, err := json.Marshal(b)
			if err != nil {
				fmt.Println(err)
			}

			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, string(jsonResp))
		}

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprint(w, `{"Message":"Method not allowed"}`)
	}
}
