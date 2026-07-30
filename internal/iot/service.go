package iot

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"
)

// ValidationError representa una lectura inválida.
type ValidationError struct {
	Message string
}

func (err *ValidationError) Error() string {
	return err.Message
}

// Service contiene las reglas del proceso IoT.
type Service struct {
	repository     Repository
	deviceCode     string
	temperatureMin float64
	temperatureMax float64
}

// NewService crea el servicio del proceso IoT.
func NewService(
	repository Repository,
	deviceCode string,
	temperatureMin float64,
	temperatureMax float64,
) *Service {
	deviceCode = strings.TrimSpace(deviceCode)

	if deviceCode == "" {
		deviceCode = "ESP32-BODEGA-01"
	}

	if temperatureMin >= temperatureMax {
		temperatureMin = 2
		temperatureMax = 8
	}

	return &Service{
		repository:     repository,
		deviceCode:     deviceCode,
		temperatureMin: temperatureMin,
		temperatureMax: temperatureMax,
	}
}

// DefaultDeviceCode devuelve el dispositivo configurado.
func (service *Service) DefaultDeviceCode() string {
	return service.deviceCode
}

// Register ejecuta el proceso automatizado de cadena de frío.
func (service *Service) Register(
	ctx context.Context,
	input RegisterInput,
) (
	RegistrationResult,
	error,
) {
	input.DeviceCode = strings.TrimSpace(
		input.DeviceCode,
	)

	if input.DeviceCode == "" {
		return RegistrationResult{},
			&ValidationError{
				Message: "El código del dispositivo es obligatorio.",
			}
	}

	if input.DeviceCode != service.deviceCode {
		return RegistrationResult{},
			&ValidationError{
				Message: "El dispositivo indicado no está registrado.",
			}
	}

	if math.IsNaN(input.Temperature) ||
		math.IsInf(input.Temperature, 0) ||
		input.Temperature < -50 ||
		input.Temperature > 100 {
		return RegistrationResult{},
			&ValidationError{
				Message: "La temperatura recibida no es válida.",
			}
	}

	if input.Humidity != nil {
		if math.IsNaN(*input.Humidity) ||
			math.IsInf(*input.Humidity, 0) ||
			*input.Humidity < 0 ||
			*input.Humidity > 100 {
			return RegistrationResult{},
				&ValidationError{
					Message: "La humedad debe estar entre 0 y 100.",
				}
		}
	}

	if input.ReadAt.IsZero() {
		input.ReadAt = time.Now()
	}

	if input.ReadAt.After(
		time.Now().Add(5 * time.Minute),
	) {
		return RegistrationResult{},
			&ValidationError{
				Message: "La fecha de lectura no puede estar en el futuro.",
			}
	}

	result := RegistrationResult{
		Reading: Reading{
			DeviceCode:  input.DeviceCode,
			Temperature: input.Temperature,
			Humidity:    input.Humidity,
			Status:      ReadingStatusNormal,
			ReadAt:      input.ReadAt,
		},
	}

	if input.Temperature < service.temperatureMin ||
		input.Temperature > service.temperatureMax {
		severity := service.calculateSeverity(
			input.Temperature,
		)

		message := fmt.Sprintf(
			"Temperatura %.2f °C fuera del rango permitido de %.2f °C a %.2f °C.",
			input.Temperature,
			service.temperatureMin,
			service.temperatureMax,
		)

		result.Reading.Status =
			ReadingStatusOutOfRange

		result.Alert = &Alert{
			Type:     "temperatura_fuera_rango",
			Severity: severity,
			Message:  message,
			Status:   "abierta",
		}

		result.Incident = &Incident{
			Code: fmt.Sprintf(
				"INC-TEMP-%d",
				time.Now().UnixNano(),
			),
			Title:       "Incidente de cadena de frío",
			Description: message,
			Severity:    severity,
			Status:      "abierto",
		}
	}

	if err := service.repository.Save(
		ctx,
		&result,
	); err != nil {
		return RegistrationResult{},
			fmt.Errorf(
				"no se pudo guardar la lectura: %w",
				err,
			)
	}

	return result, nil
}

// Latest obtiene la última lectura registrada.
func (service *Service) Latest(
	ctx context.Context,
	deviceCode string,
) (
	Reading,
	bool,
	error,
) {
	deviceCode = strings.TrimSpace(deviceCode)

	if deviceCode == "" {
		deviceCode = service.deviceCode
	}

	if deviceCode != service.deviceCode {
		return Reading{},
			false,
			&ValidationError{
				Message: "El dispositivo indicado no está registrado.",
			}
	}

	return service.repository.Latest(
		ctx,
		deviceCode,
	)
}

