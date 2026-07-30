package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"supermarket-sig-api/internal/config"
	"supermarket-sig-api/internal/iot"
)

func iotTestConfig() config.Config {
	cfg := testConfig()

	cfg.IoT = config.IoTConfig{
		DeviceKey:      "test-device-key",
		DeviceCode:     "ESP32-BODEGA-01",
		TemperatureMin: 2,
		TemperatureMax: 8,
	}

	return cfg
}

func TestRegisterNormalReading(
	t *testing.T,
) {
	router := NewRouter(
		iotTestConfig(),
		nil,
	)

	request := httptest.NewRequest(
		http.MethodPost,
		"/api/sig/iot/lecturas",
		strings.NewReader(
			`{
				"codigoDispositivo": "ESP32-BODEGA-01",
				"temperatura": 4.5,
				"humedad": 65
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
			"se esperaba estado %d y se obtuvo %d: %s",
			http.StatusCreated,
			response.Code,
			response.Body.String(),
		)
	}

	var body struct {
		Success bool `json:"success"`
		Data    struct {
			Reading  iot.Reading   `json:"lectura"`
			Alert    *iot.Alert    `json:"alerta"`
			Incident *iot.Incident `json:"incidente"`
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

	if !body.Success {
		t.Fatal(
			"se esperaba success=true",
		)
	}

	if body.Data.Reading.Status !=
		iot.ReadingStatusNormal {
		t.Fatalf(
			"estado inesperado: %q",
			body.Data.Reading.Status,
		)
	}

	if body.Data.Alert != nil {
		t.Fatal(
			"una lectura normal no debe crear alerta",
		)
	}

	if body.Data.Incident != nil {
		t.Fatal(
			"una lectura normal no debe crear incidente",
		)
	}
}

func TestOutOfRangeReadingCreatesIncident(
	t *testing.T,
) {
	router := NewRouter(
		iotTestConfig(),
		nil,
	)

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
			"se esperaba estado %d y se obtuvo %d: %s",
			http.StatusCreated,
			response.Code,
			response.Body.String(),
		)
	}

	var body struct {
		Success bool `json:"success"`
		Data    struct {
			Reading  iot.Reading   `json:"lectura"`
			Alert    *iot.Alert    `json:"alerta"`
			Incident *iot.Incident `json:"incidente"`
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

	if body.Data.Reading.Status !=
		iot.ReadingStatusOutOfRange {
		t.Fatalf(
			"estado inesperado: %q",
			body.Data.Reading.Status,
		)
	}

	if body.Data.Alert == nil {
		t.Fatal(
			"se esperaba una alerta automática",
		)
	}

	if body.Data.Incident == nil {
		t.Fatal(
			"se esperaba un incidente automático",
		)
	}
}

func TestRegisterReadingRejectsInvalidDeviceKey(
	t *testing.T,
) {
	router := NewRouter(
		iotTestConfig(),
		nil,
	)

	request := httptest.NewRequest(
		http.MethodPost,
		"/api/sig/iot/lecturas",
		strings.NewReader(
			`{
				"codigoDispositivo": "ESP32-BODEGA-01",
				"temperatura": 4.5
			}`,
		),
	)

	request.Header.Set(
		"Content-Type",
		"application/json",
	)

	request.Header.Set(
		"X-Device-Key",
		"incorrecta",
	)

	response := httptest.NewRecorder()

	router.ServeHTTP(
		response,
		request,
	)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf(
			"se esperaba estado %d y se obtuvo %d",
			http.StatusUnauthorized,
			response.Code,
		)
	}
}

func TestLatestReading(
	t *testing.T,
) {
	router := NewRouter(
		iotTestConfig(),
		nil,
	)

	registerRequest := httptest.NewRequest(
		http.MethodPost,
		"/api/sig/iot/lecturas",
		strings.NewReader(
			`{
				"codigoDispositivo": "ESP32-BODEGA-01",
				"temperatura": 5.2,
				"humedad": 61
			}`,
		),
	)

	registerRequest.Header.Set(
		"Content-Type",
		"application/json",
	)

	registerRequest.Header.Set(
		"X-Device-Key",
		"test-device-key",
	)

	registerResponse := httptest.NewRecorder()

	router.ServeHTTP(
		registerResponse,
		registerRequest,
	)

	if registerResponse.Code != http.StatusCreated {
		t.Fatalf(
			"no se pudo preparar la lectura: %s",
			registerResponse.Body.String(),
		)
	}

	latestRequest := httptest.NewRequest(
		http.MethodGet,
		"/api/sig/iot/lecturas/ultima",
		nil,
	)

	latestResponse := httptest.NewRecorder()

	router.ServeHTTP(
		latestResponse,
		latestRequest,
	)

	if latestResponse.Code != http.StatusOK {
		t.Fatalf(
			"se esperaba estado %d y se obtuvo %d: %s",
			http.StatusOK,
			latestResponse.Code,
			latestResponse.Body.String(),
		)
	}
}
