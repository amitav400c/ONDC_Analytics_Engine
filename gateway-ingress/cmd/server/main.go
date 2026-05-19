// ONDC Analytics Gateway — Ingress Server
// Receives Beckn protocol webhook payloads, optionally sanitizes via edge sandbox,
// and publishes to Redpanda for downstream OLAP ingestion.

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

	"github.com/amitav400c/ondc-analytics-gateway/gateway-ingress/internal/handler"
	"github.com/amitav400c/ondc-analytics-gateway/gateway-ingress/internal/producer"
	"github.com/amitav400c/ondc-analytics-gateway/gateway-ingress/internal/sandbox"
	"github.com/amitav400c/ondc-analytics-gateway/gateway-ingress/internal/store"
)

func main() {
	port := getEnv("GATEWAY_PORT", "8080")
	brokers := getEnv("REDPANDA_BROKERS", "localhost:19092")
	topic := getEnv("REDPANDA_TOPIC", "ondc_events_sanitized")
	socketPath := getEnv("SANDBOX_SOCKET_PATH", "/tmp/sandbox.sock")

	// Initialize Redpanda producer
	prod, err := producer.New(brokers, topic)
	if err != nil {
		log.Fatalf("failed to create producer: %v", err)
	}
	defer prod.Close()

	// Initialize edge sandbox gRPC client (graceful degradation if unavailable)
	sbx := sandbox.NewClient(socketPath)
	defer sbx.Close()

	redisAddr := getEnv("REDIS_ADDR", "localhost:6379")

	// Initialize Redis for rate limiting & quotas
	redisStore, err := store.NewRedisStore(redisAddr)
	if err != nil {
		log.Printf("warning: redis connection failed (%v), rate limiting will be disabled", err)
		redisStore = nil // handled gracefully in handler if nil
	}

	// Wire up handler
	wh := handler.NewWebhook(prod, sbx, redisStore)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RealIP)
	r.Use(middleware.Timeout(30 * time.Second))

	r.Post("/webhooks/ondc", wh.Handle)
	r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok","service":"gateway-ingress"}`))
	})

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	go func() {
		log.Printf("gateway-ingress listening on :%s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	log.Println("shutting down gateway-ingress...")
	srv.Shutdown(ctx)
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
