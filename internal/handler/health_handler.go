package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"log/slog"
)

type HealthHandler struct {
	logger *slog.Logger
}

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{
		logger: slog.Default(),
	}
}

// HealthCheck returns the health status of the API
func (h *HealthHandler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	h.respondJSON(w, http.StatusOK, map[string]interface{}{
		"status":    "ok",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

func (h *HealthHandler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error("failed to encode response", slog.String("error", err.Error()))
	}
}
