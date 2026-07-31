package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"supermarket-sig-api/internal/erpdata"
)

func TestSynchronizeDirectlyFromERP(
	t *testing.T,
) {
	upstream := httptest.NewServer(
		http.HandlerFunc(
			func(
				w http.ResponseWriter,
				r *http.Request,
			) {
				if r.Method != http.MethodGet {
					t.Fatalf(
						"método inesperado: %s",
						r.Method,
					)
				}

				if r.URL.Path !=
					"/api/integracion/sig/snapshot" {
					t.Fatalf(
						"ruta inesperada: %s",
						r.URL.Path,
					)
				}

				if r.Header.Get(
					"X-SIG-Key",
				) != "test-sync-key" {
					t.Fatal(
						"el cliente no envió la clave de integración",
					)
				}

				w.Header().Set(
					"Content-Type",
					"application/json",
				)

				w.WriteHeader(
					http.StatusOK,
				)

				_, _ = w.Write(
					[]byte(
						validSnapshotJSON,
					),
				)
			},
		),
	)

	defer upstream.Close()

	cfg := integrationTestConfig()

	cfg.Integration.ERPBaseURL =
		upstream.URL + "/api"

	cfg.Integration.Timeout =
		2 * time.Second

	router := NewRouter(
		cfg,
		nil,
	)

	request := httptest.NewRequest(
		http.MethodPost,
		"/api/sig/integracion/sincronizar",
		nil,
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

	if response.Code != http.StatusOK {
		t.Fatalf(
			"se esperaba %d y se obtuvo %d: %s",
			http.StatusOK,
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

	if body.Data.ContractVersion != "1.0" {
		t.Fatalf(
			"versión inesperada: %q",
			body.Data.ContractVersion,
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

func TestSynchronizeHandlesERPFailure(
	t *testing.T,
) {
	upstream := httptest.NewServer(
		http.HandlerFunc(
			func(
				w http.ResponseWriter,
				_ *http.Request,
			) {
				w.Header().Set(
					"Content-Type",
					"application/json",
				)

				w.WriteHeader(
					http.StatusUnauthorized,
				)

				_, _ = w.Write(
					[]byte(
						`{
							"statusCode": 401,
							"message": "Unauthorized"
						}`,
					),
				)
			},
		),
	)

	defer upstream.Close()

	cfg := integrationTestConfig()

	cfg.Integration.ERPBaseURL =
		upstream.URL + "/api"

	cfg.Integration.Timeout =
		2 * time.Second

	router := NewRouter(
		cfg,
		nil,
	)

	request := httptest.NewRequest(
		http.MethodPost,
		"/api/sig/integracion/sincronizar",
		nil,
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

	if response.Code !=
		http.StatusBadGateway {
		t.Fatalf(
			"se esperaba %d y se obtuvo %d: %s",
			http.StatusBadGateway,
			response.Code,
			response.Body.String(),
		)
	}

	var body struct {
		Success bool `json:"success"`

		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}

	if err := json.NewDecoder(
		response.Body,
	).Decode(&body); err != nil {
		t.Fatalf(
			"no se pudo leer la respuesta: %v",
			err,
		)
	}

	if body.Success {
		t.Fatal(
			"se esperaba success=false",
		)
	}

	if body.Error.Code !=
		"erp_upstream_error" {
		t.Fatalf(
			"código inesperado: %q",
			body.Error.Code,
		)
	}
}
