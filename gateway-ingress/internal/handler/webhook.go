package handler

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/amitav400c/ondc-analytics-gateway/gateway-ingress/internal/producer"
	"github.com/amitav400c/ondc-analytics-gateway/gateway-ingress/internal/sandbox"
)

type Webhook struct {
	prod *producer.Producer
	sbx  *sandbox.Client
}

func NewWebhook(prod *producer.Producer, sbx *sandbox.Client) *Webhook {
	return &Webhook{prod: prod, sbx: sbx}
}

// BecknPayload is a minimal representation of a Beckn protocol message.
// We only parse what we need; the full JSON goes to Redpanda.
type BecknPayload struct {
	Context struct {
		Action      string `json:"action"`
		Domain      string `json:"domain"`
		City        string `json:"city"`
		TransactionID string `json:"transaction_id"`
		Timestamp   string `json:"timestamp"`
	} `json:"context"`
	Message json.RawMessage `json:"message"`
}

// FlatEvent is the denormalized event pushed to Redpanda/ClickHouse
type FlatEvent struct {
	EventID    string  `json:"event_id"`
	EventType  string  `json:"event_type"`
	Action     string  `json:"action"`
	City       string  `json:"city"`
	Timestamp  string  `json:"timestamp"`
	OrderID    string  `json:"order_id"`
	SellerID   string  `json:"seller_id"`
	BuyerHash  string  `json:"buyer_hash"`
	GPSLat     float64 `json:"gps_lat"`
	GPSLng     float64 `json:"gps_lng"`
	Amount     float64 `json:"amount"`
	Status     string  `json:"status"`
	Domain           string  `json:"domain"`
	RawPayload       string  `json:"raw_payload"`
	SandboxLatencyMs float64 `json:"sandbox_latency_ms"`
	KafkaLatencyMs   float64 `json:"kafka_latency_ms"`
	TotalLatencyMs   float64 `json:"total_latency_ms"`
}

func (wh *Webhook) Handle(w http.ResponseWriter, r *http.Request) {
	totalStart := time.Now()

	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20)) // 1MB limit
	if err != nil {
		http.Error(w, `{"error":"failed to read body"}`, http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var payload BecknPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		http.Error(w, `{"error":"invalid JSON"}`, http.StatusBadRequest)
		return
	}

	if payload.Context.Action == "" {
		http.Error(w, `{"error":"missing context.action"}`, http.StatusBadRequest)
		return
	}

	// Attempt PII redaction and WAF checks via edge sandbox
	sanitized := string(body)
	var redactErr error

	sandboxStart := time.Now()
	if wh.sbx.IsConnected() {
		result, isSafe, reason, err := wh.sbx.Sanitize(r.Context(), sanitized)
		if err != nil {
			log.Printf("sandbox redaction failed (passthrough): %v", err)
			// Fallback: do basic redaction in Go
			sanitized, redactErr = basicRedact(sanitized)
		} else {
			if !isSafe {
				log.Printf("WAF blocked request: %s", reason)
				http.Error(w, fmt.Sprintf(`{"error":"%s"}`, reason), http.StatusBadRequest)
				return
			}
			sanitized = result
			redactErr = nil
		}
	} else {
		sanitized, redactErr = basicRedact(string(body))
	}
	sandboxLatency := float64(time.Since(sandboxStart).Microseconds()) / 1000.0

	if redactErr != nil {
		log.Printf("redaction completely failed: %v", redactErr)
		http.Error(w, `{"error":"invalid payload for redaction"}`, http.StatusBadRequest)
		return
	}

	// Extract fields for the flat event
	event := extractEvent(payload, sanitized)
	event.SandboxLatencyMs = sandboxLatency
	event.TotalLatencyMs = float64(time.Since(totalStart).Microseconds()) / 1000.0
	// We can't know Kafka publish latency before we publish, so we leave it as 0 for this event

	eventJSON, err := json.Marshal(event)
	if err != nil {
		http.Error(w, `{"error":"internal error"}`, http.StatusInternalServerError)
		return
	}

	if err := wh.prod.Publish(r.Context(), event.EventID, eventJSON); err != nil {
		log.Printf("publish error: %v", err)
		http.Error(w, `{"error":"failed to queue event"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte(`{"status":"accepted","event_id":"` + event.EventID + `"}`))
}

func extractEvent(p BecknPayload, sanitizedJSON string) FlatEvent {
	ts := p.Context.Timestamp
	if ts == "" {
		ts = time.Now().UTC().Format("2006-01-02T15:04:05.000Z")
	}
	return FlatEvent{
		EventID:          fmt.Sprintf("%s-%d", p.Context.TransactionID, time.Now().UnixNano()),
		EventType:        p.Context.Action,
		Action:           p.Context.Action,
		City:             p.Context.City,
		Timestamp:        ts,
		OrderID:          p.Context.TransactionID,
		SellerID:         extractField(sanitizedJSON, "seller_id"),
		BuyerHash:        extractField(sanitizedJSON, "buyer_hash"),
		GPSLat:           extractFloat(sanitizedJSON, "gps_lat"),
		GPSLng:           extractFloat(sanitizedJSON, "gps_lng"),
		Amount:           extractFloat(sanitizedJSON, "amount"),
		Status:           "active",
		Domain:           p.Context.Domain,
		RawPayload:       sanitizedJSON,
		SandboxLatencyMs: 0, // Set later
		KafkaLatencyMs:   0, // Set later
		TotalLatencyMs:   0, // Set later
	}
}

// basicRedact performs simple PII redaction when the WASM sandbox is unavailable.
// Returns an error if the JSON is invalid, ensuring we fail-closed to prevent PII leaks.
func basicRedact(payload string) (string, error) {
	var raw map[string]interface{}
	if err := json.Unmarshal([]byte(payload), &raw); err != nil {
		return "", fmt.Errorf("failed to parse json for redaction: %w", err)
	}
	redactRecursive(raw)
	out, err := json.Marshal(raw)
	if err != nil {
		return "", fmt.Errorf("failed to marshal redacted json: %w", err)
	}
	return string(out), nil
}

func redactRecursive(m map[string]interface{}) {
	for k, v := range m {
		switch k {
		case "phone", "mobile", "contact":
			if s, ok := v.(string); ok {
				h := sha256.Sum256([]byte(s))
				m[k] = fmt.Sprintf("%x", h[:8]) // Short hash
			}
		case "gps":
			if s, ok := v.(string); ok {
				m[k] = fuzzGPS(s)
			}
		}
		if nested, ok := v.(map[string]interface{}); ok {
			redactRecursive(nested)
		}
	}
}

func fuzzGPS(gps string) string {
	var lat, lng float64
	fmt.Sscanf(gps, "%f,%f", &lat, &lng)
	// Add random noise ±0.01 (~1km)
	lat += (rand.Float64() - 0.5) * 0.02
	lng += (rand.Float64() - 0.5) * 0.02
	return fmt.Sprintf("%.4f,%.4f", lat, lng)
}

func extractField(jsonStr, field string) string {
	var m map[string]interface{}
	json.Unmarshal([]byte(jsonStr), &m)
	if v, ok := m[field]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func extractFloat(jsonStr, field string) float64 {
	var m map[string]interface{}
	json.Unmarshal([]byte(jsonStr), &m)
	if v, ok := m[field]; ok {
		if f, ok := v.(float64); ok {
			return f
		}
	}
	return 0
}
