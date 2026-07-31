package iot

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

// Esta verificación obliga al compilador a confirmar que OracleRepository
// implementa todos los métodos definidos por Repository.
var _ Repository = (*OracleRepository)(nil)

// ListAudit devuelve los eventos de auditoría más recientes.
func (repository *OracleRepository) ListAudit(
	ctx context.Context,
	action string,
	limit int,
) (
	[]AuditEvent,
	error,
) {
	if err := repository.validate(); err != nil {
		return nil, err
	}

	action = strings.TrimSpace(
		action,
	)

	var (
		query string
		args  []any
	)

	if action == "" {
		query = `
SELECT id,
       actor,
       accion,
       modulo,
       entidad,
       id_registro,
       resultado,
       detalle,
       fecha_evento
FROM (
    SELECT id,
           actor,
           accion,
           modulo,
           entidad,
           id_registro,
           resultado,
           detalle,
           fecha_evento
    FROM sig_auditoria
    ORDER BY fecha_evento DESC,
             id DESC
)
WHERE ROWNUM <= :1`

		args = []any{
			limit,
		}
	} else {
		query = `
SELECT id,
       actor,
       accion,
       modulo,
       entidad,
       id_registro,
       resultado,
       detalle,
       fecha_evento
FROM (
    SELECT id,
           actor,
           accion,
           modulo,
           entidad,
           id_registro,
           resultado,
           detalle,
           fecha_evento
    FROM sig_auditoria
    WHERE accion = :1
    ORDER BY fecha_evento DESC,
             id DESC
)
WHERE ROWNUM <= :2`

		args = []any{
			action,
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
				"no se pudo consultar la auditoría IoT en Oracle: %w",
				err,
			)
	}

	defer rows.Close()

	events := make(
		[]AuditEvent,
		0,
	)

	for rows.Next() {
		event, err := scanOracleAuditEvent(
			rows,
		)

		if err != nil {
			return nil,
				fmt.Errorf(
					"no se pudo leer un evento de auditoría desde Oracle: %w",
					err,
				)
		}

		events = append(
			events,
			event,
		)
	}

	if err := rows.Err(); err != nil {
		return nil,
			fmt.Errorf(
				"falló la consulta de auditoría IoT en Oracle: %w",
				err,
			)
	}

	return events,
		nil
}

// Summary calcula los indicadores de cadena de frío para un dispositivo.
func (repository *OracleRepository) Summary(
	ctx context.Context,
	deviceCode string,
) (
	IoTSummary,
	error,
) {
	if err := repository.validate(); err != nil {
		return IoTSummary{},
			err
	}

	deviceCode = strings.TrimSpace(
		deviceCode,
	)

	if deviceCode == "" {
		deviceCode = repository.deviceCode
	}

	const query = `
SELECT COUNT(lectura.id) AS total_lecturas,

       NVL(
           SUM(
               CASE
                   WHEN lectura.estado = 'normal'
                   THEN 1
                   ELSE 0
               END
           ),
           0
       ) AS lecturas_normales,

       NVL(
           SUM(
               CASE
                   WHEN lectura.estado = 'fuera_rango'
                   THEN 1
                   ELSE 0
               END
           ),
           0
       ) AS lecturas_fuera_rango,

       NVL(
           SUM(
               CASE
                   WHEN alerta.estado = 'abierta'
                   THEN 1
                   ELSE 0
               END
           ),
           0
       ) AS alertas_abiertas,

       NVL(
           SUM(
               CASE
                   WHEN alerta.estado = 'reconocida'
                   THEN 1
                   ELSE 0
               END
           ),
           0
       ) AS alertas_reconocidas,

       NVL(
           SUM(
               CASE
                   WHEN alerta.estado = 'cerrada'
                   THEN 1
                   ELSE 0
               END
           ),
           0
       ) AS alertas_cerradas,

       NVL(
           SUM(
               CASE
                   WHEN incidente.estado = 'abierto'
                   THEN 1
                   ELSE 0
               END
           ),
           0
       ) AS incidentes_abiertos,

       NVL(
           SUM(
               CASE
                   WHEN incidente.estado = 'reconocido'
                   THEN 1
                   ELSE 0
               END
           ),
           0
       ) AS incidentes_reconocidos,

       NVL(
           SUM(
               CASE
                   WHEN incidente.estado = 'en_tratamiento'
                   THEN 1
                   ELSE 0
               END
           ),
           0
       ) AS incidentes_en_tratamiento,

       NVL(
           SUM(
               CASE
                   WHEN incidente.estado = 'resuelto'
                   THEN 1
                   ELSE 0
               END
           ),
           0
       ) AS incidentes_resueltos,

       NVL(
           SUM(
               CASE
                   WHEN incidente.estado = 'cerrado'
                   THEN 1
                   ELSE 0
               END
           ),
           0
       ) AS incidentes_cerrados

FROM sig_lectura_iot lectura

LEFT JOIN sig_alerta alerta
       ON alerta.id_lectura = lectura.id

LEFT JOIN sig_incidente incidente
       ON incidente.id_alerta = alerta.id

WHERE lectura.codigo_dispositivo = :1`

	summary := IoTSummary{
		DeviceCode: deviceCode,

		Recommendations: make(
			[]Recommendation,
			0,
		),

		UpdatedAt: time.Now(),
	}

	err := repository.db.QueryRowContext(
		ctx,
		query,
		deviceCode,
	).Scan(
		&summary.TotalReadings,
		&summary.NormalReadings,
		&summary.OutOfRangeReadings,
		&summary.Alerts.Open,
		&summary.Alerts.Recognized,
		&summary.Alerts.Closed,
		&summary.Incidents.Open,
		&summary.Incidents.Recognized,
		&summary.Incidents.InTreatment,
		&summary.Incidents.Resolved,
		&summary.Incidents.Closed,
	)

	if err != nil {
		return IoTSummary{},
			fmt.Errorf(
				"no se pudo calcular el resumen IoT en Oracle: %w",
				err,
			)
	}

	latestReading,
		exists,
		err := repository.Latest(
		ctx,
		deviceCode,
	)

	if err != nil {
		return IoTSummary{},
			err
	}

	if exists {
		summary.LatestReading =
			&latestReading
	}

	if summary.TotalReadings > 0 {
		summary.NormalPercentage =
			float64(
				summary.NormalReadings,
			) *
				100 /
				float64(
					summary.TotalReadings,
				)
	}

	return summary,
		nil
}

func scanOracleAuditEvent(
	scanner oracleRowScanner,
) (
	AuditEvent,
	error,
) {
	var event AuditEvent
	var detail sql.NullString

	err := scanner.Scan(
		&event.ID,
		&event.Actor,
		&event.Action,
		&event.Module,
		&event.Entity,
		&event.RecordID,
		&event.Result,
		&detail,
		&event.OccurredAt,
	)

	if err != nil {
		return AuditEvent{},
			err
	}

	if detail.Valid {
		event.Detail =
			detail.String
	}

	return event,
		nil
}
