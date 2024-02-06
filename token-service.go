package main

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
)

type TokenService interface {
	GenerateToken(user *User) (string, error)
	ValidateToken(token string) (*User, error)
	RejectToken(token string) error
	GetRejectedTokens() ([]string, error)
}

type tokenService struct {
	secret         string
	rejectedTokens []string
}

// NewTokenService returns a new token service.
func NewTokenService(secret string) TokenService {
	return &tokenService{secret: secret}
}

// GenerateToken generates a JWT token for the given user.
// The token is valid for 10 minutes.
func (t *tokenService) GenerateToken(user *User) (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)
	claims := token.Claims.(jwt.MapClaims)
	claims["user_id"] = user.ID
	claims["role"] = user.Role
	claims["name"] = user.Name
	claims["exp"] = time.Now().Add(time.Minute * 10).Unix()
	claims["jti"] = uuid.NewString()
	tokenString, err := token.SignedString([]byte(t.secret))
	return tokenString, err
}

// ValidateToken validates the given JWT token.
// If the token is valid, it returns the user associated with the token.
// If the token is not valid, it returns an error.
func (t *tokenService) ValidateToken(tokenString string) (*User, error) {
	// parse the token string
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(t.secret), nil
	})
	if err != nil {
		return nil, err
	}
	// check if tokens jti claim is on rejected list
	for _, rejectedJTI := range t.rejectedTokens {
		if rejectedJTI == token.Claims.(jwt.MapClaims)["jti"].(string) {
			return nil, fmt.Errorf("token is rejected")
		}
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return &User{
				ID:   string(claims["user_id"].(string)),
				Name: string(claims["name"].(string))},
			nil
	}
	return nil, fmt.Errorf("invalid token")
}

// RejectToken decodes token and rejects the given JWT token by adding it's id to the list of rejected tokens.
func (t *tokenService) RejectToken(tokenString string) error {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(t.secret), nil
	})
	if err != nil {
		return err
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		t.rejectedTokens = append(t.rejectedTokens, string(claims["jti"].(string)))
	}
	return nil
}

// GetRejectedTokens returns the list of rejected tokens.
func (t *tokenService) GetRejectedTokens() ([]string, error) {
	return t.rejectedTokens, nil
}
