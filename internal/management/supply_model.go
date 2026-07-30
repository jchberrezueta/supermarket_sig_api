package management

import "time"

// SupplyOverview contiene los indicadores generales de abastecimiento.
type SupplyOverview struct {
	TotalOrders           int     `json:"pedidos"`
	CompletedOrders       int     `json:"pedidosCompletados"`
	PendingOrders         int     `json:"pedidosPendientes"`
	CancelledOrders       int     `json:"pedidosCancelados"`
	RequestedQuantity     int64   `json:"cantidadSolicitada"`
	ReceivedQuantity      int64   `json:"cantidadRecibida"`
	FulfillmentPercentage float64 `json:"porcentajeCumplimiento"`
	PurchaseTotal         float64 `json:"totalPedidos"`
	TotalDeliveries       int     `json:"entregas"`
	CompletedDeliveries   int     `json:"entregasCompletas"`
	PendingDeliveries     int     `json:"entregasPendientes"`
	OnTimeDeliveries      int     `json:"entregasPuntuales"`
	LateDeliveries        int     `json:"entregasTardias"`
	OnTimePercentage      float64 `json:"porcentajePuntualidad"`
}

// SupplierPerformanceItem representa el desempeño de una empresa proveedora.
type SupplierPerformanceItem struct {
	CompanyID             int64   `json:"idEmpresaOrigen"`
	Company               string  `json:"empresa"`
	Suppliers             int     `json:"proveedores"`
	Orders                int     `json:"pedidos"`
	CompletedOrders       int     `json:"pedidosCompletados"`
	PendingOrders         int     `json:"pedidosPendientes"`
	RequestedQuantity     int64   `json:"cantidadSolicitada"`
	ReceivedQuantity      int64   `json:"cantidadRecibida"`
	FulfillmentPercentage float64 `json:"porcentajeCumplimiento"`
	Deliveries            int     `json:"entregas"`
	OnTimeDeliveries      int     `json:"entregasPuntuales"`
	LateDeliveries        int     `json:"entregasTardias"`
	OnTimePercentage      float64 `json:"porcentajePuntualidad"`
	PurchaseTotal         float64 `json:"totalPedidos"`
}

// SupplierPerformanceReport contiene el ranking de proveedores.
type SupplierPerformanceReport struct {
	Items []SupplierPerformanceItem `json:"items"`
	Total int                       `json:"total"`
	Limit int                       `json:"limite"`
}

// SupplyOrderItem representa un pedido sincronizado desde el ERP.
type SupplyOrderItem struct {
	OrderID               int64      `json:"idPedidoOrigen"`
	CompanyID             int64      `json:"idEmpresaOrigen"`
	Company               string     `json:"empresa"`
	Reason                string     `json:"motivo"`
	Status                string     `json:"estado"`
	OrderDate             time.Time  `json:"fechaPedido"`
	ExpectedDate          *time.Time `json:"fechaEsperada,omitempty"`
	RequestedQuantity     int64      `json:"cantidadSolicitada"`
	ReceivedQuantity      int64      `json:"cantidadRecibida"`
	FulfillmentPercentage float64    `json:"porcentajeCumplimiento"`
	Total                 float64    `json:"total"`
}

// SupplyOrderReport contiene los pedidos más recientes.
type SupplyOrderReport struct {
	Status string            `json:"estado,omitempty"`
	Items  []SupplyOrderItem `json:"items"`
	Total  int               `json:"total"`
	Limit  int               `json:"limite"`
}
