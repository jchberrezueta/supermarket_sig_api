package iot

import (
	"context"
	"testing"
)

func TestServiceAvoidsDuplicateActiveIncidents(
	t *testing.T,
) {
	t.Parallel()

	ctx := context.Background()

	repository :=
		NewMemoryRepository()

	service :=
		NewService(
			repository,
			"ESP32-BODEGA-01",
			2,
			8,
		)

	firstResult, err :=
		service.Register(
			ctx,
			RegisterInput{
				DeviceCode:
					"ESP32-BODEGA-01",

				Temperature: 12.5,
			},
		)

	if err != nil {
		t.Fatalf(
			"no se pudo registrar la primera desviación: %v",
			err,
		)
	}

	if firstResult.Alert == nil {
		t.Fatal(
			"la primera desviación debía crear una alerta",
		)
	}

	if firstResult.Incident == nil {
		t.Fatal(
			"la primera desviación debía crear un incidente",
		)
	}

	firstIncidentID :=
		firstResult.Incident.ID

	secondResult, err :=
		service.Register(
			ctx,
			RegisterInput{
				DeviceCode:
					"ESP32-BODEGA-01",

				Temperature: 13,
			},
		)

	if err != nil {
		t.Fatalf(
			"no se pudo registrar la segunda desviación: %v",
			err,
		)
	}

	if secondResult.Reading.Status !=
		ReadingStatusOutOfRange {
		t.Fatalf(
			"estado inesperado para la segunda lectura: %q",
			secondResult.Reading.Status,
		)
	}

	if secondResult.Alert != nil {
		t.Fatal(
			"la segunda desviación no debía crear otra alerta mientras existe un incidente activo",
		)
	}

	if secondResult.Incident != nil {
		t.Fatal(
			"la segunda desviación no debía crear otro incidente mientras existe uno activo",
		)
	}

	readings, err :=
		service.ListReadings(
			ctx,
			"ESP32-BODEGA-01",
			20,
		)

	if err != nil {
		t.Fatalf(
			"no se pudo consultar el historial: %v",
			err,
		)
	}

	if len(readings) != 2 {
		t.Fatalf(
			"se esperaban 2 lecturas y se obtuvieron %d",
			len(readings),
		)
	}

	alerts, err :=
		service.ListAlerts(
			ctx,
			"",
			20,
		)

	if err != nil {
		t.Fatalf(
			"no se pudieron consultar las alertas: %v",
			err,
		)
	}

	if len(alerts) != 1 {
		t.Fatalf(
			"se esperaba 1 alerta y se obtuvieron %d",
			len(alerts),
		)
	}

	incidents, err :=
		service.ListIncidents(
			ctx,
			"",
			20,
		)

	if err != nil {
		t.Fatalf(
			"no se pudieron consultar los incidentes: %v",
			err,
		)
	}

	if len(incidents) != 1 {
		t.Fatalf(
			"se esperaba 1 incidente y se obtuvieron %d",
			len(incidents),
		)
	}

	_, err =
		service.RecognizeIncident(
			ctx,
			firstIncidentID,
			"Gerente de bodega",
		)

	if err != nil {
		t.Fatalf(
			"no se pudo reconocer el incidente: %v",
			err,
		)
	}

	_, err =
		service.AddCorrectiveAction(
			ctx,
			firstIncidentID,
			"Se ajustó el sistema de refrigeración.",
			"Técnico de mantenimiento",
			"La temperatura comenzó a descender.",
		)

	if err != nil {
		t.Fatalf(
			"no se pudo registrar la acción correctiva: %v",
			err,
		)
	}

	_, err =
		service.ResolveIncident(
			ctx,
			firstIncidentID,
		)

	if err != nil {
		t.Fatalf(
			"no se pudo resolver el incidente: %v",
			err,
		)
	}

	closedDetail, err :=
		service.CloseIncident(
			ctx,
			firstIncidentID,
		)

	if err != nil {
		t.Fatalf(
			"no se pudo cerrar el incidente: %v",
			err,
		)
	}

	if closedDetail.Incident.Status !=
		"cerrado" {
		t.Fatalf(
			"estado final inesperado: %q",
			closedDetail.Incident.Status,
		)
	}

	thirdResult, err :=
		service.Register(
			ctx,
			RegisterInput{
				DeviceCode:
					"ESP32-BODEGA-01",

				Temperature: 11.5,
			},
		)

	if err != nil {
		t.Fatalf(
			"no se pudo registrar una nueva desviación después del cierre: %v",
			err,
		)
	}

	if thirdResult.Alert == nil {
		t.Fatal(
			"una nueva desviación después del cierre debía crear una alerta",
		)
	}

	if thirdResult.Incident == nil {
		t.Fatal(
			"una nueva desviación después del cierre debía crear un incidente",
		)
	}

	if thirdResult.Incident.ID ==
		firstIncidentID {
		t.Fatal(
			"el nuevo incidente no puede reutilizar el identificador anterior",
		)
	}

	summary, err :=
		service.Summary(
			ctx,
			"ESP32-BODEGA-01",
		)

	if err != nil {
		t.Fatalf(
			"no se pudo calcular el resumen IoT: %v",
			err,
		)
	}

	if summary.TotalReadings != 3 {
		t.Fatalf(
			"se esperaban 3 lecturas y se obtuvieron %d",
			summary.TotalReadings,
		)
	}

	if summary.OutOfRangeReadings != 3 {
		t.Fatalf(
			"se esperaban 3 lecturas fuera de rango y se obtuvieron %d",
			summary.OutOfRangeReadings,
		)
	}

	if summary.Alerts.Closed != 1 {
		t.Fatalf(
			"se esperaba 1 alerta cerrada y se obtuvieron %d",
			summary.Alerts.Closed,
		)
	}

	if summary.Alerts.Open != 1 {
		t.Fatalf(
			"se esperaba 1 alerta abierta y se obtuvieron %d",
			summary.Alerts.Open,
		)
	}

	if summary.Incidents.Closed != 1 {
		t.Fatalf(
			"se esperaba 1 incidente cerrado y se obtuvieron %d",
			summary.Incidents.Closed,
		)
	}

	if summary.Incidents.Open != 1 {
		t.Fatalf(
			"se esperaba 1 incidente abierto y se obtuvieron %d",
			summary.Incidents.Open,
		)
	}
}