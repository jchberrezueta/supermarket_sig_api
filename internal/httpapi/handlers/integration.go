package handlers

import (
	"crypto/subtle"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strings"

	"supermarket-sig-api/internal/erpdata"
)

// IntegrationHandler gestiona la sincronización con el ERP.
type IntegrationHandler struct {
	service *erpdata.Service
	syncKey string
}

// NewIntegrationHandler crea el controlador de integración.
func NewIntegrationHandler(
	service *erpdata.Service,
	syncKey string,
) *IntegrationHandler {
	return &IntegrationHandler{
		service: service,
		syncKey: strings.TrimSpace(
			syncKey,
		),
	}
}

// ImportSnapshot importa una copia completa del ERP.
func (handler *IntegrationHandler) ImportSnapshot(
	w http.ResponseWriter,
	r *http.Request,
) {
	if !handler.authorized(r) {
		WriteError(
			w,
			http.StatusUnauthorized,
			"invalid_sync_key",
			"La clave de sincronización no es válida.",
		)

		return
	}

	r.Body = http.MaxBytesReader(
		w,
		r.Body,
		10*1024*1024,
	)

	decoder := json.NewDecoder(
		r.Body,
	)

	decoder.DisallowUnknownFields()

	var snapshot erpdata.Snapshot

	if err := decoder.Decode(
		&snapshot,
	); err != nil {
		WriteError(
			w,
			http.StatusBadRequest,
			"invalid_snapshot_json",
			"El snapshot empresarial no contiene un JSON válido.",
		)

		return
	}

	if err := decoder.Decode(
		&struct{}{},
	); err != io.EOF {
		WriteError(
			w,
			http.StatusBadRequest,
			"invalid_snapshot_json",
			"El cuerpo debe contener un único objeto JSON.",
		)

		return
	}

	result, err :=
		handler.service.ImportSnapshot(
			r.Context(),
			snapshot,
		)

	if err != nil {
		var validationError *erpdata.ValidationError

		if errors.As(
			err,
			&validationError,
		) {
			WriteError(
				w,
				http.StatusBadRequest,
				"invalid_snapshot",
				validationError.Message,
			)

			return
		}

		log.Printf(
			"no se pudo importar el snapshot: %v",
			err,
		)

		WriteError(
			w,
			http.StatusInternalServerError,
			"snapshot_import_failed",
			"No se pudo importar la información del ERP.",
		)

		return
	}

	WriteJSON(
		w,
		http.StatusCreated,
		successResponse{
			Success: true,
			Data:    result,
		},
	)
}

// State devuelve el estado de sincronización.
func (handler *IntegrationHandler) State(
	w http.ResponseWriter,
	r *http.Request,
) {
	state, err :=
		handler.service.State(
			r.Context(),
		)

	if err != nil {
		log.Printf(
			"no se pudo consultar la sincronización: %v",
			err,
		)

		WriteError(
			w,
			http.StatusInternalServerError,
			"integration_state_failed",
			"No se pudo consultar el estado de sincronización.",
		)

		return
	}

	WriteJSON(
		w,
		http.StatusOK,
		successResponse{
			Success: true,
			Data:    state,
		},
	)
}

func (handler *IntegrationHandler) authorized(
	r *http.Request,
) bool {
	if handler.syncKey == "" {
		return false
	}

	providedKey := strings.TrimSpace(
		r.Header.Get(
			"X-SIG-Key",
		),
	)

	if providedKey == "" {
		return false
	}

	return subtle.ConstantTimeCompare(
		[]byte(providedKey),
		[]byte(handler.syncKey),
	) == 1
}
