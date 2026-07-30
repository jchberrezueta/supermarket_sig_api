package management

import (
	"time"

	"supermarket-sig-api/internal/erpdata"
)

// ChannelSummary resume las ventas de un canal.
type ChannelSummary struct {
	Transactions int     `json:"transacciones"`
	Total        float64 `json:"total"`
}

// SalesSummary contiene los principales indicadores comerciales.
type SalesSummary struct {
	CompletedSales int            `json:"ventasCompletadas"`
	CancelledSales int            `json:"ventasCanceladas"`
	Total          float64        `json:"totalVentas"`
	AverageTicket  float64        `json:"ticketPromedio"`
	UnitsSold      int64          `json:"unidadesVendidas"`
	POS            ChannelSummary `json:"pos"`
	Mobile         ChannelSummary `json:"movil"`
}

// InventorySummary contiene indicadores de inventario y caducidad.
type InventorySummary struct {
	Products       int `json:"productos"`
	OutOfStock     int `json:"productosAgotados"`
	LowStock       int `json:"productosStockBajo"`
	ExpiringLots   int `json:"lotesProximosCaducar"`
	ExpiredLots    int `json:"lotesCaducadosConStock"`
	AvailableUnits int `json:"unidadesDisponibles"`
}

// QualitySummary resume el estado de la cadena de frío.
type QualitySummary struct {
	OpenAlerts          int        `json:"alertasAbiertas"`
	RecognizedAlerts    int        `json:"alertasReconocidas"`
	ActiveIncidents     int        `json:"incidentesActivos"`
	ClosedIncidents     int        `json:"incidentesCerrados"`
	LatestTemperature   *float64   `json:"ultimaTemperatura,omitempty"`
	LatestReadingStatus string     `json:"estadoUltimaLectura,omitempty"`
	LatestReadingAt     *time.Time `json:"fechaUltimaLectura,omitempty"`
}

// Recommendation representa una decisión sugerida al gerente.
type Recommendation struct {
	Code     string `json:"codigo"`
	Module   string `json:"modulo"`
	Priority string `json:"prioridad"`
	Title    string `json:"titulo"`
	Message  string `json:"mensaje"`
}

// ExecutiveSummary contiene la información principal del SIG.
type ExecutiveSummary struct {
	HasERPData          bool                  `json:"tieneDatosERP"`
	LastSynchronization *erpdata.ImportResult `json:"ultimaSincronizacion,omitempty"`
	Sales               SalesSummary          `json:"ventas"`
	Inventory           InventorySummary      `json:"inventario"`
	Quality             QualitySummary        `json:"calidad"`
	Recommendations     []Recommendation      `json:"recomendaciones"`
	UpdatedAt           time.Time             `json:"fechaActualizacion"`
}