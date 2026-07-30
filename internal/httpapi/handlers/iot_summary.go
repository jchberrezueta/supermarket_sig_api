package handlers

import (
	"net/http"
	"strings"
)

// Summary devuelve los indicadores gerenciales de cadena de frío.
func (handler *IoTHandler) Summary(
	w http.ResponseWriter,
	r *http.Request,
) {
	deviceCode := strings.TrimSpace(
		r.URL.Query().Get(
			"codigoDispositivo",
		),
	)

	summary, err :=
		handler.service.Summary(
			r.Context(),
			deviceCode,
		)

	if err != nil {
		handler.writeServiceError(
			w,
			err,
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
