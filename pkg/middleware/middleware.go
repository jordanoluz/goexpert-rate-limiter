package middleware

import (
	"net"
	"net/http"

	rateLimiter "github.com/jordanoluz/goexpert-rate-limiter/pkg/rate_limiter"
)

func RateLimiter(rl rateLimiter.RateLimiter) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := r.Header.Get("API_KEY")
			ip, _, _ := net.SplitHostPort(r.RemoteAddr)

			if token == "" && ip == "" {
				http.Error(w, "invalid request: missing token and ip", http.StatusBadRequest)
				return
			}

			if !rl.Allow(r.Context(), token, ip) {
				http.Error(w, "you have reached the maximum number of requests or actions allowed within a certain time frame", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
