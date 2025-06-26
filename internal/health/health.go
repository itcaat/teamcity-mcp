package health

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"go.uber.org/zap"

	"github.com/itcaat/teamcity-mcp/internal/teamcity"
)

// Checker provides health check functionality
type Checker struct {
	tc     *teamcity.Client
	logger *zap.SugaredLogger
}

// New creates a new health checker
func New(tc *teamcity.Client, logger *zap.SugaredLogger) *Checker {
	return &Checker{
		tc:     tc,
		logger: logger,
	}
}

// LivenessHandler handles liveness probe requests
func (h *Checker) LivenessHandler(w http.ResponseWriter, r *http.Request) {
	// Simple liveness check - server is alive if it can respond
	response := map[string]interface{}{
		"status":    "ok",
		"timestamp": time.Now().UTC(),
		"service":   "teamcity-mcp",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// ReadinessHandler handles readiness probe requests
func (h *Checker) ReadinessHandler(w http.ResponseWriter, r *http.Request) {
	// Check if we can connect to TeamCity
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	status := "ok"
	statusCode := http.StatusOK
	checks := make(map[string]interface{})

	// Check TeamCity connectivity
	if err := h.checkTeamCity(ctx); err != nil {
		status = "error"
		statusCode = http.StatusServiceUnavailable
		checks["teamcity"] = map[string]interface{}{
			"status": "error",
			"error":  err.Error(),
		}
	} else {
		checks["teamcity"] = map[string]interface{}{
			"status": "ok",
		}
	}

	response := map[string]interface{}{
		"status":    status,
		"timestamp": time.Now().UTC(),
		"service":   "teamcity-mcp",
		"checks":    checks,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}

// checkTeamCity verifies TeamCity connectivity
func (h *Checker) checkTeamCity(ctx context.Context) error {
	// Try to list projects as a connectivity test
	_, err := h.tc.ListProjects(ctx)
	return err
}
