package handlers

import (
	"net/http"
	"strings"
)

// ListAudit devuelve el historial de auditoría del SIG.
func (handler *IoTHandler) ListAudit(
	w http.ResponseWriter,
	r *http.Request,
) {
	limit, ok := parseLimit(
		w,
		r,
	)

	if !ok {
		return
	}

	action := strings.TrimSpace(
		r.URL.Query().Get(
			"accion",
		),
	)

	events, err :=
		handler.service.ListAudit(
			r.Context(),
			action,
			limit,
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
			Data: listData{
				Items: events,
				Total: len(events),
				Limit: limit,
			},
		},
	)
}
