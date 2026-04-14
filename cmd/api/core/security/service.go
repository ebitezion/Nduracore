package security

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type Claims struct {
	Subject string
	Role    string
	Expiry  time.Time
}

type customClaims struct {
	Role string `json:"role"`
	jwt.RegisteredClaims
}

type Service struct {
	TokenSecret   string
	TokenIssuer   string
	TokenAudience string
}

func (s Service) GenerateToken(subject, role string, ttl time.Duration) (string, error) {
	claims := customClaims{
		Role: role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   subject,
			Issuer:    s.TokenIssuer,
			Audience:  jwt.ClaimStrings{s.TokenAudience},
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.TokenSecret))
}

func (s Service) ParseToken(tokenString string) (Claims, error) {
	claims := customClaims{}

	token, err := jwt.ParseWithClaims(tokenString, &claims, func(token *jwt.Token) (interface{}, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(s.TokenSecret), nil
	}, jwt.WithIssuer(s.TokenIssuer), jwt.WithAudience(s.TokenAudience))
	if err != nil {
		return Claims{}, err
	}

	if !token.Valid {
		return Claims{}, errors.New("invalid token")
	}

	return Claims{
		Subject: claims.Subject,
		Role:    claims.Role,
		Expiry:  claims.ExpiresAt.Time,
	}, nil
}

func VerifyPasswordHash(plainPassword, passwordHash string) bool {
	if plainPassword == "" || passwordHash == "" {
		return false
	}

	err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(plainPassword))
	return err == nil
}
