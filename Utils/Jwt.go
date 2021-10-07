package NoQ_RoomQ_Utils

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/golang-jwt/jwt"
)

type JwtClaims struct {
	RoomID    string `json:"room_id"`
	SessionID string `json:"session_id"`
	Type      string `json:"type"`
	jwt.StandardClaims
}

func (claim JwtClaims) jsonify() JSON {
	mapping := map[string]interface{}{
		"room_id":    claim.RoomID,
		"session_id": claim.SessionID,
		"type":       claim.Type,
	}
	return JSON{Val: mapping}
}

func JwtEncode(claims JwtClaims, jwtSecret string) string {
	if val, err := json.Marshal(claims); err == nil {
		fmt.Println(string(val))
	}
	tokenClaims := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	if token, err := tokenClaims.SignedString([]byte(jwtSecret)); err == nil {
		return token
	} else {
		panic("failed to generate JWT")
	}
}

func JwtDecode(token string, jwtSecret string) (jwtData JSON, ok bool) {
	if tokenClaims, err := jwt.ParseWithClaims(token, &JwtClaims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte(jwtSecret), nil
	}); err != nil {
		var message string
		if ve, ok := err.(*jwt.ValidationError); ok {
			if ve.Errors&jwt.ValidationErrorMalformed != 0 {
				message = "token is malformed"
			} else if ve.Errors&jwt.ValidationErrorUnverifiable != 0 {
				message = "token could not be verified because of signing problems"
			} else if ve.Errors&jwt.ValidationErrorSignatureInvalid != 0 {
				message = "signature validation failed"
			} else if ve.Errors&jwt.ValidationErrorExpired != 0 {
				message = "token is expired"
			} else if ve.Errors&jwt.ValidationErrorNotValidYet != 0 {
				message = "token is not yet valid before sometime"
			} else {
				message = "can not handle this token"
			}
		}
		log.Println("Failed to decode jwt: " + message)
		return JSON{}, false
	} else {
		if claim, ok := tokenClaims.Claims.(*JwtClaims); ok && tokenClaims.Valid {
			return claim.jsonify(), true
		} else {
			log.Println("Failed to validate JWT")
			return JSON{}, false
		}
	}
}
