package handlers

import (
	"encoding/json"
	"log"
	"net/http"
)

// WriteJSON escribe una respuesta HTTP en formato JSON.
func WriteJSON(
	w http.ResponseWriter,
	status int,
	payload any,
) {
	w.Header().Set(
		"Content-Type",
		"application/json; charset=utf-8",
	)

	w.WriteHeader(status)

	if status == http.StatusNoContent {
		return
	}

	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf(
			"no se pudo escribir la respuesta JSON: %v",
			err,
		)
	}
}
