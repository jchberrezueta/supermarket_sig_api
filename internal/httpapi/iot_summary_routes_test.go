package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"supermarket-sig-api/internal/iot"
)

func TestIoTSummaryWithoutReadings(
	t *testing.T,
) {
	router := NewRouter(
		iotTestConfig(),
		nil,
	)

	request := httptest.NewRequest(
		http.MethodGet,
		"/api/sig/iot/resumen",
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
		Success bool           `json:"success"`
		Data    iot.IoTSummary `json:"data"`
	}

	if err := json.NewDecoder(
		response.Body,
	).Decode(&body); err != nil {
		t.Fatalf(
			"no se pudo leer el resumen: %v",
			err,
		)
	}

	if body.Data.TotalReadings != 0 {
		t.Fatalf(
			"se esperaban cero lecturas y se obtuvieron %d",
			body.Data.TotalReadings,
		)
	}

	if len(body.Data.Recommendations) != 1 {
		t.Fatalf(
			"se esperaba una recomendación y se obtuvieron %d",
			len(body.Data.Recommendations),
		)
	}

	if body.Data.Recommendations[0].Code !=
		"IOT_SIN_LECTURAS" {
		t.Fatalf(
			"recomendación inesperada: %q",
			body.Data.Recommendations[0].Code,
		)
	}
}

func TestIoTSummaryWithOutOfRangeReading(
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
		"/api/sig/iot/resumen",
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
		Success bool           `json:"success"`
		Data    iot.IoTSummary `json:"data"`
	}

	if err := json.NewDecoder(
		response.Body,
	).Decode(&body); err != nil {
		t.Fatalf(
			"no se pudo leer el resumen: %v",
			err,
		)
	}

	if body.Data.TotalReadings != 1 {
		t.Fatalf(
			"se esperaba una lectura y se obtuvieron %d",
			body.Data.TotalReadings,
		)
	}

	if body.Data.OutOfRangeReadings != 1 {
		t.Fatalf(
			"se esperaba una lectura fuera de rango y se obtuvieron %d",
			body.Data.OutOfRangeReadings,
		)
	}

	if body.Data.Alerts.Open != 1 {
		t.Fatalf(
			"se esperaba una alerta abierta y se obtuvieron %d",
			body.Data.Alerts.Open,
		)
	}

	if body.Data.Incidents.Open != 1 {
		t.Fatalf(
			"se esperaba un incidente abierto y se obtuvieron %d",
			body.Data.Incidents.Open,
		)
	}

	if body.Data.LatestReading == nil {
		t.Fatal(
			"se esperaba la última lectura",
		)
	}

	if body.Data.LatestReading.Status !=
		iot.ReadingStatusOutOfRange {
		t.Fatalf(
			"estado inesperado: %q",
			body.Data.LatestReading.Status,
		)
	}
}
