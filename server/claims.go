package server

import "github.com/golang-jwt/jwt/v4"

type Claims struct {
	GUID string
	jwt.RegisteredClaims
}
