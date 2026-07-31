package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"supermarket-sig-api/internal/management"
)

func TestInventoryMovements(
	t *testing.T,
) {
	router := prepareInventoryRouter(
		t,
		validSnapshotJSON,
	)

	request := httptest.NewRequest(
		http.MethodGet,
		"/api/sig/inventario/movimientos?desde=2026-07-01&hasta=2026-07-31&limite=20",
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
		Success bool                               `json:"success"`
		Data    management.InventoryMovementReport `json:"data"`
	}

	if err := json.NewDecoder(
		response.Body,
	).Decode(&body); err != nil {
		t.Fatalf(
			"no se pudo leer los movimientos: %v",
			err,
		)
	}

	if !body.Success {
		t.Fatal(
			"se esperaba success=true",
		)
	}

	if body.Data.Total != 1 {
		t.Fatalf(
			"se esperaba un movimiento y se obtuvieron %d",
			body.Data.Total,
		)
	}

	if body.Data.Summary.TotalMovements != 1 {
		t.Fatalf(
			"resumen de movimientos inesperado: %d",
			body.Data.Summary.TotalMovements,
		)
	}

	if body.Data.Summary.Outputs != 1 {
		t.Fatalf(
			"se esperaba una salida y se obtuvieron %d",
			body.Data.Summary.Outputs,
		)
	}

	if body.Data.Summary.OutputUnits != 1 {
		t.Fatalf(
			"unidades de salida inesperadas: %d",
			body.Data.Summary.OutputUnits,
		)
	}

	if body.Data.Summary.NetUnits != -1 {
		t.Fatalf(
			"unidades netas inesperadas: %d",
			body.Data.Summary.NetUnits,
		)
	}

	item := body.Data.Items[0]

	if item.Product !=
		"Producto refrigerado" {
		t.Fatalf(
			"producto inesperado: %q",
			item.Product,
		)
	}

	if item.Type !=
		"salida_venta" {
		t.Fatalf(
			"tipo inesperado: %q",
			item.Type,
		)
	}

	if item.Quantity != -1 {
		t.Fatalf(
			"cantidad inesperada: %d",
			item.Quantity,
		)
	}

	if item.ProductStockBefore == nil ||
		*item.ProductStockBefore != 10 {
		t.Fatal(
			"stock anterior del producto inesperado",
		)
	}

	if item.ProductStockAfter == nil ||
		*item.ProductStockAfter != 9 {
		t.Fatal(
			"stock posterior del producto inesperado",
		)
	}
}

func TestInventoryMovementsTypeFilter(
	t *testing.T,
) {
	router := prepareInventoryRouter(
		t,
		validSnapshotJSON,
	)

	request := httptest.NewRequest(
		http.MethodGet,
		"/api/sig/inventario/movimientos?tipo=entrada_compra&limite=20",
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
		Data management.InventoryMovementReport `json:"data"`
	}

	if err := json.NewDecoder(
		response.Body,
	).Decode(&body); err != nil {
		t.Fatalf(
			"no se pudo leer la respuesta: %v",
			err,
		)
	}

	if body.Data.Total != 0 {
		t.Fatalf(
			"se esperaban cero movimientos y se obtuvieron %d",
			body.Data.Total,
		)
	}

	if len(body.Data.Items) != 0 {
		t.Fatalf(
			"se esperaba un arreglo vacío y se obtuvieron %d elementos",
			len(body.Data.Items),
		)
	}
}

func TestInventoryMovementsPeriodValidation(
	t *testing.T,
) {
	router := NewRouter(
		integrationTestConfig(),
		nil,
	)

	request := httptest.NewRequest(
		http.MethodGet,
		"/api/sig/inventario/movimientos?desde=fecha-invalida",
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
