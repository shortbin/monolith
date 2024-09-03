package jwt

import (
	"encoding/json"
	"shortbin/pkg/logger"
	"strings"
	"time"

	"github.com/golang-jwt/jwt"
	"shortbin/pkg/config"
)

const (
	AccessTokenExpiredTime  = 5 * 60 * 60   // 5 hours
	RefreshTokenExpiredTime = 7 * 24 * 3600 // 7 days
	AccessTokenType         = "x-access"
	RefreshTokenType        = "x-refresh"
)

func GenerateAccessToken(payload map[string]interface{}) string {
	cfg := config.GetConfig()
	payload["type"] = AccessTokenType
	tokenContent := jwt.MapClaims{
		"payload": payload,
		"exp":     time.Now().Add(time.Second * AccessTokenExpiredTime).Unix(),
	}
	jwtToken := jwt.NewWithClaims(jwt.GetSigningMethod("HS256"), tokenContent)
	token, err := jwtToken.SignedString([]byte(cfg.AuthSecret))
	if err != nil {
		logger.Error("Failed to generate access token: ", err)
		return ""
	}

	return token
}

func GenerateRefreshToken(payload map[string]interface{}) string {
	cfg := config.GetConfig()
	payload["type"] = RefreshTokenType
	tokenContent := jwt.MapClaims{
		"payload": payload,
		"exp":     time.Now().Add(time.Second * RefreshTokenExpiredTime).Unix(),
	}
	jwtToken := jwt.NewWithClaims(jwt.GetSigningMethod("HS256"), tokenContent)
	token, err := jwtToken.SignedString([]byte(cfg.AuthSecret))
	if err != nil {
		logger.Error("Failed to generate refresh token: ", err)
		return ""
	}

	return token
}

func ValidateToken(jwtToken string) (map[string]interface{}, error) {
	cfg := config.GetConfig()
	cleanJWT := strings.Replace(jwtToken, "Bearer ", "", 1)
	tokenData := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(cleanJWT, tokenData, func(token *jwt.Token) (interface{}, error) {
		return []byte(cfg.AuthSecret), nil
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
