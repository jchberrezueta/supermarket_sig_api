package erpdata

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// OracleRepository almacena el snapshot empresarial en Oracle.
type OracleRepository struct {
	db *sql.DB
}

var _ Repository = (*OracleRepository)(nil)

// NewOracleRepository crea el repositorio empresarial Oracle.
func NewOracleRepository(
	db *sql.DB,
) *OracleRepository {
	return &OracleRepository{
		db: db,
	}
}

// ReplaceSnapshot reemplaza lógicamente el snapshot
// completo dentro de una transacción Oracle.
func (repository *OracleRepository) ReplaceSnapshot(
	ctx context.Context,
	snapshot Snapshot,
	result ImportResult,
) error {
	if repository == nil ||
		repository.db == nil {
		return errors.New(
			"el repositorio Oracle no está configurado",
		)
	}

	tx, err :=
		repository.db.BeginTx(
			ctx,
			nil,
		)

	if err != nil {
		return fmt.Errorf(
			"no se pudo iniciar la transacción Oracle: %w",
			err,
		)
	}

	defer func() {
		_ = tx.Rollback()
	}()

	importedAt :=
		result.ImportedAt

	if importedAt.IsZero() {
		importedAt = time.Now()
	}

	syncID, err :=
		insertSynchronization(
			ctx,
			tx,
			snapshot,
			result,
			importedAt,
		)

	if err != nil {
		return err
	}

	if err := deactivatePreviousSnapshot(
		ctx,
		tx,
		syncID,
		importedAt,
	); err != nil {
		return err
	}

	if err := mergeSnapshot(
		ctx,
		tx,
		snapshot,
		syncID,
		importedAt,
	); err != nil {
		return err
	}

	if err := completeSynchronization(
		ctx,
		tx,
		syncID,
		importedAt,
	); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf(
			"no se pudo confirmar la sincronización Oracle: %w",
			err,
		)
	}

	return nil
}

// CurrentSnapshot reconstruye desde Oracle
// el snapshot empresarial activo.
func (repository *OracleRepository) CurrentSnapshot(
	ctx context.Context,
) (
	Snapshot,
	bool,
	error,
) {
	if repository == nil ||
		repository.db == nil {
		return Snapshot{},
			false,
			errors.New(
				"el repositorio Oracle no está configurado",
			)
	}

	tx, err :=
		repository.db.BeginTx(
			ctx,
			nil,
		)

	if err != nil {
		return Snapshot{},
			false,
			fmt.Errorf(
				"no se pudo iniciar la lectura Oracle: %w",
				err,
			)
	}

	defer func() {
		_ = tx.Rollback()
	}()

	/*
		Oracle mantendrá una vista consistente durante
		todas las consultas que reconstruyen el snapshot.
	*/
	if _, err := tx.ExecContext(
		ctx,
		"SET TRANSACTION READ ONLY",
	); err != nil {
		return Snapshot{},
			false,
			fmt.Errorf(
				"no se pudo iniciar la lectura consistente Oracle: %w",
				err,
			)
	}

	state, err :=
		loadIntegrationState(
			ctx,
			tx,
		)

	if err != nil {
		return Snapshot{},
			false,
			err
	}

	if !state.HasData ||
		state.LastImport == nil {
		if err := tx.Commit(); err != nil {
			return Snapshot{},
				false,
				fmt.Errorf(
					"no se pudo cerrar la lectura Oracle: %w",
					err,
				)
		}

		return Snapshot{},
			false,
			nil
	}

	snapshot := Snapshot{
		ContractVersion: state.LastImport.ContractVersion,

		Source: SnapshotSourceERP,

		GeneratedAt: state.LastImport.GeneratedAt,
	}

	snapshot.Categories, err =
		loadCategories(
			ctx,
			tx,
		)

	if err != nil {
		return Snapshot{},
			false,
			err
	}

	snapshot.Companies, err =
		loadCompanies(
			ctx,
			tx,
		)

	if err != nil {
		return Snapshot{},
			false,
			err
	}

	snapshot.Suppliers, err =
		loadSuppliers(
			ctx,
			tx,
		)

	if err != nil {
		return Snapshot{},
			false,
			err
	}

	snapshot.Products, err =
		loadProducts(
			ctx,
			tx,
		)

	if err != nil {
		return Snapshot{},
			false,
			err
	}

	snapshot.Customers, err =
		loadCustomers(
			ctx,
			tx,
		)

	if err != nil {
		return Snapshot{},
			false,
			err
	}

	snapshot.Sales, err =
		loadSales(
			ctx,
			tx,
		)

	if err != nil {
		return Snapshot{},
			false,
			err
	}

	snapshot.SaleDetails, err =
		loadSaleDetails(
			ctx,
			tx,
		)

	if err != nil {
		return Snapshot{},
			false,
			err
	}

	snapshot.Orders, err =
		loadOrders(
			ctx,
			tx,
		)

	if err != nil {
		return Snapshot{},
			false,
			err
	}

	snapshot.Deliveries, err =
		loadDeliveries(
			ctx,
			tx,
		)

	if err != nil {
		return Snapshot{},
			false,
			err
	}

	snapshot.Lots, err =
		loadLots(
			ctx,
			tx,
		)

	if err != nil {
		return Snapshot{},
			false,
			err
	}

	snapshot.Movements, err =
		loadMovements(
			ctx,
			tx,
		)

	if err != nil {
		return Snapshot{},
			false,
			err
	}

	if err := tx.Commit(); err != nil {
		return Snapshot{},
			false,
			fmt.Errorf(
				"no se pudo cerrar la lectura Oracle: %w",
				err,
			)
	}

	return snapshot,
		true,
		nil
}

