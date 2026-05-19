// ONDC Analytics API — REST server bridging ClickHouse OLAP data to the dashboard.
// Endpoints: /api/v1/health, /api/v1/metrics/funnel, /api/v1/metrics/cancellations, /api/v1/auth/login

package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/amitav400c/ondc-analytics-gateway/analytics-api/internal/handler"
	"github.com/amitav400c/ondc-analytics-gateway/analytics-api/internal/store"
)

func main() {
	port := getEnv("API_PORT", "8081")
	chAddr := getEnv("CLICKHOUSE_ADDR", "localhost:9000")
	chDB := getEnv("CLICKHOUSE_DB", "ondc")
	jwtSecret := getEnv("JWT_SECRET", "change-me-in-production")
	pgHost := getEnv("POSTGRES_HOST", "localhost")
	pgPort := getEnv("POSTGRES_PORT", "5432")
	pgUser := getEnv("POSTGRES_USER", "ondc")
	pgPass := getEnv("POSTGRES_PASSWORD", "ondc_secret")
	pgDB := getEnv("POSTGRES_DB", "ondc_auth")

	ch, err := store.NewClickHouse(chAddr, chDB)
	if err != nil {
		log.Fatalf("clickhouse connection failed: %v", err)
	}

	pg, err := store.NewPostgres(pgHost, pgPort, pgUser, pgPass, pgDB)
	if err != nil {
		log.Fatalf("postgres connection failed: %v", err)
	}
	defer pg.Close()

	h := handler.New(ch, pg, jwtSecret)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173", "http://localhost:3000", "*"},
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Public routes
	r.Get("/api/v1/health", h.Health)
	r.Post("/api/v1/auth/login", h.Login)

	// Protected routes
	r.Group(func(r chi.Router) {
		r.Use(h.AuthMiddleware)
		r.Get("/api/v1/metrics/funnel", h.Funnel)
		r.Get("/api/v1/metrics/cancellations", h.Cancellations)
		r.Get("/api/v1/metrics/volume", h.Volume)
		r.Get("/api/v1/events/recent", h.RecentEvents)
	})

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	go func() {
		log.Printf("analytics-api listening on :%s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	log.Println("shutting down analytics-api...")
	srv.Shutdown(ctx)
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
