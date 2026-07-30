package handlers

import (
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"

	"supermarket-sig-api/internal/iot"

	"github.com/go-chi/chi/v5"
)

type listData struct {
	Items any `json:"items"`
	Total int `json:"total"`
	Limit int `json:"limite"`
}

// ListReadings devuelve el historial de lecturas.
func (handler *IoTHandler) ListReadings(
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

	deviceCode := strings.TrimSpace(
		r.URL.Query().Get(
			"codigoDispositivo",
		),
	)

	readings, err :=
		handler.service.ListReadings(
			r.Context(),
			deviceCode,
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
				Items: readings,
				Total: len(readings),
				Limit: limit,
			},
		},
	)
}

// ListAlerts devuelve alertas filtradas por estado.
func (handler *IoTHandler) ListAlerts(
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

	status := strings.TrimSpace(
		r.URL.Query().Get(
			"estado",
		),
	)

	alerts, err :=
		handler.service.ListAlerts(
			r.Context(),
			status,
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
				Items: alerts,
				Total: len(alerts),
				Limit: limit,
			},
		},
	)
}

// ListIncidents devuelve incidentes filtrados por estado.
func (handler *IoTHandler) ListIncidents(
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

	status := strings.TrimSpace(
		r.URL.Query().Get(
			"estado",
		),
	)

	incidents, err :=
		handler.service.ListIncidents(
			r.Context(),
			status,
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
				Items: incidents,
				Total: len(incidents),
				Limit: limit,
			},
		},
	)
}

// IncidentDetail devuelve el detalle de un incidente.
func (handler *IoTHandler) IncidentDetail(
	w http.ResponseWriter,
	r *http.Request,
) {
	incidentID, err := strconv.ParseInt(
		chi.URLParam(
			r,
			"incidentID",
		),
		10,
		64,
	)

	if err != nil || incidentID <= 0 {
		WriteError(
			w,
			http.StatusBadRequest,
			"invalid_incident_id",
			"El identificador del incidente no es válido.",
		)

		return
	}

	detail, err :=
		handler.service.IncidentDetail(
			r.Context(),
			incidentID,
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
			Data:    detail,
		},
	)
}

func parseLimit(
	w http.ResponseWriter,
	r *http.Request,
) (
	int,
	bool,
) {
	value := strings.TrimSpace(
		r.URL.Query().Get(
			"limite",
		),
	)

	if value == "" {
		return 50, true
	}

	limit, err := strconv.Atoi(value)

	if err != nil ||
		limit < 1 ||
		limit > 200 {
		WriteError(
			w,
			http.StatusBadRequest,
			"invalid_limit",
			"El límite debe ser un número entre 1 y 200.",
		)

		return 0, false
	}

	return limit, true
}

func (handler *IoTHandler) writeServiceError(
	w http.ResponseWriter,
	err error,
) {
	var validationError *iot.ValidationError

	if errors.As(
		err,
		&validationError,
	) {
		WriteError(
			w,
			http.StatusBadRequest,
			"validation_error",
			validationError.Message,
		)

		return
	}

	var notFoundError *iot.NotFoundError

	if errors.As(
		err,
		&notFoundError,
	) {
		WriteError(
			w,
			http.StatusNotFound,
			"resource_not_found",
			notFoundError.Message,
		)

		return
	}

	log.Printf(
		"error del servicio IoT: %v",
		err,
	)

	WriteError(
		w,
		http.StatusInternalServerError,
		"iot_operation_failed",
		"No se pudo completar la operación IoT.",
	)
}
