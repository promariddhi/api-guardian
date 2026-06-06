package middleware

import (
	"context"
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	Subject string `json:"sub"`
	Role    string `json:"role"`
	jwt.RegisteredClaims
}

func Auth(next http.Handler, allowedRoles []string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authorization := r.Header.Get("Authorization")
		parts := strings.Fields(authorization)
		ctx := r.Context()
		if authorization == "" || len(parts) < 2 || (len(parts) >= 2 && parts[0] != "Bearer") {
			http.Error(w, "missing authorization error", http.StatusUnauthorized)
			return
		}
		if claims, ok := validateToken(parts[1]); ok {
			if !roleAllowed(claims.Role, allowedRoles) {
				http.Error(w, "forbidden", http.StatusForbidden)
			}
			ctx = context.WithValue(r.Context(), "claims", claims)
		} else {
			http.Error(w, "missing authorization error", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func validateToken(tokenString string) (*Claims, bool) {
	secret := []byte(os.Getenv("JWT_SECRET"))
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (any, error) {
		return secret, nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	if err != nil {
		return &Claims{}, false
	}
	if claims, ok := token.Claims.(*Claims); ok {
		return claims, true
	} else {
		return &Claims{}, false
	}
}

func roleAllowed(role string, allowedRoles []string) bool {
	if len(allowedRoles) == 0 {
		return true
	}
	for _, r := range allowedRoles {
		if r == role {
			return true
		}
	}
	return false
}
