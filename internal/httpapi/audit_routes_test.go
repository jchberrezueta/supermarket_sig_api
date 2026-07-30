package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"supermarket-sig-api/internal/iot"
)

func TestAuditCreatedAutomatically(
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
		"/api/sig/auditoria",
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

		Data struct {
			Items []iot.AuditEvent `json:"items"`
			Total int              `json:"total"`
		} `json:"data"`
	}

	if err := json.NewDecoder(
		response.Body,
	).Decode(&body); err != nil {
		t.Fatalf(
			"no se pudo leer la auditoría: %v",
			err,
		)
	}

	if !body.Success {
		t.Fatal(
			"se esperaba success=true",
		)
	}

	if body.Data.Total != 3 {
		t.Fatalf(
			"se esperaban 3 eventos iniciales y se obtuvieron %d",
			body.Data.Total,
		)
	}
}

func TestAuditIncidentClosed(
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
			"descripcion": "Se ajustó el sistema de refrigeración.",
			"responsable": "Técnico de mantenimiento",
			"resultado": "Temperatura normalizada."
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
		"/api/sig/auditoria?accion=incidente_cerrado",
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
		Data struct {
			Items []iot.AuditEvent `json:"items"`
			Total int              `json:"total"`
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

	if body.Data.Total != 1 {
		t.Fatalf(
			"se esperaba un cierre auditado y se obtuvieron %d",
			body.Data.Total,
		)
	}

	if body.Data.Items[0].Action !=
		"incidente_cerrado" {
		t.Fatalf(
			"acción inesperada: %q",
			body.Data.Items[0].Action,
		)
	}
}

func TestEmptyActionsReturnsArray(
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
			"se esperaba %d y se obtuvo %d",
			http.StatusOK,
			response.Code,
		)
	}

	var body struct {
		Data iot.IncidentDetail `json:"data"`
	}

	if err := json.NewDecoder(
		response.Body,
	).Decode(&body); err != nil {
		t.Fatalf(
			"no se pudo leer la respuesta: %v",
			err,
		)
	}

	if body.Data.Actions == nil {
		t.Fatal(
			"acciones debe ser un arreglo vacío y no null",
		)
	}

	if len(body.Data.Actions) != 0 {
		t.Fatalf(
			"se esperaban cero acciones y se obtuvieron %d",
			len(body.Data.Actions),
		)
	}
}
