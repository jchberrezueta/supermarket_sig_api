package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"supermarket-sig-api/internal/management"
)

func TestExecutiveSummaryWithoutERPData(
	t *testing.T,
) {
	router := NewRouter(
		integrationTestConfig(),
		nil,
	)

	request := httptest.NewRequest(
		http.MethodGet,
		"/api/sig/resumen-ejecutivo",
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
		Success bool                        `json:"success"`
		Data    management.ExecutiveSummary `json:"data"`
	}

	if err := json.NewDecoder(
		response.Body,
	).Decode(&body); err != nil {
		t.Fatalf(
			"no se pudo leer la respuesta: %v",
			err,
		)
	}

	if body.Data.HasERPData {
		t.Fatal(
			"se esperaba tieneDatosERP=false",
		)
	}

	if len(body.Data.Recommendations) == 0 {
		t.Fatal(
			"se esperaba al menos una recomendación",
		)
	}
}

func TestExecutiveSummaryWithERPAndIoTData(
	t *testing.T,
) {
	router := NewRouter(
		integrationTestConfig(),
		nil,
	)

	importRequest := httptest.NewRequest(
		http.MethodPost,
		"/api/sig/integracion/snapshot",
		strings.NewReader(
			validSnapshotJSON,
		),
	)

	importRequest.Header.Set(
		"Content-Type",
		"application/json",
	)

	importRequest.Header.Set(
		"X-SIG-Key",
		"test-sync-key",
	)

	importResponse := httptest.NewRecorder()

	router.ServeHTTP(
		importResponse,
		importRequest,
	)

	if importResponse.Code != http.StatusCreated {
		t.Fatalf(
			"no se pudo importar el snapshot: %s",
			importResponse.Body.String(),
		)
	}

	registerOutOfRangeReading(
		t,
		router,
	)

	summaryRequest := httptest.NewRequest(
		http.MethodGet,
		"/api/sig/resumen-ejecutivo",
		nil,
	)

	summaryResponse := httptest.NewRecorder()

	router.ServeHTTP(
		summaryResponse,
		summaryRequest,
	)

	if summaryResponse.Code != http.StatusOK {
		t.Fatalf(
			"se esperaba %d y se obtuvo %d: %s",
			http.StatusOK,
			summaryResponse.Code,
			summaryResponse.Body.String(),
		)
	}

	var body struct {
		Success bool                        `json:"success"`
		Data    management.ExecutiveSummary `json:"data"`
	}

	if err := json.NewDecoder(
		summaryResponse.Body,
	).Decode(&body); err != nil {
		t.Fatalf(
			"no se pudo leer el resumen: %v",
			err,
		)
	}

	if !body.Data.HasERPData {
		t.Fatal(
			"se esperaba tieneDatosERP=true",
		)
	}

	if body.Data.Sales.CompletedSales != 1 {
		t.Fatalf(
			"se esperaba una venta completada y se obtuvieron %d",
			body.Data.Sales.CompletedSales,
		)
	}

	if body.Data.Sales.Total != 2.5 {
		t.Fatalf(
			"total de ventas inesperado: %.2f",
			body.Data.Sales.Total,
		)
	}

	if body.Data.Sales.AverageTicket != 2.5 {
		t.Fatalf(
			"ticket promedio inesperado: %.2f",
			body.Data.Sales.AverageTicket,
		)
	}

	if body.Data.Sales.UnitsSold != 1 {
		t.Fatalf(
			"unidades vendidas inesperadas: %d",
			body.Data.Sales.UnitsSold,
		)
	}

	if body.Data.Sales.POS.Transactions != 1 {
		t.Fatalf(
			"transacciones POS inesperadas: %d",
			body.Data.Sales.POS.Transactions,
		)
	}

	if body.Data.Inventory.Products != 1 {
		t.Fatalf(
			"productos inesperados: %d",
			body.Data.Inventory.Products,
		)
	}

	if body.Data.Quality.OpenAlerts != 1 {
		t.Fatalf(
			"alertas abiertas inesperadas: %d",
			body.Data.Quality.OpenAlerts,
		)
	}

	if body.Data.Quality.ActiveIncidents != 1 {
		t.Fatalf(
			"incidentes activos inesperados: %d",
			body.Data.Quality.ActiveIncidents,
		)
	}

	if body.Data.LastSynchronization == nil {
		t.Fatal(
			"se esperaba información de sincronización",
		)
	}
}