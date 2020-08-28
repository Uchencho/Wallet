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

type verifyPaystackResponse struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
	Data    struct {
		Amount          int         `json:"amount"`
		Currency        string      `json:"currency"`
		TransactionDate time.Time   `json:"transaction_date"`
		Status          string      `json:"status"`
		Reference       string      `json:"reference"`
		Domain          string      `json:"domain"`
		Metadata        string      `json:"metadata"`
		GatewayResponse string      `json:"gateway_response"`
		Message         interface{} `json:"message"`
		Channel         string      `json:"channel"`
		IPAddress       string      `json:"ip_address"`
		Log             struct {
			TimeSpent      int           `json:"time_spent"`
			Attempts       int           `json:"attempts"`
			Authentication interface{}   `json:"authentication"`
			Errors         int           `json:"errors"`
			Success        bool          `json:"success"`
			Mobile         bool          `json:"mobile"`
			Input          []interface{} `json:"input"`
			Channel        interface{}   `json:"channel"`
			History        []struct {
				Type    string `json:"type"`
				Message string `json:"message"`
				Time    int    `json:"time"`
			} `json:"history"`
		} `json:"log"`
		Fees          interface{} `json:"fees"`
		Authorization struct {
			AuthorizationCode string `json:"authorization_code"`
			CardType          string `json:"card_type"`
			Last4             string `json:"last4"`
			ExpMonth          string `json:"exp_month"`
			ExpYear           string `json:"exp_year"`
			Bin               string `json:"bin"`
			Bank              string `json:"bank"`
			Channel           string `json:"channel"`
			Signature         string `json:"signature"`
			Reusable          bool   `json:"reusable"`
			CountryCode       string `json:"country_code"`
		} `json:"authorization"`
		Customer struct {
			ID           int    `json:"id"`
			CustomerCode string `json:"customer_code"`
			FirstName    string `json:"first_name"`
			LastName     string `json:"last_name"`
			Email        string `json:"email"`
		} `json:"customer"`
		Plan            string `json:"plan"`
		RequestedAmount int    `json:"requested_amount"`
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
	claims["client"] = user.Email
	claims["exp"] = time.Now().Add(time.Minute * 30).Unix()

	accessString, err := token.SignedString(signingKey)

	if err != nil {
		return "", "", err
	}

	// Refresh token
	refreshToken := jwt.New(jwt.SigningMethodHS256)
	refreshClaims := refreshToken.Claims.(jwt.MapClaims)

	refreshClaims["authorized"] = true
	refreshClaims["client"] = user.Email
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
	// cookie := http.Cookie{Name: "Refreshtoken", Value: "", Path: "/",
	// 	MaxAge: -1, HttpOnly: true}
	// http.SetCookie(w, &cookie)
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

func newAccessToken(email string) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)

	claims["authorized"] = true
	claims["client"] = email
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

func hitPaystack(email string, amount int) (r config.PaystackResponse, err error) {
	p := config.GeneratePayment{
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
		return config.PaystackResponse{}, err
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

func paystackVerify(reference string) (res config.Transactions, attempted bool) {

	var verifyLink = "https://api.paystack.co/transaction/verify/" + reference

	req, err := http.NewRequest("GET", verifyLink, bytes.NewBuffer([]byte{}))
	if err != nil {
		log.Println(err)
	}
	value := "Bearer " + config.Paystack_key
	req.Header.Set("Content-Type", "application/json")
	req.Header.Add("Authorization", value)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Error making request to paystack, ", err)
		return
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("Error reading the body, ", err)
		return
	}

	var verifyPayload verifyPaystackResponse
	err = json.Unmarshal(body, &verifyPayload)
	if err != nil {
		log.Println("Error Unmarshalling json, ", err)
		return
	}

	res.Reference = reference
	res.Payment_channel = verifyPayload.Data.Channel
	res.Amount = verifyPayload.Data.Amount
	if verifyPayload.Data.Status == "failed" || verifyPayload.Data.Status == "success" {
		res.Verify_status = true
		res.Payment_status = true
		return res, true
	}
	return res, false
}
