package iot

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

// ListAlerts devuelve las alertas más recientes primero.
func (repository *OracleRepository) ListAlerts(
	ctx context.Context,
	status string,
	limit int,
) (
	[]Alert,
	error,
) {
	if err := repository.validate(); err != nil {
		return nil, err
	}

	status = strings.TrimSpace(
		strings.ToLower(status),
	)

	var (
		query string
		args  []any
	)

	if status == "" {
		query = `
SELECT id,
       id_lectura,
       tipo,
       severidad,
       mensaje,
       estado,
       fecha_apertura,
       fecha_reconocimiento,
       fecha_cierre
FROM (
    SELECT id,
           id_lectura,
           tipo,
           severidad,
           mensaje,
           estado,
           fecha_apertura,
           fecha_reconocimiento,
           fecha_cierre
    FROM sig_alerta
    ORDER BY fecha_apertura DESC,
             id DESC
)
WHERE ROWNUM <= :1`

		args = []any{
			limit,
		}
	} else {
		query = `
SELECT id,
       id_lectura,
       tipo,
       severidad,
       mensaje,
       estado,
       fecha_apertura,
       fecha_reconocimiento,
       fecha_cierre
FROM (
    SELECT id,
           id_lectura,
           tipo,
           severidad,
           mensaje,
           estado,
           fecha_apertura,
           fecha_reconocimiento,
           fecha_cierre
    FROM sig_alerta
    WHERE estado = :1
    ORDER BY fecha_apertura DESC,
             id DESC
)
WHERE ROWNUM <= :2`

		args = []any{
			status,
			limit,
		}
	}

	rows, err := repository.db.QueryContext(
		ctx,
		query,
		args...,
	)

	if err != nil {
		return nil,
			fmt.Errorf(
				"no se pudieron consultar las alertas IoT en Oracle: %w",
				err,
			)
	}

	defer rows.Close()

	alerts := make(
		[]Alert,
		0,
	)

	for rows.Next() {
		alert, err := scanOracleAlert(
			rows,
		)

		if err != nil {
			return nil,
				fmt.Errorf(
					"no se pudo leer una alerta IoT desde Oracle: %w",
					err,
				)
		}

		alerts = append(
			alerts,
			alert,
		)
	}

	if err := rows.Err(); err != nil {
		return nil,
			fmt.Errorf(
				"falló la consulta de alertas IoT en Oracle: %w",
				err,
			)
	}

	return alerts,
		nil
}

// ListIncidents devuelve los incidentes más recientes primero.
func (repository *OracleRepository) ListIncidents(
	ctx context.Context,
	status string,
	limit int,
) (
	[]Incident,
	error,
) {
	if err := repository.validate(); err != nil {
		return nil, err
	}

	status = strings.TrimSpace(
		strings.ToLower(status),
	)

	var (
		query string
		args  []any
	)

	if status == "" {
		query = `
SELECT id,
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
FROM (
    SELECT id,
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
    FROM sig_incidente
    ORDER BY fecha_apertura DESC,
             id DESC
)
WHERE ROWNUM <= :1`

		args = []any{
			limit,
		}
	} else {
		query = `
SELECT id,
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
FROM (
    SELECT id,
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
    FROM sig_incidente
    WHERE estado = :1
    ORDER BY fecha_apertura DESC,
             id DESC
)
WHERE ROWNUM <= :2`

		args = []any{
			status,
			limit,
		}
	}

	rows, err := repository.db.QueryContext(
		ctx,
		query,
		args...,
	)

	if err != nil {
		return nil,
			fmt.Errorf(
				"no se pudieron consultar los incidentes IoT en Oracle: %w",
				err,
			)
	}

	defer rows.Close()

	incidents := make(
		[]Incident,
		0,
	)

	for rows.Next() {
		incident, err := scanOracleIncident(
			rows,
		)

		if err != nil {
			return nil,
				fmt.Errorf(
					"no se pudo leer un incidente IoT desde Oracle: %w",
					err,
				)
		}

		incidents = append(
			incidents,
			incident,
		)
	}

	if err := rows.Err(); err != nil {
		return nil,
			fmt.Errorf(
				"falló la consulta de incidentes IoT en Oracle: %w",
				err,
			)
	}

	return incidents,
		nil
}

func scanOracleAlert(
	scanner oracleRowScanner,
) (
	Alert,
	error,
) {
	var alert Alert

	var recognizedAt sql.NullTime
	var closedAt sql.NullTime

	err := scanner.Scan(
		&alert.ID,
		&alert.ReadingID,
		&alert.Type,
		&alert.Severity,
		&alert.Message,
		&alert.Status,
		&alert.CreatedAt,
		&recognizedAt,
		&closedAt,
	)

	if err != nil {
		return Alert{},
			err
	}

	alert.RecognizedAt =
		nullableTimePointer(
			recognizedAt,
		)

	alert.ClosedAt =
		nullableTimePointer(
			closedAt,
		)

	return alert,
		nil
}

func scanOracleIncident(
	scanner oracleRowScanner,
) (
	Incident,
	error,
) {
	var incident Incident

	var responsible sql.NullString
	var recognizedAt sql.NullTime
	var resolvedAt sql.NullTime
	var closedAt sql.NullTime

	err := scanner.Scan(
		&incident.ID,
		&incident.AlertID,
		&incident.Code,
		&incident.Title,
		&incident.Description,
		&incident.Severity,
		&incident.Status,
		&responsible,
		&incident.CreatedAt,
		&recognizedAt,
		&resolvedAt,
		&closedAt,
	)

	if err != nil {
		return Incident{},
			err
	}

	if responsible.Valid {
		incident.Responsible =
			responsible.String
	}

	incident.RecognizedAt =
		nullableTimePointer(
			recognizedAt,
		)

	incident.ResolvedAt =
		nullableTimePointer(
			resolvedAt,
		)

	incident.ClosedAt =
		nullableTimePointer(
			closedAt,
		)

	return incident,
		nil
}

func nullableTimePointer(
	value sql.NullTime,
) *time.Time {
	if !value.Valid {
		return nil
	}

	timeValue := value.Time

	return &timeValue
}