// State devuelve la última sincronización
// completada en Oracle.
func (repository *OracleRepository) State(
	ctx context.Context,
) (
	IntegrationState,
	error,
) {
	if repository == nil ||
		repository.db == nil {
		return IntegrationState{},
			errors.New(
				"el repositorio Oracle no está configurado",
			)
	}

	return loadIntegrationState(
		ctx,
		repository.db,
	)
}

type oracleQuerier interface {
	QueryContext(
		ctx context.Context,
		query string,
		args ...any,
	) (
		*sql.Rows,
		error,
	)

	QueryRowContext(
		ctx context.Context,
		query string,
		args ...any,
	) *sql.Row
}

func loadIntegrationState(
	ctx context.Context,
	querier oracleQuerier,
) (
	IntegrationState,
	error,
) {
	const query = `
SELECT
    version_contrato,
    modo,
    fecha_generacion,
    fecha_fin,
    total_categorias,
    total_empresas,
    total_proveedores,
    total_productos,
    total_clientes,
    total_ventas,
    total_detalles_venta,
    total_pedidos,
    total_entregas,
    total_lotes,
    total_movimientos
FROM (
    SELECT
        version_contrato,
        modo,
        fecha_generacion,
        fecha_fin,
        total_categorias,
        total_empresas,
        total_proveedores,
        total_productos,
        total_clientes,
        total_ventas,
        total_detalles_venta,
        total_pedidos,
        total_entregas,
        total_lotes,
        total_movimientos
    FROM sig_sincronizacion
    WHERE estado = 'completada'
    ORDER BY id DESC
)
WHERE ROWNUM = 1`

	var result ImportResult

	err := querier.QueryRowContext(
		ctx,
		query,
	).Scan(
		&result.ContractVersion,
		&result.Mode,
		&result.GeneratedAt,
		&result.ImportedAt,
		&result.Counts.Categories,
		&result.Counts.Companies,
		&result.Counts.Suppliers,
		&result.Counts.Products,
		&result.Counts.Customers,
		&result.Counts.Sales,
		&result.Counts.SaleDetails,
		&result.Counts.Orders,
		&result.Counts.Deliveries,
		&result.Counts.Lots,
		&result.Counts.Movements,
	)

	if errors.Is(
		err,
		sql.ErrNoRows,
	) {
		return IntegrationState{
			HasData: false,
		}, nil
	}

	if err != nil {
		return IntegrationState{},
			fmt.Errorf(
				"no se pudo consultar la última sincronización Oracle: %w",
				err,
			)
	}

	return IntegrationState{
		HasData: true,

		LastImport: &result,

		Counts: result.Counts,
	}, nil
}
