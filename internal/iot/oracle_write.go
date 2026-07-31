package iot

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

// Save guarda atómicamente una lectura y los resultados BPM generados.
func (repository *OracleRepository) Save(
	ctx context.Context,
	result *RegistrationResult,
) error {
	if err := repository.validate(); err != nil {
		return err
	}

	if result == nil {
		return errors.New(
			"el resultado IoT que se desea guardar es obligatorio",
		)
	}

	if strings.TrimSpace(
		result.Reading.DeviceCode,
	) == "" {
		return errors.New(
			"el código del dispositivo de la lectura es obligatorio",
		)
	}

	if result.Reading.DeviceCode !=
		repository.deviceCode {
		return fmt.Errorf(
			"el dispositivo %q no coincide con el dispositivo configurado %q",
			result.Reading.DeviceCode,
			repository.deviceCode,
		)
	}

	if result.Incident != nil &&
		result.Alert == nil {
		return errors.New(
			"no se puede guardar un incidente IoT sin una alerta asociada",
		)
	}

	tx, err := repository.db.BeginTx(
		ctx,
		nil,
	)

	if err != nil {
		return fmt.Errorf(
			"no se pudo iniciar la transacción IoT en Oracle: %w",
			err,
		)
	}

	defer func() {
		_ = tx.Rollback()
	}()

	if err := repository.ensureDevice(
		ctx,
		tx,
	); err != nil {
		return err
	}

	if err := lockOracleDevice(
		ctx,
		tx,
		result.Reading.DeviceCode,
	); err != nil {
		return err
	}

	shouldCreateAlert :=
		result.Alert != nil

	shouldCreateIncident :=
		result.Incident != nil

	if shouldCreateAlert &&
		shouldCreateIncident {
		hasActiveIncident, err :=
			hasActiveIncidentForDevice(
				ctx,
				tx,
				result.Reading.DeviceCode,
			)

		if err != nil {
			return err
		}

		if hasActiveIncident {
			shouldCreateAlert = false
			shouldCreateIncident = false
		}
	}

	now := time.Now()

	savedReading := result.Reading
	savedReading.CreatedAt = now

	readingID, err := insertOracleReading(
		ctx,
		tx,
		savedReading,
	)

	if err != nil {
		return err
	}

	savedReading.ID = readingID

	if err := insertOracleAudit(
		ctx,
		tx,
		AuditEvent{
			Actor:    savedReading.DeviceCode,
			Action:   "lectura_registrada",
			Module:   "iot",
			Entity:   "lectura_sensor",
			RecordID: fmt.Sprint(savedReading.ID),
			Result:   "correcto",
			Detail: fmt.Sprintf(
				"Temperatura %.2f °C, estado %s.",
				savedReading.Temperature,
				savedReading.Status,
			),
			OccurredAt: now,
		},
	); err != nil {
		return err
	}

	var savedAlert *Alert

	if shouldCreateAlert {
		alert := *result.Alert

		alert.ReadingID = savedReading.ID
		alert.CreatedAt = now

		alertID, err := insertOracleAlert(
			ctx,
			tx,
			alert,
		)

		if err != nil {
			return err
		}

		alert.ID = alertID
		savedAlert = &alert

		if err := insertOracleAudit(
			ctx,
			tx,
			AuditEvent{
				Actor:      "sistema",
				Action:     "alerta_creada",
				Module:     "cadena_frio",
				Entity:     "alerta",
				RecordID:   fmt.Sprint(alert.ID),
				Result:     "correcto",
				Detail:     alert.Message,
				OccurredAt: now,
			},
		); err != nil {
			return err
		}
	}

	var savedIncident *Incident

	if shouldCreateIncident {
		incident := *result.Incident

		incident.AlertID = savedAlert.ID
		incident.CreatedAt = now

		incidentID, err := insertOracleIncident(
			ctx,
			tx,
			incident,
		)

		if err != nil {
			return err
		}

		incident.ID = incidentID
		savedIncident = &incident

		if err := insertOracleAudit(
			ctx,
			tx,
			AuditEvent{
				Actor:      "sistema",
				Action:     "incidente_creado",
				Module:     "calidad",
				Entity:     "incidente",
				RecordID:   fmt.Sprint(incident.ID),
				Result:     "correcto",
				Detail:     incident.Description,
				OccurredAt: now,
			},
		); err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf(
			"no se pudo confirmar la transacción IoT en Oracle: %w",
			err,
		)
	}

	result.Reading = savedReading
	result.Alert = savedAlert
	result.Incident = savedIncident

	return nil
}

