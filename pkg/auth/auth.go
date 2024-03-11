package auth

import (
	"net/http"
)

const authHeader = "Authorization"

func Authenticate(authKey string) func(handler http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			for _, key := range r.Header.Values(authHeader) {
				if key == "Bearer "+authKey {
					next.ServeHTTP(w, r)
					return
				}
			}
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		})
	}
}
