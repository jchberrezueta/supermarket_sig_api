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
	"supermarket-sig-api/internal/iot"
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

func TestRuntimeSharesIoTServiceWithRouter(
	t *testing.T,
) {
	cfg := iotTestConfig()

	runtime :=
		NewRuntime(
			cfg,
			nil,
		)

	if runtime.IoTService == nil {
		t.Fatal(
			"Runtime no expuso el servicio IoT compartido",
		)
	}

	request :=
		httptest.NewRequest(
			http.MethodPost,
			"/api/sig/iot/lecturas",
			bytes.NewBufferString(
				`{
					"codigoDispositivo": "ESP32-BODEGA-01",
					"temperatura": 4.5,
					"humedad": 65
				}`,
			),
		)

	request.Header.Set(
		"Content-Type",
		"application/json",
	)

	request.Header.Set(
		"X-Device-Key",
		cfg.IoT.DeviceKey,
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

	reading,
		exists,
		err := runtime.IoTService.Latest(
		context.Background(),
		cfg.IoT.DeviceCode,
	)

	if err != nil {
		t.Fatalf(
			"no se pudo consultar el servicio IoT compartido: %v",
			err,
		)
	}

	if !exists {
		t.Fatal(
			"la lectura registrada mediante el router no quedó disponible en IoTService",
		)
	}

	if reading.DeviceCode !=
		cfg.IoT.DeviceCode {
		t.Fatalf(
			"código de dispositivo inesperado: %q",
			reading.DeviceCode,
		)
	}

	if reading.Status !=
		iot.ReadingStatusNormal {
		t.Fatalf(
			"estado de lectura inesperado: %q",
			reading.Status,
		)
	}

	if reading.Temperature != 4.5 {
		t.Fatalf(
			"temperatura inesperada: %.2f",
			reading.Temperature,
		)
	}

	summary, err :=
		runtime.IoTService.Summary(
			context.Background(),
			cfg.IoT.DeviceCode,
		)

	if err != nil {
		t.Fatalf(
			"no se pudo consultar el resumen IoT compartido: %v",
			err,
		)
	}

	if summary.TotalReadings != 1 {
		t.Fatalf(
			"se esperaba 1 lectura y se obtuvieron %d",
			summary.TotalReadings,
		)
	}

	if summary.NormalReadings != 1 {
		t.Fatalf(
			"se esperaba 1 lectura normal y se obtuvieron %d",
			summary.NormalReadings,
		)
	}
}
