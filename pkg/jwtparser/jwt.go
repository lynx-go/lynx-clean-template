package jwtparser

import (
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
)

func ParseToken(hmacSecretByte []byte, tokenString string) (userID string, username string, vars map[string]string, exp int64, tokenId string, issuedAt int64, ok bool) {
	jwtToken, err := jwt.ParseWithClaims(tokenString, &SessionTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		return hmacSecretByte, nil
	}, jwt.WithExpirationRequired(), jwt.WithValidMethods([]string{"HS256"}))
	if err != nil {
		return
	}
	claims, ok := jwtToken.Claims.(*SessionTokenClaims)
	if !ok || !jwtToken.Valid {
		return
	}
	return claims.UserId, claims.Username, claims.Vars, claims.ExpiresAt, claims.TokenId, claims.IssuedAt, true
}

type SessionTokenClaims struct {
	TokenId   string            `json:"tid,omitempty"`
	UserId    string            `json:"uid,omitempty"`
	Username  string            `json:"usn,omitempty"`
	Vars      map[string]string `json:"vrs,omitempty"`
	ExpiresAt int64             `json:"exp,omitempty"`
	IssuedAt  int64             `json:"iat,omitempty"`
}

func (s *SessionTokenClaims) GetExpirationTime() (*jwt.NumericDate, error) {
	return jwt.NewNumericDate(time.Unix(s.ExpiresAt, 0)), nil
}
func (s *SessionTokenClaims) GetNotBefore() (*jwt.NumericDate, error) {
	return nil, nil
}
func (s *SessionTokenClaims) GetIssuedAt() (*jwt.NumericDate, error) {
	return jwt.NewNumericDate(time.Unix(s.IssuedAt, 0)), nil
}
func (s *SessionTokenClaims) GetAudience() (jwt.ClaimStrings, error) {
	return []string{}, nil
}
func (s *SessionTokenClaims) GetIssuer() (string, error) {
	return "", nil
}
func (s *SessionTokenClaims) GetSubject() (string, error) {
	return s.UserId, nil
}
