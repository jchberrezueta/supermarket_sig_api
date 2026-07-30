package iot

import "time"

const (
	ReadingStatusNormal     = "normal"
	ReadingStatusOutOfRange = "fuera_rango"
)

// RegisterInput contiene una lectura recibida desde un dispositivo.
type RegisterInput struct {
	DeviceCode  string
	Temperature float64
	Humidity    *float64
	ReadAt      time.Time
}

// Reading representa una lectura registrada.
type Reading struct {
	ID          int64     `json:"id"`
	DeviceCode  string    `json:"codigoDispositivo"`
	Temperature float64   `json:"temperatura"`
	Humidity    *float64  `json:"humedad,omitempty"`
	Status      string    `json:"estado"`
	ReadAt      time.Time `json:"fechaLectura"`
	CreatedAt   time.Time `json:"fechaRegistro"`
}

// Alert representa una alerta generada automáticamente.
type Alert struct {
	ID        int64     `json:"id"`
	ReadingID int64     `json:"idLectura"`
	Type      string    `json:"tipo"`
	Severity  string    `json:"severidad"`
	Message   string    `json:"mensaje"`
	Status    string    `json:"estado"`
	CreatedAt time.Time `json:"fechaApertura"`
}

// Incident representa una no conformidad de calidad.
type Incident struct {
	ID          int64     `json:"id"`
	AlertID     int64     `json:"idAlerta"`
	Code        string    `json:"codigo"`
	Title       string    `json:"titulo"`
	Description string    `json:"descripcion"`
	Severity    string    `json:"severidad"`
	Status      string    `json:"estado"`
	CreatedAt   time.Time `json:"fechaApertura"`
}

// RegistrationResult contiene el resultado del proceso automático.
type RegistrationResult struct {
	Reading  Reading   `json:"lectura"`
	Alert    *Alert    `json:"alerta,omitempty"`
	Incident *Incident `json:"incidente,omitempty"`
}

// IncidentDetail agrupa el incidente con la alerta que lo originó.
type IncidentDetail struct {
	Incident Incident `json:"incidente"`
	Alert    Alert    `json:"alerta"`
}
