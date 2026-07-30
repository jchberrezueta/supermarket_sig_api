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