func insertOracleReading(
	ctx context.Context,
	tx *sql.Tx,
	reading Reading,
) (
	int64,
	error,
) {
	const statement = `
INSERT INTO sig_lectura_iot (
    codigo_dispositivo,
    temperatura,
    humedad,
    estado,
    fecha_lectura,
    fecha_registro
) VALUES (
    :1,
    :2,
    :3,
    :4,
    :5,
    :6
)
RETURNING id INTO :7`

	var humidity any

	if reading.Humidity != nil {
		humidity = *reading.Humidity
	}

	var readingID int64

	_, err := tx.ExecContext(
		ctx,
		statement,
		reading.DeviceCode,
		reading.Temperature,
		humidity,
		reading.Status,
		reading.ReadAt,
		reading.CreatedAt,
		sql.Out{
			Dest: &readingID,
		},
	)

	if err != nil {
		return 0,
			fmt.Errorf(
				"no se pudo guardar la lectura IoT en Oracle: %w",
				err,
			)
	}

	return readingID, nil
}

func insertOracleAlert(
	ctx context.Context,
	tx *sql.Tx,
	alert Alert,
) (
	int64,
	error,
) {
	const statement = `
INSERT INTO sig_alerta (
    id_lectura,
    tipo,
    severidad,
    mensaje,
    estado,
    fecha_apertura,
    fecha_reconocimiento,
    fecha_cierre
) VALUES (
    :1,
    :2,
    :3,
    :4,
    :5,
    :6,
    :7,
    :8
)
RETURNING id INTO :9`

	var alertID int64

	_, err := tx.ExecContext(
		ctx,
		statement,
		alert.ReadingID,
		alert.Type,
		alert.Severity,
		alert.Message,
		alert.Status,
		alert.CreatedAt,
		alert.RecognizedAt,
		alert.ClosedAt,
		sql.Out{
			Dest: &alertID,
		},
	)

	if err != nil {
		return 0,
			fmt.Errorf(
				"no se pudo guardar la alerta IoT en Oracle: %w",
				err,
			)
	}

	return alertID, nil
}

func insertOracleIncident(
	ctx context.Context,
	tx *sql.Tx,
	incident Incident,
) (
	int64,
	error,
) {
	const statement = `
INSERT INTO sig_incidente (
    id_alerta,
    codigo,
    titulo,
    descripcion,
    severidad,
    estado,
    responsable,
    fecha_apertura,
    fecha_reconocimiento,
    fecha_resolucion,
    fecha_cierre
) VALUES (
    :1,
    :2,
    :3,
    :4,
    :5,
    :6,
    :7,
    :8,
    :9,
    :10,
    :11
)
RETURNING id INTO :12`

	var incidentID int64

	_, err := tx.ExecContext(
		ctx,
		statement,
		incident.AlertID,
		incident.Code,
		incident.Title,
		incident.Description,
		incident.Severity,
		incident.Status,
		nullableString(incident.Responsible),
		incident.CreatedAt,
		incident.RecognizedAt,
		incident.ResolvedAt,
		incident.ClosedAt,
		sql.Out{
			Dest: &incidentID,
		},
	)

	if err != nil {
		return 0,
			fmt.Errorf(
				"no se pudo guardar el incidente IoT en Oracle: %w",
				err,
			)
	}

	return incidentID, nil
}

func insertOracleAudit(
	ctx context.Context,
	tx *sql.Tx,
	event AuditEvent,
) error {
	const statement = `
INSERT INTO sig_auditoria (
    actor,
    accion,
    modulo,
    entidad,
    id_registro,
    resultado,
    detalle,
    fecha_evento
) VALUES (
    :1,
    :2,
    :3,
    :4,
    :5,
    :6,
    :7,
    :8
)`

	_, err := tx.ExecContext(
		ctx,
		statement,
		event.Actor,
		event.Action,
		event.Module,
		event.Entity,
		event.RecordID,
		event.Result,
		nullableString(event.Detail),
		event.OccurredAt,
	)

	if err != nil {
		return fmt.Errorf(
			"no se pudo guardar la auditoría IoT en Oracle: %w",
			err,
		)
	}

	return nil
}

func nullableString(
	value string,
) any {
	value = strings.TrimSpace(value)

	if value == "" {
		return nil
	}

	return value
}
