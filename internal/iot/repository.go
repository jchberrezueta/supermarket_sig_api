package iot

import "context"

// Repository define el almacenamiento requerido por el proceso IoT.
type Repository interface {
	Save(
		ctx context.Context,
		result *RegistrationResult,
	) error

	Latest(
		ctx context.Context,
		deviceCode string,
	) (
		Reading,
		bool,
		error,
	)

	ListReadings(
		ctx context.Context,
		deviceCode string,
		limit int,
	) (
		[]Reading,
		error,
	)

	ListAlerts(
		ctx context.Context,
		status string,
		limit int,
	) (
		[]Alert,
		error,
	)

	ListIncidents(
		ctx context.Context,
		status string,
		limit int,
	) (
		[]Incident,
		error,
	)

	FindIncident(
		ctx context.Context,
		incidentID int64,
	) (
		IncidentDetail,
		bool,
		error,
	)

	ApplyIncidentWorkflow(
		ctx context.Context,
		detail *IncidentDetail,
		newAction *CorrectiveAction,
	) error

	ListAudit(
		ctx context.Context,
		action string,
		limit int,
	) (
		[]AuditEvent,
		error,
	)

	Summary(
		ctx context.Context,
		deviceCode string,
	) (
		IoTSummary,
		error,
	)
}
