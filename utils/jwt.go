package utils

import (
	"time"

	"github.com/Madhav-Gupta-28/0xmart-backend-go/config"
	"github.com/golang-jwt/jwt"
)

type Claims struct {
	UserID string `json:"userId"`
	jwt.StandardClaims
}

func GenerateJWT(userID string) (string, error) {
	claims := &Claims{
		UserID: userID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.GetEnv("JWT_SECRET", "")))
}

func ValidateJWT(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.GetEnv("JWT_SECRET", "")), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, jwt.ErrSignatureInvalid
	}

	return claims, nil
}
