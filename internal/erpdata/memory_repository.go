package erpdata

import (
	"context"
	"sync"
)

// MemoryRepository mantiene temporalmente el snapshot empresarial.
type MemoryRepository struct {
	mu sync.RWMutex

	snapshot   Snapshot
	lastImport *ImportResult
	hasData    bool
}

// NewMemoryRepository crea un repositorio vacío.
func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{}
}

// ReplaceSnapshot reemplaza todo el conjunto de forma atómica.
func (repository *MemoryRepository) ReplaceSnapshot(
	_ context.Context,
	snapshot Snapshot,
	result ImportResult,
) error {
	repository.mu.Lock()
	defer repository.mu.Unlock()

	snapshotCopy := cloneSnapshot(
		snapshot,
	)

	resultCopy := result

	repository.snapshot = snapshotCopy
	repository.lastImport = &resultCopy
	repository.hasData = true

	return nil
}

// CurrentSnapshot devuelve una copia del snapshot actual.
func (repository *MemoryRepository) CurrentSnapshot(
	_ context.Context,
) (
	Snapshot,
	bool,
	error,
) {
	repository.mu.RLock()
	defer repository.mu.RUnlock()

	if !repository.hasData {
		return Snapshot{},
			false,
			nil
	}

	return cloneSnapshot(repository.snapshot),
		true,
		nil
}

// State devuelve el estado de la última importación.
func (repository *MemoryRepository) State(
	_ context.Context,
) (
	IntegrationState,
	error,
) {
	repository.mu.RLock()
	defer repository.mu.RUnlock()

	state := IntegrationState{
		HasData: repository.hasData,
	}

	if repository.lastImport != nil {
		resultCopy := *repository.lastImport

		state.LastImport = &resultCopy
		state.Counts = resultCopy.Counts
	}

	return state, nil
}

func cloneSnapshot(
	source Snapshot,
) Snapshot {
	result := source

	result.Categories = append(
		[]Category(nil),
		source.Categories...,
	)

	result.Companies = append(
		[]Company(nil),
		source.Companies...,
	)

	result.Suppliers = append(
		[]Supplier(nil),
		source.Suppliers...,
	)

	result.Products = append(
		[]Product(nil),
		source.Products...,
	)

	result.Customers = append(
		[]Customer(nil),
		source.Customers...,
	)

	result.Sales = append(
		[]Sale(nil),
		source.Sales...,
	)

	result.SaleDetails = append(
		[]SaleDetail(nil),
		source.SaleDetails...,
	)

	result.Orders = append(
		[]Order(nil),
		source.Orders...,
	)

	result.Deliveries = append(
		[]Delivery(nil),
		source.Deliveries...,
	)

	result.Lots = append(
		[]Lot(nil),
		source.Lots...,
	)

	result.Movements = append(
		[]InventoryMovement(nil),
		source.Movements...,
	)

	return result
}
