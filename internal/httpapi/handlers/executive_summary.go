package handlers

import (
	"log"
	"net/http"

	"supermarket-sig-api/internal/management"
)

// ManagementHandler gestiona las consultas gerenciales.
type ManagementHandler struct {
	service *management.Service
}

// NewManagementHandler crea el controlador del SIG.
func NewManagementHandler(
	service *management.Service,
) *ManagementHandler {
	return &ManagementHandler{
		service: service,
	}
}

// ExecutiveSummary devuelve los indicadores principales.
func (handler *ManagementHandler) ExecutiveSummary(
	w http.ResponseWriter,
	r *http.Request,
) {
	summary, err :=
		handler.service.ExecutiveSummary(
			r.Context(),
		)

	if err != nil {
		log.Printf(
			"no se pudo calcular el resumen ejecutivo: %v",
			err,
		)

		WriteError(
			w,
			http.StatusInternalServerError,
			"executive_summary_failed",
			"No se pudo calcular el resumen ejecutivo.",
		)

		return
	}

	WriteJSON(
		w,
		http.StatusOK,
		successResponse{
			Success: true,
			Data:    summary,
		},
	)
}
