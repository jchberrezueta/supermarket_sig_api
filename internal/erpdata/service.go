package erpdata

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
)

// ValidationError representa un snapshot empresarial inválido.
type ValidationError struct {
	Message string
}

func (err *ValidationError) Error() string {
	return err.Message
}

// Service contiene las reglas de sincronización.
type Service struct {
	repository Repository
	importMu   sync.Mutex
}

// NewService crea el servicio de integración.
func NewService(
	repository Repository,
) *Service {
	return &Service{
		repository: repository,
	}
}

// ImportSnapshot valida y reemplaza el snapshot empresarial.
func (service *Service) ImportSnapshot(
	ctx context.Context,
	snapshot Snapshot,
) (
	ImportResult,
	error,
) {
	normalizeSnapshot(
		&snapshot,
	)

	if err := validateSnapshot(
		snapshot,
	); err != nil {
		return ImportResult{}, err
	}

	service.importMu.Lock()
	defer service.importMu.Unlock()

	result := ImportResult{
		ContractVersion: snapshot.ContractVersion,
		Mode:            "completa",
		GeneratedAt:     snapshot.GeneratedAt,
		ImportedAt:      time.Now(),
		Counts: ImportCounts{
			Categories:  len(snapshot.Categories),
			Companies:   len(snapshot.Companies),
			Suppliers:   len(snapshot.Suppliers),
			Products:    len(snapshot.Products),
			Customers:   len(snapshot.Customers),
			Sales:       len(snapshot.Sales),
			SaleDetails: len(snapshot.SaleDetails),
			Orders:      len(snapshot.Orders),
			Deliveries:  len(snapshot.Deliveries),
			Lots:        len(snapshot.Lots),
			Movements:   len(snapshot.Movements),
		},
	}

	if err := service.repository.ReplaceSnapshot(
		ctx,
		snapshot,
		result,
	); err != nil {
		return ImportResult{},
			fmt.Errorf(
				"no se pudo almacenar el snapshot: %w",
				err,
			)
	}

	return result, nil
}

// Synchronize obtiene el snapshot desde el ERP y lo importa.
func (service *Service) Synchronize(
	ctx context.Context,
	source SnapshotSource,
) (
	ImportResult,
	error,
) {
	if source == nil {
		return ImportResult{},
			errors.New(
				"la fuente de sincronización no está configurada",
			)
	}

	snapshot, err :=
		source.FetchSnapshot(
			ctx,
		)

	if err != nil {
		return ImportResult{},
			fmt.Errorf(
				"no se pudo obtener el snapshot del ERP: %w",
				err,
			)
	}

	return service.ImportSnapshot(
		ctx,
		snapshot,
	)
}

// State devuelve el estado de sincronización.
func (service *Service) State(
	ctx context.Context,
) (
	IntegrationState,
	error,
) {
	return service.repository.State(
		ctx,
	)
}

