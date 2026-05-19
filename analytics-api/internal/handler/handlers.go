package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/amitav400c/ondc-analytics-gateway/analytics-api/internal/store"
)

type Handler struct {
	ch        *store.ClickHouse
	pg        *store.Postgres
	jwtSecret []byte
}

func New(ch *store.ClickHouse, pg *store.Postgres, secret string) *Handler {
	return &Handler{ch: ch, pg: pg, jwtSecret: []byte(secret)}
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// --- Health ---

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	chOk := "up"
	if err := h.ch.Ping(r.Context()); err != nil {
		chOk = "down"
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":     "ok",
		"service":    "analytics-api",
		"clickhouse": chOk,
		"timestamp":  time.Now().UTC(),
	})
}

// --- Auth ---

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	user, err := h.pg.FindUserByEmail(req.Email)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		writeError(w, http.StatusUnauthorized, "invalid credentials")
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":   user.ID,
		"email": user.Email,
		"name":  user.Name,
		"role":  user.Role,
		"exp":   time.Now().Add(24 * time.Hour).Unix(),
	})

	tokenStr, err := token.SignedString(h.jwtSecret)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "token generation failed")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"token": tokenStr,
		"user": map[string]interface{}{
			"id":    user.ID,
			"email": user.Email,
			"name":  user.Name,
			"role":  user.Role,
		},
	})
}

// AuthMiddleware validates JWT from the Authorization header
func (h *Handler) AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			writeError(w, http.StatusUnauthorized, "missing or invalid authorization header")
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			return h.jwtSecret, nil
		})
		if err != nil || !token.Valid {
			writeError(w, http.StatusUnauthorized, "invalid token")
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			writeError(w, http.StatusUnauthorized, "invalid token claims")
			return
		}

		ctx := context.WithValue(r.Context(), "user_email", claims["email"])
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// --- Metrics ---

func (h *Handler) Funnel(w http.ResponseWriter, r *http.Request) {
	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")

	stages, err := h.ch.Funnel(r.Context(), from, to)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "query failed: "+err.Error())
		return
	}
	if stages == nil {
		stages = []store.FunnelStage{}
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"stages": stages})
}

func (h *Handler) Cancellations(w http.ResponseWriter, r *http.Request) {
	city := r.URL.Query().Get("city")
	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")

	rows, err := h.ch.Cancellations(r.Context(), city, from, to)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "query failed: "+err.Error())
		return
	}
	if rows == nil {
		rows = []store.CancellationRow{}
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"cancellations": rows})
}

func (h *Handler) Volume(w http.ResponseWriter, r *http.Request) {
	days := 7
	if d := r.URL.Query().Get("days"); d != "" {
		if parsed, err := strconv.Atoi(d); err == nil {
			days = parsed
		}
	}

	points, err := h.ch.Volume(r.Context(), days)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "query failed: "+err.Error())
		return
	}
	if points == nil {
		points = []store.VolumePoint{}
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"volume": points})
}

func (h *Handler) RecentEvents(w http.ResponseWriter, r *http.Request) {
	limit := 20
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil {
			limit = parsed
		}
	}

	events, err := h.ch.RecentEvents(r.Context(), limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "query failed: "+err.Error())
		return
	}
	if events == nil {
		events = []store.RecentEvent{}
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"events": events})
}
