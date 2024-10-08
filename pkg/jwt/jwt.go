package jwt

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"

	"shortbin/pkg/config"
	"shortbin/pkg/logger"
)

const (
	LoginTokenExpiryTime    = 5 * 60 * 60   // 5 hours
	RefreshTokenExpiryTime  = 7 * 24 * 3600 // 7 days
	ResetPasswordExpiryTime = 15 * 60       // 15 minutes
	LoginTokenType          = "x-access"
	RefreshTokenType        = "x-refresh"
	ResetPasswordTokenType  = "x-reset"
)

func GenerateAccessToken(payload map[string]interface{}, scope string) string {
	var exp int64
	if scope == LoginTokenType {
		exp = time.Now().Add(time.Second * LoginTokenExpiryTime).Unix()
	} else if scope == ResetPasswordTokenType {
		exp = time.Now().Add(time.Second * ResetPasswordExpiryTime).Unix()
	}

	payload["type"] = scope
	tokenContent := jwt.MapClaims{
		"payload": payload,
		"exp":     exp,
	}
	jwtToken := jwt.NewWithClaims(jwt.GetSigningMethod("HS256"), tokenContent)
	token, err := jwtToken.SignedString([]byte(config.GetConfig().AuthSecret))
	if err != nil {
		logger.Error("Failed to generate access token: ", err)
		return ""
	}

	return token
}

func GenerateRefreshToken(payload map[string]interface{}) string {
	payload["type"] = RefreshTokenType
	tokenContent := jwt.MapClaims{
		"payload": payload,
		"exp":     time.Now().Add(time.Second * RefreshTokenExpiryTime).Unix(),
	}
	jwtToken := jwt.NewWithClaims(jwt.GetSigningMethod("HS256"), tokenContent)
	token, err := jwtToken.SignedString([]byte(config.GetConfig().AuthSecret))
	if err != nil {
		logger.Error("Failed to generate refresh token: ", err)
		return ""
	}

	return token
}

func ValidateToken(jwtToken string) (map[string]interface{}, error) {
	cleanJWT := strings.Replace(jwtToken, "Bearer ", "", 1)
	tokenData := jwt.MapClaims{}
	//nolint:revive
	token, err := jwt.ParseWithClaims(cleanJWT, tokenData, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.GetConfig().AuthSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, jwt.ErrInvalidKey
	}

	b, err := json.Marshal(tokenData["payload"])
	if err != nil {
		return nil, err
	}

	var payload map[string]interface{}
	err = json.Unmarshal(b, &payload)
	if err != nil {
		return nil, err
	}

	return payload, nil
}
