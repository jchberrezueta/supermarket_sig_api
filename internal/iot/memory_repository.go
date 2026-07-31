package iot

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"
)

// MemoryRepository almacena temporalmente los datos en memoria.
// Más adelante será reemplazado por OracleRepository.
type MemoryRepository struct {
	mu sync.RWMutex

	nextReadingID  int64
	nextAlertID    int64
	nextIncidentID int64
	nextActionID   int64
	nextAuditID    int64

	readings []Reading

	alerts     map[int64]Alert
	alertOrder []int64

	incidents     map[int64]Incident
	incidentOrder []int64

	actions map[int64][]CorrectiveAction
	latest  map[string]Reading
	audit   []AuditEvent
}

// NewMemoryRepository crea un repositorio local vacío.
func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		nextReadingID:  1,
		nextAlertID:    1,
		nextIncidentID: 1,
		nextActionID:   1,
		nextAuditID:    1,

		readings: make(
			[]Reading,
			0,
		),

		alerts: make(
			map[int64]Alert,
		),

		alertOrder: make(
			[]int64,
			0,
		),

		incidents: make(
			map[int64]Incident,
		),

		incidentOrder: make(
			[]int64,
			0,
		),

		actions: make(
			map[int64][]CorrectiveAction,
		),

		latest: make(
			map[string]Reading,
		),

		audit: make(
			[]AuditEvent,
			0,
		),
	}
}

// Save asigna identificadores y guarda el resultado.
func (repository *MemoryRepository) Save(
	_ context.Context,
	result *RegistrationResult,
) error {
	repository.mu.Lock()
	defer repository.mu.Unlock()

	now := time.Now()

	result.Reading.ID = repository.nextReadingID
	result.Reading.CreatedAt = now

	repository.nextReadingID++

	repository.readings = append(
		repository.readings,
		result.Reading,
	)

	repository.latest[result.Reading.DeviceCode] =
		result.Reading

	repository.appendAuditLocked(
		AuditEvent{
			Actor:    result.Reading.DeviceCode,
			Action:   "lectura_registrada",
			Module:   "iot",
			Entity:   "lectura_sensor",
			RecordID: fmt.Sprint(result.Reading.ID),
			Result:   "correcto",
			Detail: fmt.Sprintf(
				"Temperatura %.2f °C, estado %s.",
				result.Reading.Temperature,
				result.Reading.Status,
			),
			OccurredAt: now,
		},
	)

	if result.Alert != nil &&
		result.Incident != nil &&
		repository.hasActiveIncidentForDeviceLocked(
			result.Reading.DeviceCode,
		) {
		result.Alert = nil
		result.Incident = nil
	}

	if result.Alert != nil {
		result.Alert.ID = repository.nextAlertID
		result.Alert.ReadingID = result.Reading.ID
		result.Alert.CreatedAt = now

		repository.nextAlertID++

		repository.alerts[result.Alert.ID] =
			*result.Alert

		repository.alertOrder = append(
			repository.alertOrder,
			result.Alert.ID,
		)

		repository.appendAuditLocked(
			AuditEvent{
				Actor:      "sistema",
				Action:     "alerta_creada",
				Module:     "cadena_frio",
				Entity:     "alerta",
				RecordID:   fmt.Sprint(result.Alert.ID),
				Result:     "correcto",
				Detail:     result.Alert.Message,
				OccurredAt: now,
			},
		)
	}

	if result.Incident != nil {
		result.Incident.ID = repository.nextIncidentID
		result.Incident.CreatedAt = now

		if result.Alert != nil {
			result.Incident.AlertID =
				result.Alert.ID
		}

		repository.nextIncidentID++

		repository.incidents[result.Incident.ID] =
			*result.Incident

		repository.incidentOrder = append(
			repository.incidentOrder,
			result.Incident.ID,
		)

		repository.actions[result.Incident.ID] =
			make([]CorrectiveAction, 0)

		repository.appendAuditLocked(
			AuditEvent{
				Actor:      "sistema",
				Action:     "incidente_creado",
				Module:     "calidad",
				Entity:     "incidente",
				RecordID:   fmt.Sprint(result.Incident.ID),
				Result:     "correcto",
				Detail:     result.Incident.Description,
				OccurredAt: now,
			},
		)
	}

	return nil
}

func (
	repository *MemoryRepository,
) hasActiveIncidentForDeviceLocked(
	deviceCode string,
) bool {
	for index :=
		len(repository.incidentOrder) - 1; index >= 0; index-- {
		incidentID :=
			repository.incidentOrder[index]

		incident, exists :=
			repository.incidents[incidentID]

		if !exists ||
			incident.Status == "cerrado" {
			continue
		}

		alert, exists :=
			repository.alerts[incident.AlertID]

		if !exists {
			continue
		}

		for readingIndex :=
			len(repository.readings) - 1; readingIndex >= 0; readingIndex-- {
			reading :=
				repository.readings[readingIndex]

			if reading.ID != alert.ReadingID {
				continue
			}

			if reading.DeviceCode == deviceCode {
				return true
			}

			break
		}
	}

	return false
}

