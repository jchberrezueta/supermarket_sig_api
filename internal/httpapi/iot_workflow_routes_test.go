package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"supermarket-sig-api/internal/iot"
)

func performWorkflowRequest(
	t *testing.T,
	router http.Handler,
	method string,
	path string,
	body string,
	expectedStatus int,
) {
	t.Helper()

	var request *http.Request

	if body == "" {
		request = httptest.NewRequest(
			method,
			path,
			nil,
		)
	} else {
		request = httptest.NewRequest(
			method,
			path,
			strings.NewReader(body),
		)

		request.Header.Set(
			"Content-Type",
			"application/json",
		)
	}

	request.Header.Set(
		"Authorization",
		"Bearer "+
			signAdminTestToken(
				t,
				"secreto-pruebas-jwt",
				validAdminTestClaims(
					"padmin",
				),
			),
	)

	response := httptest.NewRecorder()

	router.ServeHTTP(
		response,
		request,
	)

	if response.Code != expectedStatus {
		t.Fatalf(
			"%s %s: se esperaba %d y se obtuvo %d: %s",
			method,
			path,
			expectedStatus,
			response.Code,
			response.Body.String(),
		)
	}
}

func TestCompleteIncidentWorkflow(
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

	performWorkflowRequest(
		t,
		router,
		http.MethodPatch,
		"/api/sig/iot/incidentes/1/reconocer",
		`{
			"responsable": "Gerente de bodega"
		}`,
		http.StatusOK,
	)

	performWorkflowRequest(
		t,
		router,
		http.MethodPost,
		"/api/sig/iot/incidentes/1/acciones",
		`{
			"descripcion": "Se revisó y ajustó el sistema de refrigeración.",
			"responsable": "Técnico de mantenimiento",
			"resultado": "La temperatura comenzó a descender."
		}`,
		http.StatusCreated,
	)

	performWorkflowRequest(
		t,
		router,
		http.MethodPatch,
		"/api/sig/iot/incidentes/1/resolver",
		"",
		http.StatusOK,
	)

	performWorkflowRequest(
		t,
		router,
		http.MethodPatch,
		"/api/sig/iot/incidentes/1/cerrar",
		"",
		http.StatusOK,
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
		Success bool               `json:"success"`
		Data    iot.IncidentDetail `json:"data"`
	}

	if err := json.NewDecoder(
		response.Body,
	).Decode(&body); err != nil {
		t.Fatalf(
			"no se pudo leer la respuesta: %v",
			err,
		)
	}

	if body.Data.Incident.Status != "cerrado" {
		t.Fatalf(
			"estado final inesperado: %q",
			body.Data.Incident.Status,
		)
	}

	if body.Data.Alert.Status != "cerrada" {
		t.Fatalf(
			"estado de alerta inesperado: %q",
			body.Data.Alert.Status,
		)
	}

	if len(body.Data.Actions) != 1 {
		t.Fatalf(
			"se esperaba una acción correctiva y se obtuvieron %d",
			len(body.Data.Actions),
		)
	}

	if body.Data.Incident.RecognizedAt == nil {
		t.Fatal(
			"falta la fecha de reconocimiento",
		)
	}

	if body.Data.Incident.ResolvedAt == nil {
		t.Fatal(
			"falta la fecha de resolución",
		)
	}

	if body.Data.Incident.ClosedAt == nil {
		t.Fatal(
			"falta la fecha de cierre",
		)
	}
}

func TestResolveOpenIncidentIsRejected(
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

	performWorkflowRequest(
		t,
		router,
		http.MethodPatch,
		"/api/sig/iot/incidentes/1/resolver",
		"",
		http.StatusConflict,
	)
}
