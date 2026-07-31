package iot

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

// ApplyIncidentWorkflow guarda atómicamente una transición BPM,
// la actualización de la alerta y una posible acción correctiva.
func (repository *OracleRepository) ApplyIncidentWorkflow(
	ctx context.Context,
	detail *IncidentDetail,
	newAction *CorrectiveAction,
) error {
	if err := repository.validate(); err != nil {
		return err
	}

	if detail == nil {
		return errors.New(
			"el detalle del incidente es obligatorio",
		)
	}

	if detail.Incident.ID <= 0 {
		return errors.New(
			"el identificador del incidente no es válido",
		)
	}

	if detail.Alert.ID <= 0 {
		return errors.New(
			"el identificador de la alerta no es válido",
		)
	}

	if newAction != nil {
		if strings.TrimSpace(
			newAction.Description,
		) == "" {
			return errors.New(
				"la descripción de la acción correctiva es obligatoria",
			)
		}

		if strings.TrimSpace(
			newAction.Responsible,
		) == "" {
			return errors.New(
				"el responsable de la acción correctiva es obligatorio",
			)
		}
	}

	tx, err := repository.db.BeginTx(
		ctx,
		nil,
	)

	if err != nil {
		return fmt.Errorf(
			"no se pudo iniciar la transacción BPM en Oracle: %w",
			err,
		)
	}

	defer func() {
		_ = tx.Rollback()
	}()

	previousIncidentStatus,
		storedAlertID,
		err := lockOracleIncident(
		ctx,
		tx,
		detail.Incident.ID,
	)

	if err != nil {
		return err
	}

	if storedAlertID != detail.Alert.ID {
		return fmt.Errorf(
			"la alerta %d no pertenece al incidente %d",
			detail.Alert.ID,
			detail.Incident.ID,
		)
	}

	previousAlertStatus, err :=
		lockOracleAlert(
			ctx,
			tx,
			detail.Alert.ID,
		)

	if err != nil {
		return err
	}

	if err := validateOracleIncidentTransition(
		previousIncidentStatus,
		detail.Incident.Status,
		newAction != nil,
	); err != nil {
		return err
	}

	if err := validateOracleAlertTransition(
		previousAlertStatus,
		detail.Alert.Status,
	); err != nil {
		return err
	}

	if err := validateOracleWorkflowConsistency(
		detail.Incident.Status,
		detail.Alert.Status,
	); err != nil {
		return err
	}

	if err := updateOracleIncident(
		ctx,
		tx,
		detail.Incident,
	); err != nil {
		return err
	}

	if err := updateOracleAlert(
		ctx,
		tx,
		detail.Alert,
	); err != nil {
		return err
	}

	now := time.Now()

	if previousIncidentStatus !=
		detail.Incident.Status {
		actor := strings.TrimSpace(
			detail.Incident.Responsible,
		)

		if actor == "" {
			actor = "sistema"
		}

		if err := insertOracleAudit(
			ctx,
			tx,
			AuditEvent{
				Actor: actor,

				Action: auditActionForIncidentStatus(
					detail.Incident.Status,
				),

				Module: "calidad",
				Entity: "incidente",
				RecordID: fmt.Sprint(
					detail.Incident.ID,
				),
				Result: "correcto",

				Detail: fmt.Sprintf(
					"Transición de %s a %s.",
					previousIncidentStatus,
					detail.Incident.Status,
				),

				OccurredAt: now,
			},
		); err != nil {
			return err
		}
	}

	if newAction != nil {
		action := *newAction

		action.IncidentID =
			detail.Incident.ID

		if action.CreatedAt.IsZero() {
			action.CreatedAt = now
		}

		actionID, err :=
			insertOracleCorrectiveAction(
				ctx,
				tx,
				action,
			)

		if err != nil {
			return err
		}

		action.ID = actionID

		if err := insertOracleAudit(
			ctx,
			tx,
			AuditEvent{
				Actor: action.Responsible,

				Action: "accion_correctiva_registrada",

				Module: "calidad",
				Entity: "accion_correctiva",

				RecordID: fmt.Sprint(
					action.ID,
				),

				Result:     "correcto",
				Detail:     action.Description,
				OccurredAt: action.CreatedAt,
			},
		); err != nil {
			return err
		}

		detail.Actions = append(
			detail.Actions,
			action,
		)

		*newAction = action
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf(
			"no se pudo confirmar la transacción BPM en Oracle: %w",
			err,
		)
	}

	return nil
}

func lockOracleIncident(
	ctx context.Context,
	tx *sql.Tx,
	incidentID int64,
) (
	string,
	int64,
	error,
) {
	const query = `
SELECT estado,
       id_alerta
FROM sig_incidente
WHERE id = :1
FOR UPDATE`

	var status string
	var alertID int64

	err := tx.QueryRowContext(
		ctx,
		query,
		incidentID,
	).Scan(
		&status,
		&alertID,
	)

	if errors.Is(
		err,
		sql.ErrNoRows,
	) {
		return "",
			0,
			fmt.Errorf(
				"el incidente %d no existe en Oracle",
				incidentID,
			)
	}

	if err != nil {
		return "",
			0,
			fmt.Errorf(
				"no se pudo bloquear el incidente %d en Oracle: %w",
				incidentID,
				err,
			)
	}

	return status,
		alertID,
		nil
}

