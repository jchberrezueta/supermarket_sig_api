package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"supermarket-sig-api/internal/erpdata"
)

func TestRuntimeSharesERPServiceWithRouter(
	t *testing.T,
) {
	cfg := testConfig()

	cfg.Integration.SyncKey =
		"runtime-test-key"

	runtime :=
		NewRuntime(
			cfg,
			nil,
		)

	snapshot := erpdata.Snapshot{
		ContractVersion: erpdata.SnapshotContractVersion,

		Source: erpdata.SnapshotSourceERP,

		GeneratedAt: time.Now().UTC(),
	}

	body, err :=
		json.Marshal(
			snapshot,
		)

	if err != nil {
		t.Fatalf(
			"no se pudo serializar el snapshot: %v",
			err,
		)
	}

	request :=
		httptest.NewRequest(
			http.MethodPost,
			"/api/sig/integracion/snapshot",
			bytes.NewReader(
				body,
			),
		)

	request.Header.Set(
		"Content-Type",
		"application/json",
	)

	request.Header.Set(
		"X-SIG-Key",
		cfg.Integration.SyncKey,
	)

	response :=
		httptest.NewRecorder()

	runtime.Handler.ServeHTTP(
		response,
		request,
	)

	if response.Code !=
		http.StatusCreated {
		t.Fatalf(
			"se esperaba estado %d y se obtuvo %d: %s",
			http.StatusCreated,
			response.Code,
			response.Body.String(),
		)
	}

	state, err :=
		runtime.ERPService.State(
			context.Background(),
		)

	if err != nil {
		t.Fatalf(
			"no se pudo consultar el servicio compartido: %v",
			err,
		)
	}

	if !state.HasData {
		t.Fatal(
			"el router no utilizó la misma instancia de ERPService expuesta por Runtime",
		)
	}

	if state.LastImport == nil {
		t.Fatal(
			"la importación realizada por el router no quedó disponible en ERPService",
		)
	}

	if state.LastImport.ContractVersion !=
		erpdata.SnapshotContractVersion {
		t.Fatalf(
			"versión inesperada: %q",
			state.LastImport.ContractVersion,
		)
	}
}
