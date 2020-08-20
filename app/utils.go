package app

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Uchencho/wallet/models"
	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
)

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

func generateAuthTokens(user models.Accounts) (string, string, error) {

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

func IsAuthorized(endpoint func(http.ResponseWriter, *http.Request)) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if r.Header["Token"] != nil {
			token, err := jwt.Parse(r.Header["Token"][0], func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("There was an error")
				}
				return signingKey, nil
			})

			if err != nil {
				fmt.Fprint(w, err.Error())
			}

			if token.Valid {
				endpoint(w, r)
			}
		} else {
			fmt.Fprint(w, "Not Authorized")
		}
	})
}

func checkAuth(r *http.Request) (bool, error) {
	if r.Header["Authorization"] != nil {
		accessToken := strings.Split(r.Header["Authorization"][0], " ")[1]
		token, err := jwt.Parse(accessToken, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("An error occured")
			}
			return signingKey, nil
		})
		if err != nil {
			fmt.Println(err)
			return false, err
		}
		if token.Valid {
			return true, nil
		}
	}
	fmt.Println(r.Header["Authorization"])
	return false, errors.New("Credentials not provided")
}