func validateSnapshot(
	snapshot Snapshot,
) error {
	if strings.TrimSpace(snapshot.ContractVersion) != SnapshotContractVersion {
		return validationError(
			"La versión de contrato %q no es compatible; se esperaba %q.",
			snapshot.ContractVersion,
			SnapshotContractVersion,
		)
	}

	if strings.TrimSpace(snapshot.Source) != SnapshotSourceERP {
		return validationError(
			"La fuente del snapshot %q no es válida; se esperaba %q.",
			snapshot.Source,
			SnapshotSourceERP,
		)
	}

	if snapshot.GeneratedAt.IsZero() {
		return validationError(
			"El snapshot no contiene una fecha de generación válida.",
		)
	}

	categoryIDs := make(
		map[int64]struct{},
	)

	for _, category := range snapshot.Categories {
		if err := addUniqueID(
			categoryIDs,
			category.OriginID,
			"categoría",
		); err != nil {
			return err
		}

		if strings.TrimSpace(category.Name) == "" {
			return validationError(
				"La categoría %d no tiene nombre.",
				category.OriginID,
			)
		}
	}

	companyIDs := make(
		map[int64]struct{},
	)

	for _, company := range snapshot.Companies {
		if err := addUniqueID(
			companyIDs,
			company.OriginID,
			"empresa",
		); err != nil {
			return err
		}

		if strings.TrimSpace(company.Name) == "" {
			return validationError(
				"La empresa %d no tiene nombre.",
				company.OriginID,
			)
		}
	}

	supplierIDs := make(
		map[int64]struct{},
	)

	for _, supplier := range snapshot.Suppliers {
		if err := addUniqueID(
			supplierIDs,
			supplier.OriginID,
			"proveedor",
		); err != nil {
			return err
		}

		if _, exists :=
			companyIDs[supplier.CompanyID]; !exists {
			return validationError(
				"El proveedor %d referencia una empresa inexistente.",
				supplier.OriginID,
			)
		}

		if strings.TrimSpace(supplier.Name) == "" {
			return validationError(
				"El proveedor %d no tiene nombre.",
				supplier.OriginID,
			)
		}
	}

	productIDs := make(
		map[int64]struct{},
	)

	for _, product := range snapshot.Products {
		if err := addUniqueID(
			productIDs,
			product.OriginID,
			"producto",
		); err != nil {
			return err
		}

		if _, exists :=
			categoryIDs[product.CategoryID]; !exists {
			return validationError(
				"El producto %d referencia una categoría inexistente.",
				product.OriginID,
			)
		}

		if strings.TrimSpace(product.Name) == "" {
			return validationError(
				"El producto %d no tiene nombre.",
				product.OriginID,
			)
		}

		if product.Stock < 0 ||
			product.MinimumStock < 0 ||
			product.SalePrice < 0 {
			return validationError(
				"El producto %d contiene valores negativos.",
				product.OriginID,
			)
		}
	}

	customerIDs := make(
		map[int64]struct{},
	)

	for _, customer := range snapshot.Customers {
		if err := addUniqueNonNegativeID(
			customerIDs,
			customer.OriginID,
			"cliente",
		); err != nil {
			return err
		}

		if strings.TrimSpace(customer.Name) == "" {
			return validationError(
				"El cliente %d no tiene nombre.",
				customer.OriginID,
			)
		}
	}

	saleIDs := make(
		map[int64]struct{},
	)

	for _, sale := range snapshot.Sales {
		if err := addUniqueID(
			saleIDs,
			sale.OriginID,
			"venta",
		); err != nil {
			return err
		}

		if sale.CustomerID != nil {
			if _, exists :=
				customerIDs[*sale.CustomerID]; !exists {
				return validationError(
					"La venta %d referencia un cliente inexistente.",
					sale.OriginID,
				)
			}
		}

		if strings.TrimSpace(sale.Invoice) == "" {
			return validationError(
				"La venta %d no tiene número de factura.",
				sale.OriginID,
			)
		}

		if sale.Date.IsZero() {
			return validationError(
				"La venta %d no tiene fecha.",
				sale.OriginID,
			)
		}

		channel := strings.ToLower(
			strings.TrimSpace(sale.Channel),
		)

		if channel != "pos" &&
			channel != "movil" {
			return validationError(
				"La venta %d tiene un canal no válido.",
				sale.OriginID,
			)
		}

		if sale.Subtotal < 0 ||
			sale.Discount < 0 ||
			sale.Tax < 0 ||
			sale.Total < 0 {
			return validationError(
				"La venta %d contiene valores monetarios negativos.",
				sale.OriginID,
			)
		}
	}

	detailIDs := make(
		map[int64]struct{},
	)

	for _, detail := range snapshot.SaleDetails {
		if err := addUniqueID(
			detailIDs,
			detail.OriginID,
			"detalle de venta",
		); err != nil {
			return err
		}

		if _, exists :=
			saleIDs[detail.SaleID]; !exists {
			return validationError(
				"El detalle %d referencia una venta inexistente.",
				detail.OriginID,
			)
		}

		if _, exists :=
			productIDs[detail.ProductID]; !exists {
			return validationError(
				"El detalle %d referencia un producto inexistente.",
				detail.OriginID,
			)
		}

		if detail.Quantity <= 0 {
			return validationError(
				"El detalle %d debe tener una cantidad mayor que cero.",
				detail.OriginID,
			)
		}

		if detail.UnitPrice < 0 ||
			detail.Subtotal < 0 ||
			detail.Discount < 0 ||
			detail.Tax < 0 ||
			detail.Total < 0 {
			return validationError(
				"El detalle %d contiene valores negativos.",
				detail.OriginID,
			)
		}
	}

	orderIDs := make(
		map[int64]struct{},
	)

	for _, order := range snapshot.Orders {
		if err := addUniqueID(
			orderIDs,
			order.OriginID,
			"pedido",
		); err != nil {
			return err
		}

		if _, exists :=
			companyIDs[order.CompanyID]; !exists {
			return validationError(
				"El pedido %d referencia una empresa inexistente.",
				order.OriginID,
			)
		}

		if order.OrderDate.IsZero() {
			return validationError(
				"El pedido %d no tiene fecha.",
				order.OriginID,
			)
		}

		if order.RequestedQuantity < 0 ||
			order.ReceivedQuantity < 0 ||
			order.Total < 0 {
			return validationError(
				"El pedido %d contiene valores negativos.",
				order.OriginID,
			)
		}
	}

	deliveryIDs := make(
		map[int64]struct{},
	)

	for _, delivery := range snapshot.Deliveries {
		if err := addUniqueID(
			deliveryIDs,
			delivery.OriginID,
			"entrega",
		); err != nil {
			return err
		}

		if _, exists :=
			orderIDs[delivery.OrderID]; !exists {
			return validationError(
				"La entrega %d referencia un pedido inexistente.",
				delivery.OriginID,
			)
		}

		if _, exists :=
			supplierIDs[delivery.SupplierID]; !exists {
			return validationError(
				"La entrega %d referencia un proveedor inexistente.",
				delivery.OriginID,
			)
		}

		if delivery.ReceivedQuantity < 0 {
			return validationError(
				"La entrega %d tiene una cantidad negativa.",
				delivery.OriginID,
			)
		}
	}

	lotIDs := make(
		map[int64]struct{},
	)

	for _, lot := range snapshot.Lots {
		if err := addUniqueID(
			lotIDs,
			lot.OriginID,
			"lote",
		); err != nil {
			return err
		}

		if _, exists :=
			productIDs[lot.ProductID]; !exists {
			return validationError(
				"El lote %d referencia un producto inexistente.",
				lot.OriginID,
			)
		}

		if lot.ExpirationDate.IsZero() {
			return validationError(
				"El lote %d no tiene fecha de caducidad.",
				lot.OriginID,
			)
		}

		if lot.Stock < 0 {
			return validationError(
				"El lote %d tiene stock negativo.",
				lot.OriginID,
			)
		}
	}

	movementIDs := make(
		map[int64]struct{},
	)

	for _, movement := range snapshot.Movements {
		if err := addUniqueID(
			movementIDs,
			movement.OriginID,
			"movimiento",
		); err != nil {
			return err
		}

		if _, exists :=
			productIDs[movement.ProductID]; !exists {
			return validationError(
				"El movimiento %d referencia un producto inexistente.",
				movement.OriginID,
			)
		}

		if movement.LotID != nil {
			if _, exists :=
				lotIDs[*movement.LotID]; !exists {
				return validationError(
					"El movimiento %d referencia un lote inexistente.",
					movement.OriginID,
				)
			}
		}

		if movement.Quantity == 0 {
			return validationError(
				"El movimiento %d tiene cantidad cero.",
				movement.OriginID,
			)
		}

		if movement.Date.IsZero() {
			return validationError(
				"El movimiento %d no tiene fecha.",
				movement.OriginID,
			)
		}
	}

	return nil
}

