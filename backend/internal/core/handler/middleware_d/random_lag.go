package middleware_d

import (
	"math/rand/v2"
	"net/http"
	"time"
)

func RandomLag(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lag := 300 + rand.IntN(201) // 300~500ms
		time.Sleep(time.Duration(lag) * time.Millisecond)
		next.ServeHTTP(w, r)
	})
}
