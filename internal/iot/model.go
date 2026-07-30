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
	ID           int64      `json:"id"`
	ReadingID    int64      `json:"idLectura"`
	Type         string     `json:"tipo"`
	Severity     string     `json:"severidad"`
	Message      string     `json:"mensaje"`
	Status       string     `json:"estado"`
	CreatedAt    time.Time  `json:"fechaApertura"`
	RecognizedAt *time.Time `json:"fechaReconocimiento,omitempty"`
	ClosedAt     *time.Time `json:"fechaCierre,omitempty"`
}

// Incident representa una no conformidad de calidad.
type Incident struct {
	ID           int64      `json:"id"`
	AlertID      int64      `json:"idAlerta"`
	Code         string     `json:"codigo"`
	Title        string     `json:"titulo"`
	Description  string     `json:"descripcion"`
	Severity     string     `json:"severidad"`
	Status       string     `json:"estado"`
	Responsible  string     `json:"responsable,omitempty"`
	CreatedAt    time.Time  `json:"fechaApertura"`
	RecognizedAt *time.Time `json:"fechaReconocimiento,omitempty"`
	ResolvedAt   *time.Time `json:"fechaResolucion,omitempty"`
	ClosedAt     *time.Time `json:"fechaCierre,omitempty"`
}

// CorrectiveAction representa una acción aplicada a un incidente.
type CorrectiveAction struct {
	ID          int64     `json:"id"`
	IncidentID  int64     `json:"idIncidente"`
	Description string    `json:"descripcion"`
	Responsible string    `json:"responsable"`
	Result      string    `json:"resultado,omitempty"`
	CreatedAt   time.Time `json:"fechaAccion"`
}

// RegistrationResult contiene el resultado del proceso automático.
type RegistrationResult struct {
	Reading  Reading   `json:"lectura"`
	Alert    *Alert    `json:"alerta,omitempty"`
	Incident *Incident `json:"incidente,omitempty"`
}

// IncidentDetail agrupa el incidente, la alerta y sus acciones.
type IncidentDetail struct {
	Incident Incident           `json:"incidente"`
	Alert    Alert              `json:"alerta"`
	Actions  []CorrectiveAction `json:"acciones"`
}

// AuditEvent representa una acción trazable realizada en el SIG.
type AuditEvent struct {
	ID         int64     `json:"id"`
	Actor      string    `json:"actor"`
	Action     string    `json:"accion"`
	Module     string    `json:"modulo"`
	Entity     string    `json:"entidad"`
	RecordID   string    `json:"idRegistro"`
	Result     string    `json:"resultado"`
	Detail     string    `json:"detalle,omitempty"`
	OccurredAt time.Time `json:"fechaEvento"`
}

// Recommendation representa una recomendación gerencial calculada.
type Recommendation struct {
	Code     string `json:"codigo"`
	Priority string `json:"prioridad"`
	Title    string `json:"titulo"`
	Message  string `json:"mensaje"`
}

// AlertCounts resume las alertas por estado.
type AlertCounts struct {
	Open       int `json:"abiertas"`
	Recognized int `json:"reconocidas"`
	Closed     int `json:"cerradas"`
}

// IncidentCounts resume los incidentes por estado.
type IncidentCounts struct {
	Open        int `json:"abiertos"`
	Recognized  int `json:"reconocidos"`
	InTreatment int `json:"enTratamiento"`
	Resolved    int `json:"resueltos"`
	Closed      int `json:"cerrados"`
}

// IoTSummary contiene los indicadores gerenciales de cadena de frío.
type IoTSummary struct {
	DeviceCode         string           `json:"codigoDispositivo"`
	LatestReading      *Reading         `json:"ultimaLectura,omitempty"`
	TotalReadings      int              `json:"totalLecturas"`
	NormalReadings     int              `json:"lecturasNormales"`
	OutOfRangeReadings int              `json:"lecturasFueraRango"`
	NormalPercentage   float64          `json:"porcentajeNormal"`
	Alerts             AlertCounts      `json:"alertas"`
	Incidents          IncidentCounts   `json:"incidentes"`
	Recommendations    []Recommendation `json:"recomendaciones"`
	UpdatedAt          time.Time        `json:"fechaActualizacion"`
}
