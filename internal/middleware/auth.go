package middleware

import (
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authorization := r.Header.Get("Authorization")
		parts := strings.Fields(authorization)
		if authorization == "" || len(parts) < 2 || (len(parts) >= 2 && parts[0] != "Bearer") {
			http.Error(w, "missing authorization error", http.StatusUnauthorized)
			return
		}
		if validateToken(parts[1]) {
			log.Println("Valid token")
		} else {
			http.Error(w, "missing authorization error", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func validateToken(tokenString string) bool {
	secret := []byte(os.Getenv("JWT_SECRET"))
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		return secret, nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	if err != nil {
		return false
	}
	return token.Valid
}