func addUniqueID(
	ids map[int64]struct{},
	id int64,
	entity string,
) error {
	if id <= 0 {
		return validationError(
			"Existe una entidad %s con identificador no válido.",
			entity,
		)
	}

	if _, exists := ids[id]; exists {
		return validationError(
			"El identificador %d está repetido en %s.",
			id,
			entity,
		)
	}

	ids[id] = struct{}{}

	return nil
}

func addUniqueNonNegativeID(
	ids map[int64]struct{},
	id int64,
	entity string,
) error {
	if id < 0 {
		return validationError(
			"Existe una entidad %s con identificador negativo.",
			entity,
		)
	}

	if _, exists := ids[id]; exists {
		return validationError(
			"El identificador %d está repetido en %s.",
			id,
			entity,
		)
	}

	ids[id] = struct{}{}

	return nil
}

func validationError(
	format string,
	arguments ...any,
) error {
	return &ValidationError{
		Message: fmt.Sprintf(
			format,
			arguments...,
		),
	}
}

// SnapshotSource representa una fuente capaz de entregar
// el contrato empresarial del ERP.
type SnapshotSource interface {
	FetchSnapshot(
		ctx context.Context,
	) (
		Snapshot,
		error,
	)
}

func normalizeSnapshot(
	snapshot *Snapshot,
) {
	if snapshot.Categories == nil {
		snapshot.Categories = []Category{}
	}

	if snapshot.Companies == nil {
		snapshot.Companies = []Company{}
	}

	if snapshot.Suppliers == nil {
		snapshot.Suppliers = []Supplier{}
	}

	if snapshot.Products == nil {
		snapshot.Products = []Product{}
	}

	if snapshot.Customers == nil {
		snapshot.Customers = []Customer{}
	}

	if snapshot.Sales == nil {
		snapshot.Sales = []Sale{}
	}

	if snapshot.SaleDetails == nil {
		snapshot.SaleDetails = []SaleDetail{}
	}

	if snapshot.Orders == nil {
		snapshot.Orders = []Order{}
	}

	if snapshot.Deliveries == nil {
		snapshot.Deliveries = []Delivery{}
	}

	if snapshot.Lots == nil {
		snapshot.Lots = []Lot{}
	}

	if snapshot.Movements == nil {
		snapshot.Movements = []InventoryMovement{}
	}
}

// CurrentSnapshot devuelve la información empresarial sincronizada.
func (service *Service) CurrentSnapshot(
	ctx context.Context,
) (
	Snapshot,
	bool,
	error,
) {
	return service.repository.CurrentSnapshot(
		ctx,
	)
}