// Latest obtiene la última lectura de un dispositivo.
func (repository *MemoryRepository) Latest(
	_ context.Context,
	deviceCode string,
) (
	Reading,
	bool,
	error,
) {
	repository.mu.RLock()
	defer repository.mu.RUnlock()

	reading, exists :=
		repository.latest[deviceCode]

	return reading, exists, nil
}

// ListReadings devuelve las lecturas más recientes primero.
func (repository *MemoryRepository) ListReadings(
	_ context.Context,
	deviceCode string,
	limit int,
) (
	[]Reading,
	error,
) {
	repository.mu.RLock()
	defer repository.mu.RUnlock()

	result := make(
		[]Reading,
		0,
	)

	for index := len(repository.readings) - 1; index >= 0 && len(result) < limit; index-- {
		reading := repository.readings[index]

		if reading.DeviceCode != deviceCode {
			continue
		}

		result = append(
			result,
			reading,
		)
	}

	return result, nil
}

// ListAlerts devuelve las alertas más recientes primero.
func (repository *MemoryRepository) ListAlerts(
	_ context.Context,
	status string,
	limit int,
) (
	[]Alert,
	error,
) {
	repository.mu.RLock()
	defer repository.mu.RUnlock()

	result := make(
		[]Alert,
		0,
	)

	for index := len(repository.alertOrder) - 1; index >= 0 && len(result) < limit; index-- {
		alertID := repository.alertOrder[index]
		alert := repository.alerts[alertID]

		if status != "" &&
			alert.Status != status {
			continue
		}

		result = append(
			result,
			alert,
		)
	}

	return result, nil
}

// ListIncidents devuelve los incidentes más recientes primero.
func (repository *MemoryRepository) ListIncidents(
	_ context.Context,
	status string,
	limit int,
) (
	[]Incident,
	error,
) {
	repository.mu.RLock()
	defer repository.mu.RUnlock()

	result := make(
		[]Incident,
		0,
	)

	for index := len(repository.incidentOrder) - 1; index >= 0 && len(result) < limit; index-- {
		incidentID :=
			repository.incidentOrder[index]

		incident :=
			repository.incidents[incidentID]

		if status != "" &&
			incident.Status != status {
			continue
		}

		result = append(
			result,
			incident,
		)
	}

	return result, nil
}

// FindIncident devuelve un incidente, su alerta y acciones.
func (repository *MemoryRepository) FindIncident(
	_ context.Context,
	incidentID int64,
) (
	IncidentDetail,
	bool,
	error,
) {
	repository.mu.RLock()
	defer repository.mu.RUnlock()

	incident, exists :=
		repository.incidents[incidentID]

	if !exists {
		return IncidentDetail{},
			false,
			nil
	}

	alert, alertExists :=
		repository.alerts[incident.AlertID]

	if !alertExists {
		return IncidentDetail{},
			false,
			nil
	}

	storedActions :=
		repository.actions[incidentID]

	actions := make(
		[]CorrectiveAction,
		len(storedActions),
	)

	copy(
		actions,
		storedActions,
	)

	return IncidentDetail{
			Incident: incident,
			Alert:    alert,
			Actions:  actions,
		},
		true,
		nil
}

// ApplyIncidentWorkflow guarda una transición BPM y,
// cuando corresponde, una nueva acción correctiva.
func (repository *MemoryRepository) ApplyIncidentWorkflow(
	_ context.Context,
	detail *IncidentDetail,
	newAction *CorrectiveAction,
) error {
	repository.mu.Lock()
	defer repository.mu.Unlock()

	previousIncident, exists :=
		repository.incidents[detail.Incident.ID]

	if !exists {
		return fmt.Errorf(
			"el incidente %d no existe",
			detail.Incident.ID,
		)
	}

	if _, exists :=
		repository.alerts[detail.Alert.ID]; !exists {
		return fmt.Errorf(
			"la alerta %d no existe",
			detail.Alert.ID,
		)
	}

	repository.incidents[detail.Incident.ID] =
		detail.Incident

	repository.alerts[detail.Alert.ID] =
		detail.Alert

	if previousIncident.Status !=
		detail.Incident.Status {
		actor := strings.TrimSpace(
			detail.Incident.Responsible,
		)

		if actor == "" {
			actor = "sistema"
		}

		repository.appendAuditLocked(
			AuditEvent{
				Actor: actor,
				Action: auditActionForIncidentStatus(
					detail.Incident.Status,
				),
				Module:   "calidad",
				Entity:   "incidente",
				RecordID: fmt.Sprint(detail.Incident.ID),
				Result:   "correcto",
				Detail: fmt.Sprintf(
					"Transición de %s a %s.",
					previousIncident.Status,
					detail.Incident.Status,
				),
				OccurredAt: time.Now(),
			},
		)
	}

	if newAction != nil {
		newAction.ID = repository.nextActionID
		newAction.IncidentID =
			detail.Incident.ID

		if newAction.CreatedAt.IsZero() {
			newAction.CreatedAt = time.Now()
		}

		repository.nextActionID++

		repository.actions[detail.Incident.ID] =
			append(
				repository.actions[detail.Incident.ID],
				*newAction,
			)

		repository.appendAuditLocked(
			AuditEvent{
				Actor:      newAction.Responsible,
				Action:     "accion_correctiva_registrada",
				Module:     "calidad",
				Entity:     "accion_correctiva",
				RecordID:   fmt.Sprint(newAction.ID),
				Result:     "correcto",
				Detail:     newAction.Description,
				OccurredAt: newAction.CreatedAt,
			},
		)
	}

	storedActions :=
		repository.actions[detail.Incident.ID]

	detail.Actions = make(
		[]CorrectiveAction,
		len(storedActions),
	)

	copy(
		detail.Actions,
		storedActions,
	)

	return nil
}

