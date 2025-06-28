package auth

import (
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"time"
)

type RegistryClaims struct {
	Organization []string `json:"organization"`
	jwt.MapClaims
}

func CreateJWTToken(username string, key []byte) (*string, error) {
	claims := jwt.MapClaims{
		"sub":   "terraform-cli",
		"login": username,
		"iat":   time.Now().Add(time.Millisecond * -30).Unix(),
		"exp":   time.Now().Add(time.Hour * 24 * 356).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(key)
	if err != nil {
		return nil, fmt.Errorf("error signing the token: %s", err)
	}

	return &signedToken, nil
}

func CreateJWTClaimsToken(username string, organization []string, key []byte) (*string, error) {
	claims := RegistryClaims{
		Organization: organization,
		MapClaims: jwt.MapClaims{
			"sub":   "terraform-cli",
			"login": username,
			"iat":   time.Now().Add(time.Millisecond * -30).Unix(),
			"exp":   time.Now().Add(time.Hour * 24 * 356).Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(key)
	if err != nil {
		return nil, fmt.Errorf("error signing the token: %s", err)
	}

	return &signedToken, nil
}

func GetJWTToken(signedToken string, key []byte) (*jwt.Token, error) {
	token, err := jwt.Parse(signedToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return key, nil
	})

	if err != nil {
		return nil, fmt.Errorf("error parsing token: %s", err)
	}

	return token, nil
}

func GetJWTClaimsToken(signedToken string, key []byte) (*jwt.Token, error) {
	token, err := jwt.ParseWithClaims(signedToken, &RegistryClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return key, nil
	})
	if err != nil {
		return nil, fmt.Errorf("error parsing token: %s", err)
	}

	return token, nil
}
