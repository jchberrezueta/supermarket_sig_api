package management

// SalesPeriod representa el período solicitado para el análisis.
type SalesPeriod struct {
	From string `json:"desde,omitempty"`
	To   string `json:"hasta,omitempty"`
}

// SalesOverview contiene los indicadores comerciales del período.
type SalesOverview struct {
	Period     SalesPeriod  `json:"periodo"`
	Indicators SalesSummary `json:"indicadores"`
}

// DailySalesPoint representa las ventas agrupadas por fecha.
type DailySalesPoint struct {
	Date         string  `json:"fecha"`
	Transactions int     `json:"transacciones"`
	Total        float64 `json:"total"`
	Units        int64   `json:"unidades"`
}

// SalesTrend contiene la evolución diaria de las ventas.
type SalesTrend struct {
	Period SalesPeriod       `json:"periodo"`
	Items  []DailySalesPoint `json:"items"`
	Total  int               `json:"total"`
}

// ProductSalesItem representa el desempeño comercial de un producto.
type ProductSalesItem struct {
	ProductID int64   `json:"idProductoOrigen"`
	Name      string  `json:"nombre"`
	Units     int64   `json:"unidadesVendidas"`
	Revenue   float64 `json:"ingresos"`
}

// ProductSalesRanking contiene los productos más vendidos.
type ProductSalesRanking struct {
	Period SalesPeriod        `json:"periodo"`
	Items  []ProductSalesItem `json:"items"`
	Total  int                `json:"total"`
	Limit  int                `json:"limite"`
}
