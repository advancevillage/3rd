package jwtx

import (
	"errors"

	"github.com/golang-jwt/jwt/v5"
)

var (
	JWTX_PARSE_ERROR = errors.New("failed to parse token")
)

type jwtxCtx struct {
	Payload string `json:"payload"`
	jwt.RegisteredClaims
}