func (service *Service) calculateSeverity(
	temperature float64,
) string {
	var difference float64

	if temperature < service.temperatureMin {
		difference =
			service.temperatureMin - temperature
	} else {
		difference =
			temperature - service.temperatureMax
	}

	switch {
	case difference > 5:
		return "critica"

	case difference > 2:
		return "alta"

	default:
		return "media"
	}
}

// NotFoundError representa un recurso que no existe.
type NotFoundError struct {
	Message string
}

func (err *NotFoundError) Error() string {
	return err.Message
}

// ListReadings devuelve el historial reciente del dispositivo.
func (service *Service) ListReadings(
	ctx context.Context,
	deviceCode string,
	limit int,
) (
	[]Reading,
	error,
) {
	deviceCode = strings.TrimSpace(
		deviceCode,
	)

	if deviceCode == "" {
		deviceCode = service.deviceCode
	}

	if deviceCode != service.deviceCode {
		return nil,
			&ValidationError{
				Message: "El dispositivo indicado no está registrado.",
			}
	}

	limit, err := validateLimit(limit)

	if err != nil {
		return nil, err
	}

	return service.repository.ListReadings(
		ctx,
		deviceCode,
		limit,
	)
}

// ListAlerts devuelve alertas filtradas por estado.
func (service *Service) ListAlerts(
	ctx context.Context,
	status string,
	limit int,
) (
	[]Alert,
	error,
) {
	status = strings.ToLower(
		strings.TrimSpace(
			status,
		),
	)

	validStatuses := map[string]bool{
		"":           true,
		"abierta":    true,
		"reconocida": true,
		"cerrada":    true,
	}

	if !validStatuses[status] {
		return nil,
			&ValidationError{
				Message: "El estado de alerta no es válido.",
			}
	}

	limit, err := validateLimit(limit)

	if err != nil {
		return nil, err
	}

	return service.repository.ListAlerts(
		ctx,
		status,
		limit,
	)
}

// ListIncidents devuelve incidentes filtrados por estado.
func (service *Service) ListIncidents(
	ctx context.Context,
	status string,
	limit int,
) (
	[]Incident,
	error,
) {
	status = strings.ToLower(
		strings.TrimSpace(
			status,
		),
	)

	validStatuses := map[string]bool{
		"":               true,
		"abierto":        true,
		"reconocido":     true,
		"en_tratamiento": true,
		"resuelto":       true,
		"cerrado":        true,
	}

	if !validStatuses[status] {
		return nil,
			&ValidationError{
				Message: "El estado del incidente no es válido.",
			}
	}

	limit, err := validateLimit(limit)

	if err != nil {
		return nil, err
	}

	return service.repository.ListIncidents(
		ctx,
		status,
		limit,
	)
}

// IncidentDetail devuelve un incidente específico.
func (service *Service) IncidentDetail(
	ctx context.Context,
	incidentID int64,
) (
	IncidentDetail,
	error,
) {
	if incidentID <= 0 {
		return IncidentDetail{},
			&ValidationError{
				Message: "El identificador del incidente no es válido.",
			}
	}

	detail, exists, err :=
		service.repository.FindIncident(
			ctx,
			incidentID,
		)

	if err != nil {
		return IncidentDetail{},
			fmt.Errorf(
				"no se pudo consultar el incidente: %w",
				err,
			)
	}

	if !exists {
		return IncidentDetail{},
			&NotFoundError{
				Message: "El incidente solicitado no existe.",
			}
	}

	return detail, nil
}

func validateLimit(
	limit int,
) (
	int,
	error,
) {
	if limit == 0 {
		return 50, nil
	}

	if limit < 1 || limit > 200 {
		return 0,
			&ValidationError{
				Message: "El límite debe estar entre 1 y 200.",
			}
	}

	return limit, nil
}

// ConflictError representa una transición BPM no permitida.
type ConflictError struct {
	Message string
}

func (err *ConflictError) Error() string {
	return err.Message
}

// RecognizeIncident reconoce formalmente un incidente abierto.
func (service *Service) RecognizeIncident(
	ctx context.Context,
	incidentID int64,
	responsible string,
) (
	IncidentDetail,
	error,
) {
	responsible = strings.TrimSpace(
		responsible,
	)

	if len(responsible) < 3 ||
		len(responsible) > 120 {
		return IncidentDetail{},
			&ValidationError{
				Message: "El responsable debe tener entre 3 y 120 caracteres.",
			}
	}

	detail, err :=
		service.IncidentDetail(
			ctx,
			incidentID,
		)

	if err != nil {
		return IncidentDetail{}, err
	}

	if detail.Incident.Status != "abierto" {
		return IncidentDetail{},
			&ConflictError{
				Message: "Solo se puede reconocer un incidente abierto.",
			}
	}

	now := time.Now()

	detail.Incident.Status = "reconocido"
	detail.Incident.Responsible = responsible
	detail.Incident.RecognizedAt = &now

	detail.Alert.Status = "reconocida"
	detail.Alert.RecognizedAt = &now

	if err := service.repository.ApplyIncidentWorkflow(
		ctx,
		&detail,
		nil,
	); err != nil {
		return IncidentDetail{},
			fmt.Errorf(
				"no se pudo reconocer el incidente: %w",
				err,
			)
	}

	return detail, nil
}

