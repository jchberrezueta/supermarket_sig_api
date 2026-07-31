package management

import "time"

// InventoryOverview contiene los indicadores generales de inventario.
type InventoryOverview struct {
	Products       int   `json:"productos"`
	AvailableUnits int64 `json:"unidadesDisponibles"`
	OutOfStock     int   `json:"productosAgotados"`
	LowStock       int   `json:"productosStockBajo"`
	LotsWithStock  int   `json:"lotesConStock"`
	ExpiringLots   int   `json:"lotesProximosCaducar"`
	ExpiredLots    int   `json:"lotesCaducados"`
	ExpirationDays int   `json:"diasEvaluadosCaducidad"`
}

// CriticalStockItem representa un producto que requiere atención.
type CriticalStockItem struct {
	ProductID    int64  `json:"idProductoOrigen"`
	CategoryID   int64  `json:"idCategoriaOrigen"`
	Category     string `json:"categoria"`
	Name         string `json:"nombre"`
	CurrentStock int64  `json:"stockActual"`
	MinimumStock int64  `json:"stockMinimo"`
	Status       string `json:"estado"`
}

// CriticalStockReport contiene los productos con stock crítico.
type CriticalStockReport struct {
	Items []CriticalStockItem `json:"items"`
	Total int                 `json:"total"`
	Limit int                 `json:"limite"`
}

// ExpirationLotItem representa un lote caducado o próximo a caducar.
type ExpirationLotItem struct {
	LotID          int64     `json:"idLoteOrigen"`
	ProductID      int64     `json:"idProductoOrigen"`
	Product        string    `json:"producto"`
	ExpirationDate time.Time `json:"fechaCaducidad"`
	DaysRemaining  int       `json:"diasRestantes"`
	Stock          int64     `json:"stock"`
	Status         string    `json:"estado"`
}

// ExpirationReport contiene el análisis de caducidad.
type ExpirationReport struct {
	Days  int                 `json:"diasEvaluados"`
	Items []ExpirationLotItem `json:"items"`
	Total int                 `json:"total"`
	Limit int                 `json:"limite"`
}

// InventoryMovementSummary contiene indicadores de movimientos.
type InventoryMovementSummary struct {
	TotalMovements int   `json:"movimientos"`
	Entries        int   `json:"entradas"`
	Outputs        int   `json:"salidas"`
	EntryUnits     int64 `json:"unidadesEntrada"`
	OutputUnits    int64 `json:"unidadesSalida"`
	NetUnits       int64 `json:"unidadesNetas"`
}

// InventoryMovementItem representa un movimiento sincronizado.
type InventoryMovementItem struct {
	MovementID int64  `json:"idMovimientoOrigen"`
	ProductID  int64  `json:"idProductoOrigen"`
	Product    string `json:"producto"`
	LotID      *int64 `json:"idLoteOrigen,omitempty"`
	Type       string `json:"tipo"`
	Quantity   int64  `json:"cantidad"`

	ProductStockBefore *int64 `json:"stockProductoAnterior,omitempty"`
	ProductStockAfter  *int64 `json:"stockProductoPosterior,omitempty"`
	LotStockBefore     *int64 `json:"stockLoteAnterior,omitempty"`
	LotStockAfter      *int64 `json:"stockLotePosterior,omitempty"`

	SourceDocument string    `json:"documentoOrigen,omitempty"`
	SourceUser     string    `json:"usuarioOrigen,omitempty"`
	Date           time.Time `json:"fechaMovimiento"`
}

// InventoryMovementReport contiene la trazabilidad del inventario.
type InventoryMovementReport struct {
	Period  SalesPeriod              `json:"periodo"`
	Type    string                   `json:"tipo,omitempty"`
	Summary InventoryMovementSummary `json:"resumen"`
	Items   []InventoryMovementItem  `json:"items"`
	Total   int                      `json:"total"`
	Limit   int                      `json:"limite"`
}
