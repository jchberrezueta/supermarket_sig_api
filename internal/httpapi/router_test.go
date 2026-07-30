package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"supermarket-sig-api/internal/config"
	"supermarket-sig-api/internal/httpapi/handlers"
)

func testConfig() config.Config {
	return config.Config{
		AppEnvironment: "test",
		Port:           "8080",
		CORSOrigins: []string{
			"http://localhost:4200",
		},
		Database: config.DatabaseConfig{
			Enabled: false,
		},
	}
}

func TestHealthEndpoint(
	t *testing.T,
) {
	request := httptest.NewRequest(
		http.MethodGet,
		"/api/sig/health",
		nil,
	)

	response := httptest.NewRecorder()

	NewRouter(
		testConfig(),
		nil,
	).ServeHTTP(
		response,
		request,
	)

	if response.Code != http.StatusOK {
		t.Fatalf(
			"se esperaba estado %d y se obtuvo %d",
			http.StatusOK,
			response.Code,
		)
	}

	var body struct {
		Status      string `json:"status"`
		Service     string `json:"service"`
		Environment string `json:"environment"`
	}

	if err := json.NewDecoder(
		response.Body,
	).Decode(&body); err != nil {
		t.Fatalf(
			"no se pudo leer la respuesta: %v",
			err,
		)
	}

	if body.Status != "ok" {
		t.Fatalf(
			"se esperaba status ok y se obtuvo %q",
			body.Status,
		)
	}

	if body.Service != "SuperMarket SIG API" {
		t.Fatalf(
			"nombre de servicio inesperado: %q",
			body.Service,
		)
	}

	if body.Environment != "test" {
		t.Fatalf(
			"entorno inesperado: %q",
			body.Environment,
		)
	}
}

func TestDatabaseHealthDisabled(
	t *testing.T,
) {
	request := httptest.NewRequest(
		http.MethodGet,
		"/api/sig/db-health",
		nil,
	)

	response := httptest.NewRecorder()

	NewRouter(
		testConfig(),
		nil,
	).ServeHTTP(
		response,
		request,
	)

	if response.Code != http.StatusOK {
		t.Fatalf(
			"se esperaba estado %d y se obtuvo %d",
			http.StatusOK,
			response.Code,
		)
	}

	var body handlers.DatabaseHealthResponse

	if err := json.NewDecoder(
		response.Body,
	).Decode(&body); err != nil {
		t.Fatalf(
			"no se pudo leer la respuesta: %v",
			err,
		)
	}

	if body.Status != "disabled" {
		t.Fatalf(
			"se esperaba status disabled y se obtuvo %q",
			body.Status,
		)
	}

	if body.Enabled {
		t.Fatal(
			"se esperaba enabled=false",
		)
	}

	if body.Database != "oracle" {
		t.Fatalf(
			"base de datos inesperada: %q",
			body.Database,
		)
	}
}

func TestRouteNotFound(
	t *testing.T,
) {
	request := httptest.NewRequest(
		http.MethodGet,
		"/api/sig/ruta-inexistente",
		nil,
	)

	response := httptest.NewRecorder()

	NewRouter(
		testConfig(),
		nil,
	).ServeHTTP(
		response,
		request,
	)

	if response.Code != http.StatusNotFound {
		t.Fatalf(
			"se esperaba estado %d y se obtuvo %d",
			http.StatusNotFound,
			response.Code,
		)
	}

	var body handlers.ErrorResponse

	if err := json.NewDecoder(
		response.Body,
	).Decode(&body); err != nil {
		t.Fatalf(
			"no se pudo leer la respuesta de error: %v",
			err,
		)
	}

	if body.Success {
		t.Fatal(
			"se esperaba success=false",
		)
	}

	if body.Error.Code != "route_not_found" {
		t.Fatalf(
			"código de error inesperado: %q",
			body.Error.Code,
		)
	}
}

func TestMethodNotAllowed(
	t *testing.T,
) {
	request := httptest.NewRequest(
		http.MethodPost,
		"/api/sig/health",
		nil,
	)

	response := httptest.NewRecorder()

	NewRouter(
		testConfig(),
		nil,
	).ServeHTTP(
		response,
		request,
	)

	if response.Code != http.StatusMethodNotAllowed {
		t.Fatalf(
			"se esperaba estado %d y se obtuvo %d",
			http.StatusMethodNotAllowed,
			response.Code,
		)
	}

	var body handlers.ErrorResponse

	if err := json.NewDecoder(
		response.Body,
	).Decode(&body); err != nil {
		t.Fatalf(
			"no se pudo leer la respuesta de error: %v",
			err,
		)
	}

	if body.Error.Code != "method_not_allowed" {
		t.Fatalf(
			"código de error inesperado: %q",
			body.Error.Code,
		)
	}
}
