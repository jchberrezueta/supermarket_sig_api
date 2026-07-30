package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"supermarket-sig-api/internal/management"
)

func prepareInventoryRouter(
	t *testing.T,
	snapshotJSON string,
) http.Handler {
	t.Helper()

	router := NewRouter(
		integrationTestConfig(),
		nil,
	)

	request := httptest.NewRequest(
		http.MethodPost,
		"/api/sig/integracion/snapshot",
		strings.NewReader(
			snapshotJSON,
		),
	)

	request.Header.Set(
		"Content-Type",
		"application/json",
	)

	request.Header.Set(
		"X-SIG-Key",
		"test-sync-key",
	)

	response := httptest.NewRecorder()

	router.ServeHTTP(
		response,
		request,
	)

	if response.Code != http.StatusCreated {
		t.Fatalf(
			"no se pudo preparar el snapshot: %s",
			response.Body.String(),
		)
	}

	return router
}

func TestInventoryOverview(
	t *testing.T,
) {
	router := prepareInventoryRouter(
		t,
		validSnapshotJSON,
	)

	request := httptest.NewRequest(
		http.MethodGet,
		"/api/sig/inventario/resumen?dias=30",
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
		Success bool                         `json:"success"`
		Data    management.InventoryOverview `json:"data"`
	}

	if err := json.NewDecoder(
		response.Body,
	).Decode(&body); err != nil {
		t.Fatalf(
			"no se pudo leer el resumen: %v",
			err,
		)
	}

	if body.Data.Products != 1 {
		t.Fatalf(
			"se esperaba un producto y se obtuvieron %d",
			body.Data.Products,
		)
	}

	if body.Data.AvailableUnits != 9 {
		t.Fatalf(
			"se esperaban 9 unidades y se obtuvieron %d",
			body.Data.AvailableUnits,
		)
	}

	if body.Data.OutOfStock != 0 {
		t.Fatalf(
			"productos agotados inesperados: %d",
			body.Data.OutOfStock,
		)
	}
}

func TestCriticalStock(
	t *testing.T,
) {
	snapshotJSON := strings.Replace(
		validSnapshotJSON,
		`"stockActual": 9`,
		`"stockActual": 3`,
		1,
	)

	router := prepareInventoryRouter(
		t,
		snapshotJSON,
	)

	request := httptest.NewRequest(
		http.MethodGet,
		"/api/sig/inventario/stock-critico?limite=10",
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
		Data management.CriticalStockReport `json:"data"`
	}

	if err := json.NewDecoder(
		response.Body,
	).Decode(&body); err != nil {
		t.Fatalf(
			"no se pudo leer el stock crítico: %v",
			err,
		)
	}

	if body.Data.Total != 1 {
		t.Fatalf(
			"se esperaba un producto crítico y se obtuvieron %d",
			body.Data.Total,
		)
	}

	if body.Data.Items[0].Status !=
		"stock_bajo" {
		t.Fatalf(
			"estado inesperado: %q",
			body.Data.Items[0].Status,
		)
	}

	if body.Data.Items[0].CurrentStock != 3 {
		t.Fatalf(
			"stock inesperado: %d",
			body.Data.Items[0].CurrentStock,
		)
	}
}

func TestExpiringLots(
	t *testing.T,
) {
	expirationDate :=
		time.Now().
			AddDate(
				0,
				0,
				10,
			).
			Format(
				time.RFC3339,
			)

	snapshotJSON := strings.Replace(
		validSnapshotJSON,
		"2026-08-30T00:00:00-05:00",
		expirationDate,
		1,
	)

	router := prepareInventoryRouter(
		t,
		snapshotJSON,
	)

	request := httptest.NewRequest(
		http.MethodGet,
		"/api/sig/inventario/caducidad?dias=30&limite=20",
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
		Data management.ExpirationReport `json:"data"`
	}

	if err := json.NewDecoder(
		response.Body,
	).Decode(&body); err != nil {
		t.Fatalf(
			"no se pudo leer la caducidad: %v",
			err,
		)
	}

	if body.Data.Total != 1 {
		t.Fatalf(
			"se esperaba un lote y se obtuvieron %d",
			body.Data.Total,
		)
	}

	if body.Data.Items[0].Status !=
		"proximo_caducar" {
		t.Fatalf(
			"estado inesperado: %q",
			body.Data.Items[0].Status,
		)
	}

	if body.Data.Items[0].Product !=
		"Producto refrigerado" {
		t.Fatalf(
			"producto inesperado: %q",
			body.Data.Items[0].Product,
		)
	}
}

func TestInventoryParameterValidation(
	t *testing.T,
) {
	router := NewRouter(
		integrationTestConfig(),
		nil,
	)

	request := httptest.NewRequest(
		http.MethodGet,
		"/api/sig/inventario/caducidad?dias=500",
		nil,
	)

	response := httptest.NewRecorder()

	router.ServeHTTP(
		response,
		request,
	)

	if response.Code !=
		http.StatusBadRequest {
		t.Fatalf(
			"se esperaba %d y se obtuvo %d: %s",
			http.StatusBadRequest,
			response.Code,
			response.Body.String(),
		)
	}
}
