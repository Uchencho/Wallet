package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/Uchencho/wallet/models"
)

type loginInfo struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type accessToken struct {
	Access string `json:"access"`
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

		_ = json.NewDecoder(req.Body).Decode(&user)
		if user.Username == "" || user.Password == "" || user.Email == "" {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, `{"Message":"Username, Email and Password is Required"}`)
			return
		}

		user.CreatedOn = time.Now()
		user.LastLogin = time.Now()
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
				Fullname:  user.Fullname,
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

func UserProfile(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	authorized, username, err := checkAuth(req)
	if !authorized {
		unAuthorizedResponse(w, err)
		return
	}

	switch req.Method {
	case http.MethodGet:

		user, _ := models.GetUser(Db, fmt.Sprint(username))
		b := loginResponse{
			ID:        user.ID,
			Username:  user.Username,
			Email:     user.Email,
			Fullname:  user.Fullname,
			Gender:    user.Gender,
			Activated: user.Activated,
			CreatedOn: user.CreatedOn,
			LastLogin: user.LastLogin,
		}

		jsonResp, err := json.Marshal(b)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Fprint(w, string(jsonResp))

	case http.MethodPut:
		var user models.Accounts

		_ = json.NewDecoder(req.Body).Decode(&user)
		if user.Fullname == "" && user.Gender == "" {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, `{"Message":"Fullname or Gender is Required"}`)
			return
		}

		user.Username = fmt.Sprint(username)
		err := models.EditUser(Db, &user)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, `{"Message" : "Something went Wrong"}`)
			return
		}
		w.WriteHeader(http.StatusAccepted)
		fmt.Fprint(w, `{"Message" : "Successfully Edited"}`)

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprint(w, `{"Message" : "Method not allowed"}`)
	}
}

func RefreshToken(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	switch req.Method {
	case http.MethodPost:
		token, err := req.Cookie("Refreshtoken")
		if err != nil {
			unAuthorizedResponse(w, errors.New(`{"Message" : "Credentials Not Sent"}`))
			return
		}
		if authorized, username, _ := checkRefreshToken(token.Value); authorized {
			accessString, err := newAccessToken(fmt.Sprint(username))
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				fmt.Fprint(w, `{"Message" : "Could not generate accesstoken"}`)
				return
			}
			message := accessToken{accessString}
			jsonResp, err := json.Marshal(message)
			if err != nil {
				fmt.Println(err)
			}
			cookie := http.Cookie{Name: "Refreshtoken", Value: token.Value, Path: "/",
				HttpOnly: true} // extra agruement, Secure : true, test this on deployment
			http.SetCookie(w, &cookie)
			w.WriteHeader(http.StatusOK)
			fmt.Fprint(w, string(jsonResp))
			return
		}

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprint(w, `{"Message" : "Method not allowed"}`)
	}

}

func Logout(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	authorized, _, err := checkAuth(req)
	if !authorized {
		unAuthorizedResponse(w, err)
		return
	}
	switch req.Method {
	case http.MethodPost:
		cookie := http.Cookie{Name: "Refreshtoken", Value: "", Path: "/",
			MaxAge: -1, HttpOnly: true}
		http.SetCookie(w, &cookie)
		w.WriteHeader(http.StatusNoContent)
		fmt.Fprint(w, `{"Message" : "Goodbye!"}`)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		fmt.Fprint(w, `{"Message" : "Method not allowed"}`)
	}

}
