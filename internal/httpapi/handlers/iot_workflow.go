package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

type recognizeIncidentRequest struct {
	Responsible string `json:"responsable"`
}

type correctiveActionRequest struct {
	Description string `json:"descripcion"`
	Responsible string `json:"responsable"`
	Result      string `json:"resultado,omitempty"`
}

// RecognizeIncident reconoce un incidente abierto.
func (handler *IoTHandler) RecognizeIncident(
	w http.ResponseWriter,
	r *http.Request,
) {
	incidentID, ok := readIncidentID(
		w,
		r,
	)

	if !ok {
		return
	}

	var request recognizeIncidentRequest

	if !decodeWorkflowJSON(
		w,
		r,
		&request,
	) {
		return
	}

	detail, err :=
		handler.service.RecognizeIncident(
			r.Context(),
			incidentID,
			request.Responsible,
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

// AddCorrectiveAction registra una acción correctiva.
func (handler *IoTHandler) AddCorrectiveAction(
	w http.ResponseWriter,
	r *http.Request,
) {
	incidentID, ok := readIncidentID(
		w,
		r,
	)

	if !ok {
		return
	}

	var request correctiveActionRequest

	if !decodeWorkflowJSON(
		w,
		r,
		&request,
	) {
		return
	}

	detail, err :=
		handler.service.AddCorrectiveAction(
			r.Context(),
			incidentID,
			request.Description,
			request.Responsible,
			request.Result,
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
		http.StatusCreated,
		successResponse{
			Success: true,
			Data:    detail,
		},
	)
}

// ResolveIncident resuelve un incidente en tratamiento.
func (handler *IoTHandler) ResolveIncident(
	w http.ResponseWriter,
	r *http.Request,
) {
	incidentID, ok := readIncidentID(
		w,
		r,
	)

	if !ok {
		return
	}

	detail, err :=
		handler.service.ResolveIncident(
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

// CloseIncident cierra un incidente resuelto.
func (handler *IoTHandler) CloseIncident(
	w http.ResponseWriter,
	r *http.Request,
) {
	incidentID, ok := readIncidentID(
		w,
		r,
	)

	if !ok {
		return
	}

	detail, err :=
		handler.service.CloseIncident(
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

func readIncidentID(
	w http.ResponseWriter,
	r *http.Request,
) (
	int64,
	bool,
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

		return 0, false
	}

	return incidentID, true
}

func decodeWorkflowJSON(
	w http.ResponseWriter,
	r *http.Request,
	destination any,
) bool {
	r.Body = http.MaxBytesReader(
		w,
		r.Body,
		1024*1024,
	)

	decoder := json.NewDecoder(
		r.Body,
	)

	decoder.DisallowUnknownFields()

	if err := decoder.Decode(
		destination,
	); err != nil {
		WriteError(
			w,
			http.StatusBadRequest,
			"invalid_json",
			"El cuerpo JSON no es válido.",
		)

		return false
	}

	if err := decoder.Decode(
		&struct{}{},
	); err != io.EOF {
		WriteError(
			w,
			http.StatusBadRequest,
			"invalid_json",
			"El cuerpo debe contener un único objeto JSON.",
		)

		return false
	}

	return true
}