// AddCorrectiveAction registra una acción correctiva.
func (service *Service) AddCorrectiveAction(
	ctx context.Context,
	incidentID int64,
	description string,
	responsible string,
	result string,
) (
	IncidentDetail,
	error,
) {
	description = strings.TrimSpace(
		description,
	)

	responsible = strings.TrimSpace(
		responsible,
	)

	result = strings.TrimSpace(
		result,
	)

	if len(description) < 10 ||
		len(description) > 1000 {
		return IncidentDetail{},
			&ValidationError{
				Message: "La descripción debe tener entre 10 y 1000 caracteres.",
			}
	}

	if len(responsible) < 3 ||
		len(responsible) > 120 {
		return IncidentDetail{},
			&ValidationError{
				Message: "El responsable debe tener entre 3 y 120 caracteres.",
			}
	}

	if len(result) > 1000 {
		return IncidentDetail{},
			&ValidationError{
				Message: "El resultado no puede superar los 1000 caracteres.",
			}
	}

	detail, err :=
		service.IncidentDetail(
			ctx,
			incidentID,
		)

	if err != nil {
		return IncidentDetail{}, err
	}

	if detail.Incident.Status != "reconocido" &&
		detail.Incident.Status != "en_tratamiento" {
		return IncidentDetail{},
			&ConflictError{
				Message: "El incidente debe estar reconocido o en tratamiento.",
			}
	}

	action := &CorrectiveAction{
		Description: description,
		Responsible: responsible,
		Result:      result,
		CreatedAt:   time.Now(),
	}

	detail.Incident.Status =
		"en_tratamiento"

	if err := service.repository.ApplyIncidentWorkflow(
		ctx,
		&detail,
		action,
	); err != nil {
		return IncidentDetail{},
			fmt.Errorf(
				"no se pudo registrar la acción correctiva: %w",
				err,
			)
	}

	return detail, nil
}

// ResolveIncident marca como resuelto un incidente tratado.
func (service *Service) ResolveIncident(
	ctx context.Context,
	incidentID int64,
) (
	IncidentDetail,
	error,
) {
	detail, err :=
		service.IncidentDetail(
			ctx,
			incidentID,
		)

	if err != nil {
		return IncidentDetail{}, err
	}

	if detail.Incident.Status != "en_tratamiento" {
		return IncidentDetail{},
			&ConflictError{
				Message: "Solo se puede resolver un incidente en tratamiento.",
			}
	}

	if len(detail.Actions) == 0 {
		return IncidentDetail{},
			&ConflictError{
				Message: "Debe existir al menos una acción correctiva.",
			}
	}

	now := time.Now()

	detail.Incident.Status = "resuelto"
	detail.Incident.ResolvedAt = &now

	if err := service.repository.ApplyIncidentWorkflow(
		ctx,
		&detail,
		nil,
	); err != nil {
		return IncidentDetail{},
			fmt.Errorf(
				"no se pudo resolver el incidente: %w",
				err,
			)
	}

	return detail, nil
}

// CloseIncident cierra un incidente previamente resuelto.
func (service *Service) CloseIncident(
	ctx context.Context,
	incidentID int64,
) (
	IncidentDetail,
	error,
) {
	detail, err :=
		service.IncidentDetail(
			ctx,
			incidentID,
		)

	if err != nil {
		return IncidentDetail{}, err
	}

	if detail.Incident.Status != "resuelto" {
		return IncidentDetail{},
			&ConflictError{
				Message: "Solo se puede cerrar un incidente resuelto.",
			}
	}

	now := time.Now()

	detail.Incident.Status = "cerrado"
	detail.Incident.ClosedAt = &now

	detail.Alert.Status = "cerrada"
	detail.Alert.ClosedAt = &now

	if err := service.repository.ApplyIncidentWorkflow(
		ctx,
		&detail,
		nil,
	); err != nil {
		return IncidentDetail{},
			fmt.Errorf(
				"no se pudo cerrar el incidente: %w",
				err,
			)
	}

	return detail, nil
}
