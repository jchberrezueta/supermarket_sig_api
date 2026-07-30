package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"supermarket-sig-api/internal/management"
)

func TestSupplyOverview(
	t *testing.T,
) {
	router := prepareInventoryRouter(
		t,
		validSnapshotJSON,
	)

	request := httptest.NewRequest(
		http.MethodGet,
		"/api/sig/abastecimiento/resumen",
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
		Success bool                      `json:"success"`
		Data    management.SupplyOverview `json:"data"`
	}

	if err := json.NewDecoder(
		response.Body,
	).Decode(&body); err != nil {
		t.Fatalf(
			"no se pudo leer el resumen: %v",
			err,
		)
	}

	if body.Data.TotalOrders != 1 {
		t.Fatalf(
			"se esperaba un pedido y se obtuvieron %d",
			body.Data.TotalOrders,
		)
	}

	if body.Data.CompletedOrders != 1 {
		t.Fatalf(
			"se esperaba un pedido completado y se obtuvieron %d",
			body.Data.CompletedOrders,
		)
	}

	if body.Data.RequestedQuantity != 10 {
		t.Fatalf(
			"cantidad solicitada inesperada: %d",
			body.Data.RequestedQuantity,
		)
	}

	if body.Data.ReceivedQuantity != 10 {
		t.Fatalf(
			"cantidad recibida inesperada: %d",
			body.Data.ReceivedQuantity,
		)
	}

	if body.Data.FulfillmentPercentage != 100 {
		t.Fatalf(
			"cumplimiento inesperado: %.2f",
			body.Data.FulfillmentPercentage,
		)
	}

	if body.Data.OnTimeDeliveries != 1 {
		t.Fatalf(
			"entregas puntuales inesperadas: %d",
			body.Data.OnTimeDeliveries,
		)
	}

	if body.Data.OnTimePercentage != 100 {
		t.Fatalf(
			"puntualidad inesperada: %.2f",
			body.Data.OnTimePercentage,
		)
	}
}

func TestSupplierPerformance(
	t *testing.T,
) {
	router := prepareInventoryRouter(
		t,
		validSnapshotJSON,
	)

	request := httptest.NewRequest(
		http.MethodGet,
		"/api/sig/abastecimiento/proveedores?limite=10",
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
		Data management.SupplierPerformanceReport `json:"data"`
	}

	if err := json.NewDecoder(
		response.Body,
	).Decode(&body); err != nil {
		t.Fatalf(
			"no se pudo leer el desempeño: %v",
			err,
		)
	}

	if body.Data.Total != 1 {
		t.Fatalf(
			"se esperaba una empresa y se obtuvieron %d",
			body.Data.Total,
		)
	}

	item := body.Data.Items[0]

	if item.Company !=
		"Distribuidora Académica" {
		t.Fatalf(
			"empresa inesperada: %q",
			item.Company,
		)
	}

	if item.Suppliers != 1 {
		t.Fatalf(
			"proveedores inesperados: %d",
			item.Suppliers,
		)
	}

	if item.FulfillmentPercentage != 100 {
		t.Fatalf(
			"cumplimiento inesperado: %.2f",
			item.FulfillmentPercentage,
		)
	}

	if item.OnTimePercentage != 100 {
		t.Fatalf(
			"puntualidad inesperada: %.2f",
			item.OnTimePercentage,
		)
	}
}

func TestSupplyOrders(
	t *testing.T,
) {
	router := prepareInventoryRouter(
		t,
		validSnapshotJSON,
	)

	request := httptest.NewRequest(
		http.MethodGet,
		"/api/sig/abastecimiento/pedidos?estado=completado&limite=20",
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
		Data management.SupplyOrderReport `json:"data"`
	}

	if err := json.NewDecoder(
		response.Body,
	).Decode(&body); err != nil {
		t.Fatalf(
			"no se pudo leer los pedidos: %v",
			err,
		)
	}

	if body.Data.Total != 1 {
		t.Fatalf(
			"se esperaba un pedido y se obtuvieron %d",
			body.Data.Total,
		)
	}

	if body.Data.Items[0].Company !=
		"Distribuidora Académica" {
		t.Fatalf(
			"empresa inesperada: %q",
			body.Data.Items[0].Company,
		)
	}

	if body.Data.Items[0].FulfillmentPercentage !=
		100 {
		t.Fatalf(
			"cumplimiento inesperado: %.2f",
			body.Data.Items[0].FulfillmentPercentage,
		)
	}
}

func TestSupplyLimitValidation(
	t *testing.T,
) {
	router := NewRouter(
		integrationTestConfig(),
		nil,
	)

	request := httptest.NewRequest(
		http.MethodGet,
		"/api/sig/abastecimiento/proveedores?limite=500",
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
