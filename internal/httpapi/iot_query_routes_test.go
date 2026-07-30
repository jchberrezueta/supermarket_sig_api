package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func registerOutOfRangeReading(
	t *testing.T,
	router http.Handler,
) {
	t.Helper()

	request := httptest.NewRequest(
		http.MethodPost,
		"/api/sig/iot/lecturas",
		strings.NewReader(
			`{
				"codigoDispositivo": "ESP32-BODEGA-01",
				"temperatura": 12.5,
				"humedad": 70
			}`,
		),
	)

	request.Header.Set(
		"Content-Type",
		"application/json",
	)

	request.Header.Set(
		"X-Device-Key",
		"test-device-key",
	)

	response := httptest.NewRecorder()

	router.ServeHTTP(
		response,
		request,
	)

	if response.Code != http.StatusCreated {
		t.Fatalf(
			"no se pudo registrar la lectura: %s",
			response.Body.String(),
		)
	}
}

func TestListIoTResources(
	t *testing.T,
) {
	router := NewRouter(
		iotTestConfig(),
		nil,
	)

	registerOutOfRangeReading(
		t,
		router,
	)

	paths := []string{
		"/api/sig/iot/lecturas",
		"/api/sig/iot/alertas?estado=abierta",
		"/api/sig/iot/incidentes?estado=abierto",
	}

	for _, path := range paths {
		request := httptest.NewRequest(
			http.MethodGet,
			path,
			nil,
		)

		response := httptest.NewRecorder()

		router.ServeHTTP(
			response,
			request,
		)

		if response.Code != http.StatusOK {
			t.Fatalf(
				"ruta %s: se esperaba %d y se obtuvo %d: %s",
				path,
				http.StatusOK,
				response.Code,
				response.Body.String(),
			)
		}

		var body struct {
			Success bool `json:"success"`
			Data    struct {
				Total int `json:"total"`
			} `json:"data"`
		}

		if err := json.NewDecoder(
			response.Body,
		).Decode(&body); err != nil {
			t.Fatalf(
				"ruta %s: respuesta inválida: %v",
				path,
				err,
			)
		}

		if !body.Success {
			t.Fatalf(
				"ruta %s: se esperaba success=true",
				path,
			)
		}

		if body.Data.Total != 1 {
			t.Fatalf(
				"ruta %s: se esperaba un registro y se obtuvo %d",
				path,
				body.Data.Total,
			)
		}
	}
}

func TestIncidentDetail(
	t *testing.T,
) {
	router := NewRouter(
		iotTestConfig(),
		nil,
	)

	registerOutOfRangeReading(
		t,
		router,
	)

	request := httptest.NewRequest(
		http.MethodGet,
		"/api/sig/iot/incidentes/1",
		nil,
	)

	response := httptest.NewRecorder()

	router.ServeHTTP(
		response,
		request,
	)

	if response.Code != http.StatusOK {
		t.Fatalf(
			"se esperaba %d y se obtuvo %d: %s",
			http.StatusOK,
			response.Code,
			response.Body.String(),
		)
	}

	var body struct {
		Success bool `json:"success"`
		Data    struct {
			Incident struct {
				ID int64 `json:"id"`
			} `json:"incidente"`

			Alert struct {
				ID int64 `json:"id"`
			} `json:"alerta"`
		} `json:"data"`
	}

	if err := json.NewDecoder(
		response.Body,
	).Decode(&body); err != nil {
		t.Fatalf(
			"no se pudo leer la respuesta: %v",
			err,
		)
	}

	if body.Data.Incident.ID != 1 {
		t.Fatalf(
			"incidente inesperado: %d",
			body.Data.Incident.ID,
		)
	}

	if body.Data.Alert.ID != 1 {
		t.Fatalf(
			"alerta inesperada: %d",
			body.Data.Alert.ID,
		)
	}
}

func TestInvalidIncidentStatus(
	t *testing.T,
) {
	router := NewRouter(
		iotTestConfig(),
		nil,
	)

	request := httptest.NewRequest(
		http.MethodGet,
		"/api/sig/iot/incidentes?estado=inventado",
		nil,
	)

	response := httptest.NewRecorder()

	router.ServeHTTP(
		response,
		request,
	)

	if response.Code != http.StatusBadRequest {
		t.Fatalf(
			"se esperaba %d y se obtuvo %d",
			http.StatusBadRequest,
			response.Code,
		)
	}
}
