package app

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/Uchencho/wallet/config"
	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
)

type generatePayment struct {
	Email  string `json:"email"`
	Amount string `json:"amount"`
}

type PaystackResponse struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
	Data    struct {
		AuthorizationURL string `json:"authorization_url"`
		AccessCode       string `json:"access_code"`
		Reference        string `json:"reference"`
	} `json:"data"`
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 4)
	return string(bytes), err
}

func checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

var (
	signingKey        = []byte("3d67d77426d9878967a177437316554b0088fa88be95846252011528e8bad788")
	refreshSigningKey = []byte("b178604f6216f904f394641fd167078e426d5fe9ce20d4c07a65e8dd051a40d9")
)

func generateAuthTokens(user config.Accounts) (string, string, error) {

	// Access token
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)

	claims["authorized"] = true
	claims["client"] = user.Username
	claims["exp"] = time.Now().Add(time.Minute * 30).Unix()

	accessString, err := token.SignedString(signingKey)

	if err != nil {
		return "", "", err
	}

	// Refresh token
	refreshToken := jwt.New(jwt.SigningMethodHS256)
	refreshClaims := refreshToken.Claims.(jwt.MapClaims)

	refreshClaims["authorized"] = true
	refreshClaims["client"] = user.Username
	refreshClaims["exp"] = time.Now().Add(time.Hour * 6).Unix()

	refreshString, err := refreshToken.SignedString(refreshSigningKey)

	if err != nil {
		return "", "", err
	}

	return accessString, refreshString, nil
}

func checkAuth(r *http.Request) (bool, interface{}, error) {
	if r.Header["Authorization"] != nil {
		if len(strings.Split(r.Header["Authorization"][0], " ")) < 2 {
			return false, "", errors.New("Invalid Credentials")
		}
		accessToken := strings.Split(r.Header["Authorization"][0], " ")[1]
		token, err := jwt.Parse(accessToken, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("An error occured")
			}
			return signingKey, nil
		})
		if err != nil {
			return false, "", err
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			return true, claims["client"], nil
		}
	}
	return false, "", errors.New("Credentials not provided")
}

func unAuthorizedResponse(w http.ResponseWriter, err error) {
	cookie := http.Cookie{Name: "Refreshtoken", Value: "", Path: "/",
		MaxAge: -1, HttpOnly: true}
	http.SetCookie(w, &cookie)
	w.WriteHeader(http.StatusForbidden)
	fmt.Fprint(w, err.Error())
}

func checkRefreshToken(refreshToken string) (bool, interface{}, error) {

	token, err := jwt.Parse(refreshToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("An error occured")
		}
		return refreshSigningKey, nil
	})
	if err != nil {
		return false, "", err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return true, claims["client"], nil
	}

	return false, "", errors.New("Credentials not provided")
}

func newAccessToken(username string) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)

	claims["authorized"] = true
	claims["client"] = username
	claims["exp"] = time.Now().Add(time.Minute * 30).Unix()

	accessString, err := token.SignedString(signingKey)

	if err != nil {
		return "", err
	}

	return accessString, nil
}

func GetServerAddress() string {
	const defaultServerAddress = "127.0.0.1:8000"
	serverAddress, present := os.LookupEnv("SERVER_ADDRESS")
	if present {
		return serverAddress
	}
	return defaultServerAddress
}

func hitPaystack(email, amount string) (r PaystackResponse, err error) {
	p := generatePayment{
		Email:  email,
		Amount: amount,
	}

	const paylink = "https://api.paystack.co/transaction/initialize"

	reqBody, _ := json.Marshal(p)
	req, err := http.NewRequest("POST", paylink, bytes.NewBuffer(reqBody))
	if err != nil {
		log.Println(err)
	}
	value := "Bearer " + config.Paystack_key
	req.Header.Set("Content-Type", "application/json")
	req.Header.Add("Authorization", value)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Error making a request to Paystack ", err)
		return PaystackResponse{}, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("Error making a request to Paystack")
	}

	err = json.Unmarshal(body, &r)
	if err != nil {
		log.Println("Error making a request to Paystack")
	}
	return r, nil
}
