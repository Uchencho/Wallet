package app

import (
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

func generateAuthTokens(user models.Accounts) (string, string, error) {

	// Access token
	signingKey := []byte("3d67d77426d9878967a177437316554b0088fa88be95846252011528e8bad788")

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
	refreshSigningKey := []byte("b178604f6216f904f394641fd167078e426d5fe9ce20d4c07a65e8dd051a40d9")

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
