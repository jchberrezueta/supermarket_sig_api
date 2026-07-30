package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"supermarket-sig-api/internal/config"
	"supermarket-sig-api/internal/erpdata"
)

const validSnapshotJSON = `{
	"fechaGeneracion": "2026-07-30T00:30:00-05:00",

	"categorias": [
		{
			"idOrigen": 1,
			"nombre": "Refrigerados",
			"estado": "activo"
		}
	],

	"empresas": [
		{
			"idOrigen": 1,
			"nombre": "Distribuidora Académica",
			"responsable": "Responsable de empresa",
			"telefono": "0999999999",
			"correo": "empresa@example.com",
			"estado": "activo"
		}
	],

	"proveedores": [
		{
			"idOrigen": 1,
			"idEmpresaOrigen": 1,
			"identificacion": "0999999999",
			"nombre": "Proveedor Académico",
			"telefono": "0999999999",
			"correo": "proveedor@example.com",
			"estado": "activo"
		}
	],

	"productos": [
		{
			"idOrigen": 1,
			"idCategoriaOrigen": 1,
			"codigoBarra": "7861000000001",
			"nombre": "Producto refrigerado",
			"stockActual": 9,
			"stockMinimo": 5,
			"precioVenta": 2.5,
			"estado": "activo"
		}
	],

	"clientes": [
		{
			"idOrigen": 1,
			"identificacion": "0999999998",
			"nombre": "Cliente académico",
			"correo": "cliente@example.com",
			"telefono": "0999999998",
			"estado": "activo"
		}
	],

	"ventas": [
		{
			"idOrigen": 1,
			"idClienteOrigen": 1,
			"numeroFactura": "POS-PRUEBA-001",
			"fechaVenta": "2026-07-30T00:10:00-05:00",
			"canal": "pos",
			"estado": "completado",
			"subtotal": 2.5,
			"descuento": 0,
			"iva": 0,
			"total": 2.5
		}
	],

	"detallesVenta": [
		{
			"idOrigen": 1,
			"idVentaOrigen": 1,
			"idProductoOrigen": 1,
			"cantidad": 1,
			"precioUnitario": 2.5,
			"subtotal": 2.5,
			"descuento": 0,
			"iva": 0,
			"total": 2.5
		}
	],

	"pedidos": [
		{
			"idOrigen": 1,
			"idEmpresaOrigen": 1,
			"motivo": "peticion",
			"estado": "completado",
			"fechaPedido": "2026-07-29T20:00:00-05:00",
			"fechaEsperada": "2026-07-30T08:00:00-05:00",
			"cantidadSolicitada": 10,
			"cantidadRecibida": 10,
			"total": 20
		}
	],

	"entregas": [
		{
			"idOrigen": 1,
			"idPedidoOrigen": 1,
			"idProveedorOrigen": 1,
			"fechaEntrega": "2026-07-30T07:45:00-05:00",
			"estado": "completa",
			"cantidadRecibida": 10
		}
	],

	"lotes": [
		{
			"idOrigen": 1,
			"idProductoOrigen": 1,
			"fechaCaducidad": "2026-08-30T00:00:00-05:00",
			"stock": 9,
			"estado": "vigente"
		}
	],

	"movimientos": [
		{
			"idOrigen": 1,
			"idProductoOrigen": 1,
			"idLoteOrigen": 1,
			"tipo": "salida_venta",
			"cantidad": -1,
			"stockProductoAnterior": 10,
			"stockProductoPosterior": 9,
			"stockLoteAnterior": 10,
			"stockLotePosterior": 9,
			"documentoOrigen": "POS-PRUEBA-001",
			"usuarioOrigen": "pos",
			"fechaMovimiento": "2026-07-30T00:10:00-05:00"
		}
	]
}`

func integrationTestConfig() config.Config {
	cfg := iotTestConfig()

	cfg.Integration = config.IntegrationConfig{
		SyncKey: "test-sync-key",
	}

	return cfg
}

func TestImportFullSnapshot(
	t *testing.T,
) {
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
			"se esperaba %d y se obtuvo %d: %s",
			http.StatusCreated,
			response.Code,
			response.Body.String(),
		)
	}

	var body struct {
		Success bool                 `json:"success"`
		Data    erpdata.ImportResult `json:"data"`
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

	if body.Data.Counts.Products != 1 {
		t.Fatalf(
			"se esperaba un producto y se obtuvieron %d",
			body.Data.Counts.Products,
		)
	}

	if body.Data.Counts.Sales != 1 {
		t.Fatalf(
			"se esperaba una venta y se obtuvieron %d",
			body.Data.Counts.Sales,
		)
	}
}

func TestSnapshotRequiresSyncKey(
	t *testing.T,
) {
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

	response := httptest.NewRecorder()

	router.ServeHTTP(
		response,
		request,
	)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf(
			"se esperaba %d y se obtuvo %d",
			http.StatusUnauthorized,
			response.Code,
		)
	}
}

func TestIntegrationStateAfterImport(
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
			"no se pudo preparar el snapshot: %s",
			importResponse.Body.String(),
		)
	}

	stateRequest := httptest.NewRequest(
		http.MethodGet,
		"/api/sig/integracion/estado",
		nil,
	)

	stateResponse := httptest.NewRecorder()

	router.ServeHTTP(
		stateResponse,
		stateRequest,
	)

	if stateResponse.Code != http.StatusOK {
		t.Fatalf(
			"se esperaba %d y se obtuvo %d: %s",
			http.StatusOK,
			stateResponse.Code,
			stateResponse.Body.String(),
		)
	}

	var body struct {
		Success bool                     `json:"success"`
		Data    erpdata.IntegrationState `json:"data"`
	}

	if err := json.NewDecoder(
		stateResponse.Body,
	).Decode(&body); err != nil {
		t.Fatalf(
			"no se pudo leer el estado: %v",
			err,
		)
	}

	if !body.Data.HasData {
		t.Fatal(
			"se esperaba tieneDatos=true",
		)
	}

	if body.Data.Counts.Movements != 1 {
		t.Fatalf(
			"se esperaba un movimiento y se obtuvieron %d",
			body.Data.Counts.Movements,
		)
	}
}
