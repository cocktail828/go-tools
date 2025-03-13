package healthy

import (
	"encoding/json"
	"net/http"
)

type HealthChecker struct {
	Ready   bool   `json:"ready,omitempty"`
	Message string `json:"message,omitempty"`
}

func (hc HealthChecker) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if hc.Ready {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusForbidden) // k8s readiness check may use http.code
	}

	json.NewEncoder(w).Encode(hc)
}
