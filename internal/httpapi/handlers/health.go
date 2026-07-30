package handlers

import (
	"net/http"
	"time"
)

// HealthResponse representa el estado general de la API.
type HealthResponse struct {
	Status      string `json:"status"`
	Service     string `json:"service"`
	Environment string `json:"environment"`
	Timestamp   string `json:"timestamp"`
}

// Health devuelve un manejador para comprobar que la API está activa.
func Health(environment string) http.HandlerFunc {
	return func(
		w http.ResponseWriter,
		_ *http.Request,
	) {
		WriteJSON(
			w,
			http.StatusOK,
			HealthResponse{
				Status:      "ok",
				Service:     "SuperMarket SIG API",
				Environment: environment,
				Timestamp: time.Now().
					Format(time.RFC3339),
			},
		)
	}
}
