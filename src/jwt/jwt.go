package jwt

import (
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
)

type JWT struct {
	ExpireTime   int64
	SecretKey    string
	Issuer       string
	CookieKey    string
	CookieDomain string
}

// NewTokenString create a new valid JWT
func (c *JWT) NewTokenString(payload map[string]interface{}) (string, error) {
	now := time.Now()
	payload["exp"] = now.Add(time.Duration(c.ExpireTime) * time.Hour).Unix()
	payload["iss"] = c.Issuer
	payload["iat"] = now.Unix()

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims(payload))
	return token.SignedString([]byte(c.SecretKey))
}

// GetTokenPayload get a token from JWT string
func (c *JWT) GetTokenPayload(tokenString string) (Payload, error) {
	p := jwt.Parser{UseJSONNumber: true}
	token, err := p.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(c.SecretKey), nil
	})
	if err != nil {
		return nil, err
	}

	// Success to return token
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return Payload(claims), nil
	}

	err = fmt.Errorf("invalid token: %v", token.Claims)
	return nil, err
}
