package health

import (
	"encoding/json"
	"net/http"
)

// NewHealthIndicatorHttpHandlerFunc creates a new HTTP handler function for the given indicator.
func NewHealthIndicatorHttpHandlerFunc(indicator Indicator) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		health := indicator.Health()
		w.Header().Set("Content-Type", "application/json")
		if health.Status != Up {
			w.WriteHeader(http.StatusServiceUnavailable)
		}
		json.NewEncoder(w).Encode(health)
	}
}
