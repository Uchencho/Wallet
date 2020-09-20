package auth

import (
	"testing"
)

func TestHashPassword(t *testing.T) {
	_, err := HashPassword("myStrongPassword")
	if err != nil {
		t.Fatalf("Failed to hash password with error %s", err)
	}

	_, err = HashPassword("")
	if err == nil {
		t.Fatalf("Hashing an invalid password")
	}
}

func TestCheckPasswordHash(t *testing.T) {

	hashedPassword, err := HashPassword("myStrongPassword")
	if err != nil {
		t.Errorf("Strange, unable to hash password with error %s", err)
	}

	if equal := CheckPasswordHash("myStrongPassword", hashedPassword); !equal {
		t.Fatalf("Checkpassword is not retrieving current password")
	}
}

func TestGenerateAuthToken(t *testing.T) {

	_, _, err := GenerateAuthTokens("uche@gmail.com")
	if err != nil {
		t.Fatalf("Could not generate tokens with error, %s", err)
	}

	_, _, err = GenerateAuthTokens("")
	if err == nil {
		t.Fatalf("Generating token for an empty string as email")
	}
}

func TestCheckRefreshToken(t *testing.T) {

	_, refreshToken, err := GenerateAuthTokens("uche@gmail.com")
	if err != nil {
		t.Errorf("Could not generate token with error, %s", err)
	}

	ok, _, err := CheckRefreshToken(refreshToken)
	if !ok {
		t.Fatalf("Could not refresh token with error, %s", err)
	}

	ok, _, err = CheckRefreshToken("invalid token")
	if ok {
		t.Fatalf("Refreshing an invalid token, %s", err)
	}
}
