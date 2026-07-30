package handlers

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"time"
)

// DatabaseHealthResponse representa el estado de Oracle.
type DatabaseHealthResponse struct {
	Status    string `json:"status"`
	Database  string `json:"database"`
	Enabled   bool   `json:"enabled"`
	User      string `json:"user,omitempty"`
	Service   string `json:"service,omitempty"`
	Timestamp string `json:"timestamp"`
}

// DatabaseHealth comprueba la disponibilidad de Oracle.
func DatabaseHealth(
	db *sql.DB,
	enabled bool,
) http.HandlerFunc {
	return func(
		w http.ResponseWriter,
		_ *http.Request,
	) {
		if !enabled {
			WriteJSON(
				w,
				http.StatusOK,
				DatabaseHealthResponse{
					Status:   "disabled",
					Database: "oracle",
					Enabled:  false,
					Timestamp: time.Now().
						Format(time.RFC3339),
				},
			)

			return
		}

		if db == nil {
			WriteError(
				w,
				http.StatusServiceUnavailable,
				"database_unavailable",
				"La conexión con Oracle no está disponible.",
			)

			return
		}

		ctx, cancel := context.WithTimeout(
			context.Background(),
			3*time.Second,
		)

		defer cancel()

		var user string
		var service string

		err := db.QueryRowContext(
			ctx,
			`
				SELECT
					USER,
					SYS_CONTEXT(
						'USERENV',
						'SERVICE_NAME'
					)
				FROM dual
			`,
		).Scan(
			&user,
			&service,
		)

		if err != nil {
			log.Printf(
				"falló la comprobación de Oracle: %v",
				err,
			)

			WriteError(
				w,
				http.StatusServiceUnavailable,
				"database_unavailable",
				"Oracle no respondió correctamente.",
			)

			return
		}

		WriteJSON(
			w,
			http.StatusOK,
			DatabaseHealthResponse{
				Status:   "ok",
				Database: "oracle",
				Enabled:  true,
				User:     user,
				Service:  service,
				Timestamp: time.Now().
					Format(time.RFC3339),
			},
		)
	}
}
