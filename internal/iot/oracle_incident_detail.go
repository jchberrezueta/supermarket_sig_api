package iot

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

// FindIncident obtiene un incidente, su alerta y sus acciones correctivas.
func (repository *OracleRepository) FindIncident(
	ctx context.Context,
	incidentID int64,
) (
	IncidentDetail,
	bool,
	error,
) {
	if err := repository.validate(); err != nil {
		return IncidentDetail{},
			false,
			err
	}

	incident, exists, err :=
		repository.findIncidentByID(
			ctx,
			incidentID,
		)

	if err != nil {
		return IncidentDetail{},
			false,
			err
	}

	if !exists {
		return IncidentDetail{},
			false,
			nil
	}

	alert, err :=
		repository.findAlertByID(
			ctx,
			incident.AlertID,
		)

	if err != nil {
		return IncidentDetail{},
			false,
			err
	}

	actions, err :=
		repository.listCorrectiveActions(
			ctx,
			incident.ID,
		)

	if err != nil {
		return IncidentDetail{},
			false,
			err
	}

	return IncidentDetail{
			Incident: incident,
			Alert:    alert,
			Actions:  actions,
		},
		true,
		nil
}

func (repository *OracleRepository) findIncidentByID(
	ctx context.Context,
	incidentID int64,
) (
	Incident,
	bool,
	error,
) {
	const query = `
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
WHERE id = :1`

	row := repository.db.QueryRowContext(
		ctx,
		query,
		incidentID,
	)

	incident, err :=
		scanOracleIncident(
			row,
		)

	if errors.Is(
		err,
		sql.ErrNoRows,
	) {
		return Incident{},
			false,
			nil
	}

	if err != nil {
		return Incident{},
			false,
			fmt.Errorf(
				"no se pudo consultar el incidente %d en Oracle: %w",
				incidentID,
				err,
			)
	}

	return incident,
		true,
		nil
}

func (repository *OracleRepository) findAlertByID(
	ctx context.Context,
	alertID int64,
) (
	Alert,
	error,
) {
	const query = `
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
WHERE id = :1`

	row := repository.db.QueryRowContext(
		ctx,
		query,
		alertID,
	)

	alert, err :=
		scanOracleAlert(
			row,
		)

	if errors.Is(
		err,
		sql.ErrNoRows,
	) {
		return Alert{},
			fmt.Errorf(
				"la alerta %d asociada al incidente no existe",
				alertID,
			)
	}

	if err != nil {
		return Alert{},
			fmt.Errorf(
				"no se pudo consultar la alerta %d en Oracle: %w",
				alertID,
				err,
			)
	}

	return alert,
		nil
}

func (repository *OracleRepository) listCorrectiveActions(
	ctx context.Context,
	incidentID int64,
) (
	[]CorrectiveAction,
	error,
) {
	const query = `
SELECT id,
       id_incidente,
       descripcion,
       responsable,
       resultado,
       fecha_accion
FROM sig_accion_correctiva
WHERE id_incidente = :1
ORDER BY fecha_accion ASC,
         id ASC`

	rows, err := repository.db.QueryContext(
		ctx,
		query,
		incidentID,
	)

	if err != nil {
		return nil,
			fmt.Errorf(
				"no se pudieron consultar las acciones correctivas del incidente %d: %w",
				incidentID,
				err,
			)
	}

	defer rows.Close()

	actions := make(
		[]CorrectiveAction,
		0,
	)

	for rows.Next() {
		action, err :=
			scanOracleCorrectiveAction(
				rows,
			)

		if err != nil {
			return nil,
				fmt.Errorf(
					"no se pudo leer una acción correctiva desde Oracle: %w",
					err,
				)
		}

		actions = append(
			actions,
			action,
		)
	}

	if err := rows.Err(); err != nil {
		return nil,
			fmt.Errorf(
				"falló la consulta de acciones correctivas del incidente %d: %w",
				incidentID,
				err,
			)
	}

	return actions,
		nil
}

func scanOracleCorrectiveAction(
	scanner oracleRowScanner,
) (
	CorrectiveAction,
	error,
) {
	var action CorrectiveAction
	var result sql.NullString

	err := scanner.Scan(
		&action.ID,
		&action.IncidentID,
		&action.Description,
		&action.Responsible,
		&result,
		&action.CreatedAt,
	)

	if err != nil {
		return CorrectiveAction{},
			err
	}

	if result.Valid {
		action.Result =
			result.String
	}

	return action,
		nil
}
