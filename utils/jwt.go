package utils

import (
	"errors"
	"github.com/golang-jwt/jwt/v4"
	"os"
	"time"
)

type Claims struct {
	Userid string `json:"userid"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

func GenerateToken(userid, role string) (string, error) {
	secret := []byte(os.Getenv("JWT_SECRET"))
	expirationTime := time.Now().Add(time.Minute * 15).Unix()
	claims := Claims{
		Userid: userid,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Unix(expirationTime, 0)),
			Issuer:    "lw",
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}

func ParseToken(token string) (*Claims, error) {
	secret := os.Getenv("JWT_SECRET")
	tokenClaims, err := jwt.ParseWithClaims(token, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil {
		var ve *jwt.ValidationError
		if errors.As(err, &ve) {
			if ve.Errors&jwt.ValidationErrorMalformed != 0 {
				return nil, errors.New("that's not even a token")
			}
			if ve.Errors&(jwt.ValidationErrorExpired|jwt.ValidationErrorNotValidYet) != 0 {
				return nil, errors.New("token is expired")
			}
		}
		return nil, err
	}
	if claims, ok := tokenClaims.Claims.(*Claims); ok && tokenClaims.Valid {
		return claims, nil
	}
	return nil, errors.New("token is invalid")
}
