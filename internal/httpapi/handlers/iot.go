package handlers

import (
	"crypto/subtle"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"supermarket-sig-api/internal/iot"
)

// IoTHandler gestiona los endpoints del dispositivo.
type IoTHandler struct {
	service           *iot.Service
	deviceKey         string
	defaultDeviceCode string
}

// NewIoTHandler crea el controlador HTTP del proceso IoT.
func NewIoTHandler(
	service *iot.Service,
	deviceKey string,
) *IoTHandler {
	return &IoTHandler{
		service: service,
		deviceKey: strings.TrimSpace(
			deviceKey,
		),
		defaultDeviceCode: service.DefaultDeviceCode(),
	}
}

type registerReadingRequest struct {
	DeviceCode  string   `json:"codigoDispositivo"`
	Temperature float64  `json:"temperatura"`
	Humidity    *float64 `json:"humedad,omitempty"`
	ReadAt      string   `json:"fechaLectura,omitempty"`
}

type successResponse struct {
	Success bool `json:"success"`
	Data    any  `json:"data"`
}

// RegisterReading recibe y procesa una lectura del ESP32.
func (handler *IoTHandler) RegisterReading(
	w http.ResponseWriter,
	r *http.Request,
) {
	if !handler.authorized(r) {
		WriteError(
			w,
			http.StatusUnauthorized,
			"invalid_device_key",
			"La clave del dispositivo no es válida.",
		)

		return
	}

	r.Body = http.MaxBytesReader(
		w,
		r.Body,
		1024*1024,
	)

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	var request registerReadingRequest

	if err := decoder.Decode(&request); err != nil {
		WriteError(
			w,
			http.StatusBadRequest,
			"invalid_json",
			"El cuerpo JSON no es válido.",
		)

		return
	}

	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		WriteError(
			w,
			http.StatusBadRequest,
			"invalid_json",
			"El cuerpo debe contener un único objeto JSON.",
		)

		return
	}

	var readAt time.Time

	if strings.TrimSpace(request.ReadAt) != "" {
		parsedDate, err := time.Parse(
			time.RFC3339,
			request.ReadAt,
		)

		if err != nil {
			WriteError(
				w,
				http.StatusBadRequest,
				"invalid_reading_date",
				"La fecha de lectura debe utilizar formato RFC3339.",
			)

			return
		}

		readAt = parsedDate
	}

	result, err := handler.service.Register(
		r.Context(),
		iot.RegisterInput{
			DeviceCode:  request.DeviceCode,
			Temperature: request.Temperature,
			Humidity:    request.Humidity,
			ReadAt:      readAt,
		},
	)

	if err != nil {
		var validationError *iot.ValidationError

		if errors.As(
			err,
			&validationError,
		) {
			WriteError(
				w,
				http.StatusBadRequest,
				"invalid_reading",
				validationError.Message,
			)

			return
		}

		log.Printf(
			"no se pudo registrar la lectura IoT: %v",
			err,
		)

		WriteError(
			w,
			http.StatusInternalServerError,
			"iot_registration_failed",
			"No se pudo registrar la lectura.",
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

// LatestReading devuelve la última lectura del dispositivo.
func (handler *IoTHandler) LatestReading(
	w http.ResponseWriter,
	r *http.Request,
) {
	deviceCode := strings.TrimSpace(
		r.URL.Query().Get(
			"codigoDispositivo",
		),
	)

	if deviceCode == "" {
		deviceCode =
			handler.defaultDeviceCode
	}

	reading, exists, err :=
		handler.service.Latest(
			r.Context(),
			deviceCode,
		)

	if err != nil {
		var validationError *iot.ValidationError

		if errors.As(
			err,
			&validationError,
		) {
			WriteError(
				w,
				http.StatusBadRequest,
				"invalid_device",
				validationError.Message,
			)

			return
		}

		log.Printf(
			"no se pudo consultar la última lectura: %v",
			err,
		)

		WriteError(
			w,
			http.StatusInternalServerError,
			"iot_query_failed",
			"No se pudo consultar la última lectura.",
		)

		return
	}

	if !exists {
		WriteError(
			w,
			http.StatusNotFound,
			"reading_not_found",
			"El dispositivo todavía no tiene lecturas.",
		)

		return
	}

	WriteJSON(
		w,
		http.StatusOK,
		successResponse{
			Success: true,
			Data:    reading,
		},
	)
}

func (handler *IoTHandler) authorized(
	r *http.Request,
) bool {
	if handler.deviceKey == "" {
		return false
	}

	providedKey := strings.TrimSpace(
		r.Header.Get(
			"X-Device-Key",
		),
	)

	if providedKey == "" {
		return false
	}

	return subtle.ConstantTimeCompare(
		[]byte(providedKey),
		[]byte(handler.deviceKey),
	) == 1
}