// ListAudit devuelve los eventos de auditoría más recientes.
func (repository *MemoryRepository) ListAudit(
	_ context.Context,
	action string,
	limit int,
) (
	[]AuditEvent,
	error,
) {
	repository.mu.RLock()
	defer repository.mu.RUnlock()

	action = strings.TrimSpace(
		action,
	)

	result := make(
		[]AuditEvent,
		0,
	)

	for index := len(repository.audit) - 1; index >= 0 && len(result) < limit; index-- {
		event := repository.audit[index]

		if action != "" &&
			event.Action != action {
			continue
		}

		result = append(
			result,
			event,
		)
	}

	return result, nil
}

func (repository *MemoryRepository) appendAuditLocked(
	event AuditEvent,
) {
	event.ID = repository.nextAuditID

	if event.OccurredAt.IsZero() {
		event.OccurredAt = time.Now()
	}

	repository.nextAuditID++

	repository.audit = append(
		repository.audit,
		event,
	)
}

func auditActionForIncidentStatus(
	status string,
) string {
	switch status {
	case "reconocido":
		return "incidente_reconocido"

	case "en_tratamiento":
		return "incidente_en_tratamiento"

	case "resuelto":
		return "incidente_resuelto"

	case "cerrado":
		return "incidente_cerrado"

	default:
		return "incidente_actualizado"
	}
}

// Summary calcula los indicadores del dispositivo.
func (repository *MemoryRepository) Summary(
	_ context.Context,
	deviceCode string,
) (
	IoTSummary,
	error,
) {
	repository.mu.RLock()
	defer repository.mu.RUnlock()

	summary := IoTSummary{
		DeviceCode: deviceCode,

		Recommendations: make(
			[]Recommendation,
			0,
		),

		UpdatedAt: time.Now(),
	}

	readingIDs := make(
		map[int64]struct{},
	)

	for _, reading := range repository.readings {
		if reading.DeviceCode != deviceCode {
			continue
		}

		readingIDs[reading.ID] =
			struct{}{}

		summary.TotalReadings++

		switch reading.Status {
		case ReadingStatusNormal:
			summary.NormalReadings++

		case ReadingStatusOutOfRange:
			summary.OutOfRangeReadings++
		}
	}

	if latestReading, exists :=
		repository.latest[deviceCode]; exists {
		readingCopy := latestReading

		summary.LatestReading =
			&readingCopy
	}

	alertIDs := make(
		map[int64]struct{},
	)

	for _, alert := range repository.alerts {
		if _, exists :=
			readingIDs[alert.ReadingID]; !exists {
			continue
		}

		alertIDs[alert.ID] =
			struct{}{}

		switch alert.Status {
		case "abierta":
			summary.Alerts.Open++

		case "reconocida":
			summary.Alerts.Recognized++

		case "cerrada":
			summary.Alerts.Closed++
		}
	}

	for _, incident := range repository.incidents {
		if _, exists :=
			alertIDs[incident.AlertID]; !exists {
			continue
		}

		switch incident.Status {
		case "abierto":
			summary.Incidents.Open++

		case "reconocido":
			summary.Incidents.Recognized++

		case "en_tratamiento":
			summary.Incidents.InTreatment++

		case "resuelto":
			summary.Incidents.Resolved++

		case "cerrado":
			summary.Incidents.Closed++
		}
	}

	if summary.TotalReadings > 0 {
		summary.NormalPercentage =
			float64(summary.NormalReadings) *
				100 /
				float64(summary.TotalReadings)
	}

	return summary, nil
}
