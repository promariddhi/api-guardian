package middleware

import (
	"log"
	"net/http"
	"strings"
)

func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authorization := r.Header.Get("Authorization")
		parts := strings.Split(authorization, " ")
		if authorization == "" || len(parts) < 2 || (len(parts) >= 2 && parts[0] != "Bearer") {
			http.Error(w, "missing authorization error", http.StatusUnauthorized)
			return
		}
		if strings.Compare(parts[1], "pass") == 0 {
			log.Println("Correct")
		} else {
			http.Error(w, "missing authorization error", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}
