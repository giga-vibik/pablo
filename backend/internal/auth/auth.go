// Package auth — авторизация одного внутреннего пользователя.
//
// Pablo — внутренний инструмент с одним набором аккаунтов, регистрации нет:
// логин и пароль лежат в конфиге, в обмен выдаётся JWT.
package auth

import (
	"crypto/subtle"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/pablo/backend/internal/config"
)

var ErrInvalidCredentials = errors.New("invalid login or password")

type AuthManager interface {
	Login(login, password string) (string, error)
	ValidateToken(tokenString string) (*jwt.Token, error)
	AuthMiddleware() func(http.Handler) http.Handler
}

type authManager struct {
	secret   []byte
	login    string
	password string
}

func NewAuthManager(cfg config.Auth) AuthManager {
	return &authManager{
		secret:   []byte(cfg.Secret),
		login:    cfg.Login,
		password: cfg.Password,
	}
}

func (a *authManager) Login(login, password string) (string, error) {
	loginOK := subtle.ConstantTimeCompare([]byte(login), []byte(a.login)) == 1
	passOK := subtle.ConstantTimeCompare([]byte(password), []byte(a.password)) == 1

	if !loginOK || !passOK {
		return "", ErrInvalidCredentials
	}

	claims := jwt.MapClaims{
		"login": login,
		"type":  "access",
		"exp":   time.Now().Add(30 * 24 * time.Hour).Unix(),
	}

	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(a.secret)
}

func (a *authManager) ValidateToken(tokenString string) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("error: unexpected signing method")
		}
		return a.secret, nil
	})
	if err != nil || !token.Valid {
		return nil, fmt.Errorf("error: invalid token: %v", err)
	}

	return token, nil
}

func (a *authManager) AuthMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				log.Println("AuthMiddleware: Missing Authorization header")
				http.Error(w, "Authorization header missing", http.StatusUnauthorized)
				return
			}

			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			if _, err := a.ValidateToken(tokenString); err != nil {
				log.Println("AuthMiddleware: Invalid token")
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
