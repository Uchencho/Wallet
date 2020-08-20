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
	ID        uint      `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Fullname  string    `json:"fullname"`
	Gender    string    `json:"gender"`
	Activated bool      `json:"activated"`
	CreatedOn time.Time `json:"created_on"`
	LastLogin time.Time `json:"last_login"`
	Token     string    `json:"token"`
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

		_ = json.NewDecoder(req.Body).Decode(&loginDet)
		if loginDet.Username == "" || loginDet.Password == "" {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, `{"Message":"Username and Password is Required"}`)
			return
		}

		user, err := models.GetUser(Db, loginDet.Username)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, `{"Message":"User does not exist"}`)
			return
		}

		if checkPasswordHash(loginDet.Password, user.Password) {
			accessToken, refreshToken, err := generateAuthTokens(user)
			if err != nil {
				fmt.Println(err)
			}
			b := loginResponse{
				ID:        user.ID,
				Username:  user.Username,
				Email:     user.Email,
				Gender:    user.Gender,
				Activated: user.Activated,
				CreatedOn: user.CreatedOn,
				LastLogin: user.LastLogin,
				Token:     accessToken,
			}
			jsonResp, err := json.Marshal(b)
			if err != nil {
				fmt.Println(err)
			}

			expire := time.Now().Add(time.Hour * 6)
			cookie := http.Cookie{Name: "Refreshtoken", Value: refreshToken, Path: "/",
				Expires: expire, HttpOnly: true} // extra agruement, Secure : true, test this on deployment
			http.SetCookie(w, &cookie)
			fmt.Fprint(w, string(jsonResp))
		} else {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, `{"Message":"Invalid credentials"}`)
		}

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprint(w, `{"Message":"Method not allowed"}`)
	}
}

func TestAuth(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if authorized, err := checkAuth(req); !authorized {
		unAuthorizedResponse(w, err)
		return
	}
	w.WriteHeader(http.StatusBadRequest)
	fmt.Fprint(w, `{"Message":"Auth working properly"}`)
}
