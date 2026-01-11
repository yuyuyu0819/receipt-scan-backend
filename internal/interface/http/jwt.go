package http

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

type jwtClaims struct {
	Subject int64 `json:"sub"`
	Expires int64 `json:"exp"`
	Issued  int64 `json:"iat"`
}

type jwtHeader struct {
	Algorithm string `json:"alg"`
	Type      string `json:"typ"`
}

var errInvalidToken = errors.New("invalid token")

func buildJWT(secret string, userID int64, now time.Time, ttl time.Duration) (string, error) {
	if secret == "" {
		return "", errors.New("secret is empty")
	}
	if userID <= 0 {
		return "", errors.New("userID must be positive")
	}
	headerBytes, err := json.Marshal(jwtHeader{Algorithm: "HS256", Type: "JWT"})
	if err != nil {
		return "", fmt.Errorf("failed to encode header: %w", err)
	}
	claims := jwtClaims{
		Subject: userID,
		Expires: now.Add(ttl).Unix(),
		Issued:  now.Unix(),
	}
	claimsBytes, err := json.Marshal(claims)
	if err != nil {
		return "", fmt.Errorf("failed to encode claims: %w", err)
	}
	headerPart := base64.RawURLEncoding.EncodeToString(headerBytes)
	payloadPart := base64.RawURLEncoding.EncodeToString(claimsBytes)
	signingInput := headerPart + "." + payloadPart
	signature := signJWT(secret, signingInput)
	return signingInput + "." + signature, nil
}

func parseJWT(secret string, token string, now time.Time) (jwtClaims, error) {
	if secret == "" {
		return jwtClaims{}, errInvalidToken
	}
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return jwtClaims{}, errInvalidToken
	}
	signingInput := parts[0] + "." + parts[1]
	expectedSig := signJWT(secret, signingInput)
	if !hmac.Equal([]byte(expectedSig), []byte(parts[2])) {
		return jwtClaims{}, errInvalidToken
	}
	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return jwtClaims{}, errInvalidToken
	}
	var claims jwtClaims
	if err := json.Unmarshal(payloadBytes, &claims); err != nil {
		return jwtClaims{}, errInvalidToken
	}
	if claims.Subject <= 0 {
		return jwtClaims{}, errInvalidToken
	}
	if claims.Expires > 0 && now.Unix() > claims.Expires {
		return jwtClaims{}, errInvalidToken
	}
	return claims, nil
}

func signJWT(secret string, data string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(data))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}