func lockOracleAlert(
	ctx context.Context,
	tx *sql.Tx,
	alertID int64,
) (
	string,
	error,
) {
	const query = `
SELECT estado
FROM sig_alerta
WHERE id = :1
FOR UPDATE`

	var status string

	err := tx.QueryRowContext(
		ctx,
		query,
		alertID,
	).Scan(
		&status,
	)

	if errors.Is(
		err,
		sql.ErrNoRows,
	) {
		return "",
			fmt.Errorf(
				"la alerta %d no existe en Oracle",
				alertID,
			)
	}

	if err != nil {
		return "",
			fmt.Errorf(
				"no se pudo bloquear la alerta %d en Oracle: %w",
				alertID,
				err,
			)
	}

	return status,
		nil
}

func updateOracleIncident(
	ctx context.Context,
	tx *sql.Tx,
	incident Incident,
) error {
	const statement = `
UPDATE sig_incidente
SET estado = :1,
    responsable = :2,
    fecha_reconocimiento = :3,
    fecha_resolucion = :4,
    fecha_cierre = :5
WHERE id = :6`

	_, err := tx.ExecContext(
		ctx,
		statement,
		incident.Status,
		nullableString(
			incident.Responsible,
		),
		oracleNullableTime(
			incident.RecognizedAt,
		),
		oracleNullableTime(
			incident.ResolvedAt,
		),
		oracleNullableTime(
			incident.ClosedAt,
		),
		incident.ID,
	)

	if err != nil {
		return fmt.Errorf(
			"no se pudo actualizar el incidente %d en Oracle: %w",
			incident.ID,
			err,
		)
	}

	return nil
}

func updateOracleAlert(
	ctx context.Context,
	tx *sql.Tx,
	alert Alert,
) error {
	const statement = `
UPDATE sig_alerta
SET estado = :1,
    fecha_reconocimiento = :2,
    fecha_cierre = :3
WHERE id = :4`

	_, err := tx.ExecContext(
		ctx,
		statement,
		alert.Status,
		oracleNullableTime(
			alert.RecognizedAt,
		),
		oracleNullableTime(
			alert.ClosedAt,
		),
		alert.ID,
	)

	if err != nil {
		return fmt.Errorf(
			"no se pudo actualizar la alerta %d en Oracle: %w",
			alert.ID,
			err,
		)
	}

	return nil
}

func insertOracleCorrectiveAction(
	ctx context.Context,
	tx *sql.Tx,
	action CorrectiveAction,
) (
	int64,
	error,
) {
	const statement = `
INSERT INTO sig_accion_correctiva (
    id_incidente,
    descripcion,
    responsable,
    resultado,
    fecha_accion
) VALUES (
    :1,
    :2,
    :3,
    :4,
    :5
)
RETURNING id INTO :6`

	var actionID int64

	_, err := tx.ExecContext(
		ctx,
		statement,
		action.IncidentID,
		action.Description,
		action.Responsible,
		nullableString(
			action.Result,
		),
		action.CreatedAt,
		sql.Out{
			Dest: &actionID,
		},
	)

	if err != nil {
		return 0,
			fmt.Errorf(
				"no se pudo guardar la acción correctiva en Oracle: %w",
				err,
			)
	}

	return actionID,
		nil
}

func validateOracleIncidentTransition(
	currentStatus string,
	targetStatus string,
	hasNewAction bool,
) error {
	if currentStatus == "en_tratamiento" &&
		targetStatus == "en_tratamiento" &&
		hasNewAction {
		return nil
	}

	validTransition := false

	switch currentStatus {
	case "abierto":
		validTransition =
			targetStatus == "reconocido"

	case "reconocido":
		validTransition =
			targetStatus == "en_tratamiento"

	case "en_tratamiento":
		validTransition =
			targetStatus == "resuelto"

	case "resuelto":
		validTransition =
			targetStatus == "cerrado"
	}

	if !validTransition {
		return fmt.Errorf(
			"la transición del incidente de %s a %s no está permitida",
			currentStatus,
			targetStatus,
		)
	}

	return nil
}

func validateOracleAlertTransition(
	currentStatus string,
	targetStatus string,
) error {
	validTransition :=
		(currentStatus == "abierta" &&
			targetStatus == "reconocida") ||
			(currentStatus == "reconocida" &&
				targetStatus == "reconocida") ||
			(currentStatus == "reconocida" &&
				targetStatus == "cerrada")

	if !validTransition {
		return fmt.Errorf(
			"la transición de la alerta de %s a %s no está permitida",
			currentStatus,
			targetStatus,
		)
	}

	return nil
}

func validateOracleWorkflowConsistency(
	incidentStatus string,
	alertStatus string,
) error {
	switch incidentStatus {
	case "reconocido",
		"en_tratamiento",
		"resuelto":
		if alertStatus != "reconocida" {
			return fmt.Errorf(
				"el incidente %s requiere una alerta reconocida",
				incidentStatus,
			)
		}

	case "cerrado":
		if alertStatus != "cerrada" {
			return errors.New(
				"un incidente cerrado requiere una alerta cerrada",
			)
		}

	default:
		return fmt.Errorf(
			"el estado objetivo %q no es válido para el flujo BPM",
			incidentStatus,
		)
	}

	return nil
}

func oracleNullableTime(
	value *time.Time,
) any {
	if value == nil {
		return nil
	}

	return *value
}
