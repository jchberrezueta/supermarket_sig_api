package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"supermarket-sig-api/internal/management"
)

func prepareSalesRouter(
	t *testing.T,
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
			validSnapshotJSON,
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

func TestSalesOverview(
	t *testing.T,
) {
	router := prepareSalesRouter(
		t,
	)

	request := httptest.NewRequest(
		http.MethodGet,
		"/api/sig/ventas/resumen?desde=2026-07-01&hasta=2026-07-31",
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
		Success bool                     `json:"success"`
		Data    management.SalesOverview `json:"data"`
	}

	if err := json.NewDecoder(
		response.Body,
	).Decode(&body); err != nil {
		t.Fatalf(
			"no se pudo leer la respuesta: %v",
			err,
		)
	}

	if body.Data.Indicators.CompletedSales != 1 {
		t.Fatalf(
			"se esperaba una venta y se obtuvieron %d",
			body.Data.Indicators.CompletedSales,
		)
	}

	if body.Data.Indicators.Total != 2.5 {
		t.Fatalf(
			"total inesperado: %.2f",
			body.Data.Indicators.Total,
		)
	}

	if body.Data.Indicators.AverageTicket != 2.5 {
		t.Fatalf(
			"ticket promedio inesperado: %.2f",
			body.Data.Indicators.AverageTicket,
		)
	}

	if body.Data.Indicators.UnitsSold != 1 {
		t.Fatalf(
			"unidades inesperadas: %d",
			body.Data.Indicators.UnitsSold,
		)
	}
}

func TestSalesTrend(
	t *testing.T,
) {
	router := prepareSalesRouter(
		t,
	)

	request := httptest.NewRequest(
		http.MethodGet,
		"/api/sig/ventas/tendencia",
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
		Data management.SalesTrend `json:"data"`
	}

	if err := json.NewDecoder(
		response.Body,
	).Decode(&body); err != nil {
		t.Fatalf(
			"no se pudo leer la tendencia: %v",
			err,
		)
	}

	if body.Data.Total != 1 {
		t.Fatalf(
			"se esperaba un punto y se obtuvieron %d",
			body.Data.Total,
		)
	}

	if body.Data.Items[0].Date != "2026-07-30" {
		t.Fatalf(
			"fecha inesperada: %q",
			body.Data.Items[0].Date,
		)
	}

	if body.Data.Items[0].Total != 2.5 {
		t.Fatalf(
			"total diario inesperado: %.2f",
			body.Data.Items[0].Total,
		)
	}
}

func TestTopSellingProducts(
	t *testing.T,
) {
	router := prepareSalesRouter(
		t,
	)

	request := httptest.NewRequest(
		http.MethodGet,
		"/api/sig/ventas/productos?limite=10",
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
		Data management.ProductSalesRanking `json:"data"`
	}

	if err := json.NewDecoder(
		response.Body,
	).Decode(&body); err != nil {
		t.Fatalf(
			"no se pudo leer el ranking: %v",
			err,
		)
	}

	if body.Data.Total != 1 {
		t.Fatalf(
			"se esperaba un producto y se obtuvieron %d",
			body.Data.Total,
		)
	}

	if body.Data.Items[0].Name !=
		"Producto refrigerado" {
		t.Fatalf(
			"producto inesperado: %q",
			body.Data.Items[0].Name,
		)
	}

	if body.Data.Items[0].Units != 1 {
		t.Fatalf(
			"unidades inesperadas: %d",
			body.Data.Items[0].Units,
		)
	}

	if body.Data.Items[0].Revenue != 2.5 {
		t.Fatalf(
			"ingresos inesperados: %.2f",
			body.Data.Items[0].Revenue,
		)
	}
}

func TestSalesPeriodValidation(
	t *testing.T,
) {
	router := NewRouter(
		integrationTestConfig(),
		nil,
	)

	request := httptest.NewRequest(
		http.MethodGet,
		"/api/sig/ventas/resumen?desde=2026-08-01&hasta=2026-07-01",
		nil,
	)

	response := httptest.NewRecorder()

	router.ServeHTTP(
		response,
		request,
	)

	if response.Code != http.StatusBadRequest {
		t.Fatalf(
			"se esperaba %d y se obtuvo %d: %s",
			http.StatusBadRequest,
			response.Code,
			response.Body.String(),
		)
	}
}

func TestSalesByCategory(
	t *testing.T,
) {
	router := prepareSalesRouter(
		t,
	)

	request := httptest.NewRequest(
		http.MethodGet,
		"/api/sig/ventas/categorias?desde=2026-07-01&hasta=2026-07-31&limite=10",
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
		Success bool                            `json:"success"`
		Data    management.CategorySalesRanking `json:"data"`
	}

	if err := json.NewDecoder(
		response.Body,
	).Decode(&body); err != nil {
		t.Fatalf(
			"no se pudo leer el ranking por categoría: %v",
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
			"se esperaba una categoría y se obtuvieron %d",
			body.Data.Total,
		)
	}

	if len(body.Data.Items) != 1 {
		t.Fatalf(
			"se esperaba un elemento y se obtuvieron %d",
			len(body.Data.Items),
		)
	}

	item := body.Data.Items[0]

	if item.CategoryID != 1 {
		t.Fatalf(
			"categoría inesperada: %d",
			item.CategoryID,
		)
	}

	if item.Name != "Refrigerados" {
		t.Fatalf(
			"nombre de categoría inesperado: %q",
			item.Name,
		)
	}

	if item.Units != 1 {
		t.Fatalf(
			"unidades inesperadas: %d",
			item.Units,
		)
	}

	if item.Revenue != 2.5 {
		t.Fatalf(
			"ingresos inesperados: %.2f",
			item.Revenue,
		)
	}
}
