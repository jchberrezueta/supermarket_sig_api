package iot

import (
	"context"
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

	readings []Reading

	alerts     map[int64]Alert
	alertOrder []int64

	incidents     map[int64]Incident
	incidentOrder []int64

	latest map[string]Reading
}

// NewMemoryRepository crea un repositorio local vacío.
func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		nextReadingID:  1,
		nextAlertID:    1,
		nextIncidentID: 1,

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

		latest: make(
			map[string]Reading,
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
	}

	return nil
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

// FindIncident devuelve un incidente y la alerta relacionada.
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

	return IncidentDetail{
			Incident: incident,
			Alert:    alert,
		},
		true,
		nil
}
