package erpdata

import "time"

// Category representa una categoría sincronizada desde el ERP.
type Category struct {
	OriginID int64  `json:"idOrigen"`
	Name     string `json:"nombre"`
	Status   string `json:"estado"`
}

// Company representa una empresa proveedora del ERP.
type Company struct {
	OriginID    int64  `json:"idOrigen"`
	Name        string `json:"nombre"`
	Responsible string `json:"responsable,omitempty"`
	Phone       string `json:"telefono,omitempty"`
	Email       string `json:"correo,omitempty"`
	Status      string `json:"estado"`
}

// Supplier representa una persona proveedora vinculada a una empresa.
type Supplier struct {
	OriginID       int64  `json:"idOrigen"`
	CompanyID      int64  `json:"idEmpresaOrigen"`
	Identification string `json:"identificacion,omitempty"`
	Name           string `json:"nombre"`
	Phone          string `json:"telefono,omitempty"`
	Email          string `json:"correo,omitempty"`
	Status         string `json:"estado"`
}

// Product representa un producto sincronizado.
type Product struct {
	OriginID     int64   `json:"idOrigen"`
	CategoryID   int64   `json:"idCategoriaOrigen"`
	Barcode      string  `json:"codigoBarra,omitempty"`
	Name         string  `json:"nombre"`
	Stock        int64   `json:"stockActual"`
	MinimumStock int64   `json:"stockMinimo"`
	SalePrice    float64 `json:"precioVenta"`
	Status       string  `json:"estado"`
}

// Customer representa un cliente sincronizado.
type Customer struct {
	OriginID       int64  `json:"idOrigen"`
	Identification string `json:"identificacion,omitempty"`
	Name           string `json:"nombre"`
	Email          string `json:"correo,omitempty"`
	Phone          string `json:"telefono,omitempty"`
	Status         string `json:"estado"`
}

// Sale representa la cabecera de una venta.
type Sale struct {
	OriginID   int64     `json:"idOrigen"`
	CustomerID *int64    `json:"idClienteOrigen,omitempty"`
	Invoice    string    `json:"numeroFactura"`
	Date       time.Time `json:"fechaVenta"`
	Channel    string    `json:"canal"`
	Status     string    `json:"estado"`
	Subtotal   float64   `json:"subtotal"`
	Discount   float64   `json:"descuento"`
	Tax        float64   `json:"iva"`
	Total      float64   `json:"total"`
}

// SaleDetail representa un producto vendido.
type SaleDetail struct {
	OriginID  int64   `json:"idOrigen"`
	SaleID    int64   `json:"idVentaOrigen"`
	ProductID int64   `json:"idProductoOrigen"`
	Quantity  int64   `json:"cantidad"`
	UnitPrice float64 `json:"precioUnitario"`
	Subtotal  float64 `json:"subtotal"`
	Discount  float64 `json:"descuento"`
	Tax       float64 `json:"iva"`
	Total     float64 `json:"total"`
}

// Order representa un pedido realizado a una empresa.
type Order struct {
	OriginID          int64      `json:"idOrigen"`
	CompanyID         int64      `json:"idEmpresaOrigen"`
	Reason            string     `json:"motivo"`
	Status            string     `json:"estado"`
	OrderDate         time.Time  `json:"fechaPedido"`
	ExpectedDate      *time.Time `json:"fechaEsperada,omitempty"`
	RequestedQuantity int64      `json:"cantidadSolicitada"`
	ReceivedQuantity  int64      `json:"cantidadRecibida"`
	Total             float64    `json:"total"`
}

// Delivery representa una entrega asociada a un pedido.
type Delivery struct {
	OriginID         int64      `json:"idOrigen"`
	OrderID          int64      `json:"idPedidoOrigen"`
	SupplierID       int64      `json:"idProveedorOrigen"`
	DeliveryDate     *time.Time `json:"fechaEntrega,omitempty"`
	Status           string     `json:"estado"`
	ReceivedQuantity int64      `json:"cantidadRecibida"`
}

// Lot representa un lote de inventario.
type Lot struct {
	OriginID       int64     `json:"idOrigen"`
	ProductID      int64     `json:"idProductoOrigen"`
	ExpirationDate time.Time `json:"fechaCaducidad"`
	Stock          int64     `json:"stock"`
	Status         string    `json:"estado"`
}

// InventoryMovement representa un movimiento de inventario.
type InventoryMovement struct {
	OriginID           int64     `json:"idOrigen"`
	ProductID          int64     `json:"idProductoOrigen"`
	LotID              *int64    `json:"idLoteOrigen,omitempty"`
	Type               string    `json:"tipo"`
	Quantity           int64     `json:"cantidad"`
	ProductStockBefore *int64    `json:"stockProductoAnterior,omitempty"`
	ProductStockAfter  *int64    `json:"stockProductoPosterior,omitempty"`
	LotStockBefore     *int64    `json:"stockLoteAnterior,omitempty"`
	LotStockAfter      *int64    `json:"stockLotePosterior,omitempty"`
	SourceDocument     string    `json:"documentoOrigen,omitempty"`
	SourceUser         string    `json:"usuarioOrigen,omitempty"`
	Date               time.Time `json:"fechaMovimiento"`
}

// Snapshot contiene una copia empresarial completa del ERP.
type Snapshot struct {
	GeneratedAt time.Time `json:"fechaGeneracion"`

	Categories []Category `json:"categorias"`
	Companies  []Company  `json:"empresas"`
	Suppliers  []Supplier `json:"proveedores"`
	Products   []Product  `json:"productos"`
	Customers  []Customer `json:"clientes"`

	Sales       []Sale       `json:"ventas"`
	SaleDetails []SaleDetail `json:"detallesVenta"`

	Orders     []Order    `json:"pedidos"`
	Deliveries []Delivery `json:"entregas"`

	Lots      []Lot               `json:"lotes"`
	Movements []InventoryMovement `json:"movimientos"`
}

// ImportCounts contiene la cantidad importada por entidad.
type ImportCounts struct {
	Categories  int `json:"categorias"`
	Companies   int `json:"empresas"`
	Suppliers   int `json:"proveedores"`
	Products    int `json:"productos"`
	Customers   int `json:"clientes"`
	Sales       int `json:"ventas"`
	SaleDetails int `json:"detallesVenta"`
	Orders      int `json:"pedidos"`
	Deliveries  int `json:"entregas"`
	Lots        int `json:"lotes"`
	Movements   int `json:"movimientos"`
}

// ImportResult representa una sincronización terminada.
type ImportResult struct {
	Mode        string       `json:"modo"`
	GeneratedAt time.Time    `json:"fechaGeneracion"`
	ImportedAt  time.Time    `json:"fechaImportacion"`
	Counts      ImportCounts `json:"registros"`
}

// IntegrationState representa el estado local de sincronización.
type IntegrationState struct {
	HasData    bool          `json:"tieneDatos"`
	LastImport *ImportResult `json:"ultimaImportacion,omitempty"`
	Counts     ImportCounts  `json:"registros"`
}
