package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
	persistenceStrategy "github.com/jordanoluz/goexpert-rate-limiter/internal/infra/persistence_strategy"
	"github.com/jordanoluz/goexpert-rate-limiter/pkg/middleware"
	rateLimiter "github.com/jordanoluz/goexpert-rate-limiter/pkg/rate_limiter"
)

const apiPort = 8080

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("failed to load .env file: %v", err)
	}

	rateLimitToken, err := strconv.Atoi(os.Getenv("RATE_LIMIT_TOKEN"))
	if err != nil {
		log.Fatalf("failed to parse 'RATE_LIMIT_TOKEN' to an int: %v", err)
	}

	rateLimitIP, err := strconv.Atoi(os.Getenv("RATE_LIMIT_IP"))
	if err != nil {
		log.Fatalf("failed to parse 'RATE_LIMIT_IP' to an int: %v", err)
	}

	blockDuration, err := strconv.Atoi(os.Getenv("BLOCK_DURATION"))
	if err != nil {
		log.Fatalf("failed to parse 'BLOCK_DURATION' to an int: %v", err)
	}

	persistenceStrategy, err := persistenceStrategy.NewRedisStrategy()
	if err != nil {
		log.Fatalf("failed to initialize redis persistence strategy: %v", err)
	}

	rateLimiter := rateLimiter.NewRateLimiter(rateLimitToken, rateLimitIP, time.Duration(blockDuration)*time.Second, persistenceStrategy)

	r := chi.NewRouter()

	r.Use(chiMiddleware.Logger)
	r.Use(chiMiddleware.RealIP)
	r.Use(middleware.RateLimiter(rateLimiter))

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("welcome"))
	})

	log.Printf("listening and serving on port: %d", apiPort)

	if err := http.ListenAndServe(fmt.Sprintf(":%d", apiPort), r); err != nil {
		log.Fatalf("failed to listen and serve: %v", err)
	}
}
