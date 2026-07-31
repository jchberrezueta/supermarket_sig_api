package erpdata

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

func queryOracleRows[T any](
	ctx context.Context,
	querier oracleQuerier,
	query string,
	entity string,
	scan func(
		*sql.Rows,
	) (
		T,
		error,
	),
) (
	[]T,
	error,
) {
	rows, err :=
		querier.QueryContext(
			ctx,
			query,
		)

	if err != nil {
		return nil,
			fmt.Errorf(
				"no se pudieron consultar %s en Oracle: %w",
				entity,
				err,
			)
	}

	defer rows.Close()

	items := make(
		[]T,
		0,
	)

	for rows.Next() {
		item, err :=
			scan(rows)

		if err != nil {
			return nil,
				fmt.Errorf(
					"no se pudo leer %s en Oracle: %w",
					entity,
					err,
				)
		}

		items = append(
			items,
			item,
		)
	}

	if err := rows.Err(); err != nil {
		return nil,
			fmt.Errorf(
				"falló la lectura de %s en Oracle: %w",
				entity,
				err,
			)
	}

	return items,
		nil
}

func loadCategories(
	ctx context.Context,
	querier oracleQuerier,
) (
	[]Category,
	error,
) {
	const query = `
SELECT id_origen,
       nombre,
       descripcion
FROM sig_categoria
WHERE activo_sync = 1
ORDER BY id_origen`

	return queryOracleRows(
		ctx,
		querier,
		query,
		"categorías",

		func(
			rows *sql.Rows,
		) (
			Category,
			error,
		) {
			var item Category
			var description sql.NullString

			err := rows.Scan(
				&item.OriginID,
				&item.Name,
				&description,
			)

			item.Description =
				nullStringValue(
					description,
				)

			return item,
				err
		},
	)
}

func loadCompanies(
	ctx context.Context,
	querier oracleQuerier,
) (
	[]Company,
	error,
) {
	const query = `
SELECT id_origen,
       nombre,
       responsable,
       telefono,
       correo,
       estado
FROM sig_empresa
WHERE activo_sync = 1
ORDER BY id_origen`

	return queryOracleRows(
		ctx,
		querier,
		query,
		"empresas",

		func(
			rows *sql.Rows,
		) (
			Company,
			error,
		) {
			var item Company

			var responsible sql.NullString
			var phone sql.NullString
			var email sql.NullString

			err := rows.Scan(
				&item.OriginID,
				&item.Name,
				&responsible,
				&phone,
				&email,
				&item.Status,
			)

			item.Responsible =
				nullStringValue(
					responsible,
				)

			item.Phone =
				nullStringValue(
					phone,
				)

			item.Email =
				nullStringValue(
					email,
				)

			return item,
				err
		},
	)
}

func loadSuppliers(
	ctx context.Context,
	querier oracleQuerier,
) (
	[]Supplier,
	error,
) {
	const query = `
SELECT id_origen,
       empresa_id_origen,
       identificacion,
       nombre,
       telefono,
       correo,
       estado
FROM sig_proveedor
WHERE activo_sync = 1
ORDER BY id_origen`

	return queryOracleRows(
		ctx,
		querier,
		query,
		"proveedores",

		func(
			rows *sql.Rows,
		) (
			Supplier,
			error,
		) {
			var item Supplier

			var identification sql.NullString
			var phone sql.NullString
			var email sql.NullString

			err := rows.Scan(
				&item.OriginID,
				&item.CompanyID,
				&identification,
				&item.Name,
				&phone,
				&email,
				&item.Status,
			)

			item.Identification =
				nullStringValue(
					identification,
				)

			item.Phone =
				nullStringValue(
					phone,
				)

			item.Email =
				nullStringValue(
					email,
				)

			return item,
				err
		},
	)
}

func loadProducts(
	ctx context.Context,
	querier oracleQuerier,
) (
	[]Product,
	error,
) {
	const query = `
SELECT id_origen,
       categoria_id_origen,
       codigo_barra,
       nombre,
       stock_actual,
       stock_minimo,
       precio_venta,
       estado
FROM sig_producto
WHERE activo_sync = 1
ORDER BY id_origen`

	return queryOracleRows(
		ctx,
		querier,
		query,
		"productos",

		func(
			rows *sql.Rows,
		) (
			Product,
			error,
		) {
			var item Product
			var barcode sql.NullString

			err := rows.Scan(
				&item.OriginID,
				&item.CategoryID,
				&barcode,
				&item.Name,
				&item.Stock,
				&item.MinimumStock,
				&item.SalePrice,
				&item.Status,
			)

			item.Barcode =
				nullStringValue(
					barcode,
				)

			return item,
				err
		},
	)
}

func loadCustomers(
	ctx context.Context,
	querier oracleQuerier,
) (
	[]Customer,
	error,
) {
	const query = `
SELECT id_origen,
       identificacion,
       nombre,
       correo,
       telefono
FROM sig_cliente
WHERE activo_sync = 1
ORDER BY id_origen`

	return queryOracleRows(
		ctx,
		querier,
		query,
		"clientes",

		func(
			rows *sql.Rows,
		) (
			Customer,
			error,
		) {
			var item Customer

			var identification sql.NullString
			var email sql.NullString
			var phone sql.NullString

			err := rows.Scan(
				&item.OriginID,
				&identification,
				&item.Name,
				&email,
				&phone,
			)

			item.Identification =
				nullStringValue(
					identification,
				)

			item.Email =
				nullStringValue(
					email,
				)

			item.Phone =
				nullStringValue(
					phone,
				)

			return item,
				err
		},
	)
}

func loadSales(
	ctx context.Context,
	querier oracleQuerier,
) (
	[]Sale,
	error,
) {
	const query = `
SELECT id_origen,
       cliente_id_origen,
       numero_factura,
       fecha_venta,
       canal,
       estado,
       subtotal,
       descuento,
       iva,
       total
FROM sig_venta
WHERE activo_sync = 1
ORDER BY id_origen`

	return queryOracleRows(
		ctx,
		querier,
		query,
		"ventas",

		func(
			rows *sql.Rows,
		) (
			Sale,
			error,
		) {
			var item Sale
			var customerID sql.NullInt64

			err := rows.Scan(
				&item.OriginID,
				&customerID,
				&item.Invoice,
				&item.Date,
				&item.Channel,
				&item.Status,
				&item.Subtotal,
				&item.Discount,
				&item.Tax,
				&item.Total,
			)

			item.CustomerID =
				nullInt64Pointer(
					customerID,
				)

			return item,
				err
		},
	)
}

func loadSaleDetails(
	ctx context.Context,
	querier oracleQuerier,
) (
	[]SaleDetail,
	error,
) {
	const query = `
SELECT id_origen,
       venta_id_origen,
       producto_id_origen,
       cantidad,
       precio_unitario,
       subtotal,
       descuento,
       iva,
       total
FROM sig_detalle_venta
WHERE activo_sync = 1
ORDER BY id_origen`

	return queryOracleRows(
		ctx,
		querier,
		query,
		"detalles de venta",

		func(
			rows *sql.Rows,
		) (
			SaleDetail,
			error,
		) {
			var item SaleDetail

			err := rows.Scan(
				&item.OriginID,
				&item.SaleID,
				&item.ProductID,
				&item.Quantity,
				&item.UnitPrice,
				&item.Subtotal,
				&item.Discount,
				&item.Tax,
				&item.Total,
			)

			return item,
				err
		},
	)
}

func loadOrders(
	ctx context.Context,
	querier oracleQuerier,
) (
	[]Order,
	error,
) {
	const query = `
SELECT id_origen,
       empresa_id_origen,
       motivo,
       estado,
       fecha_pedido,
       fecha_esperada,
       cantidad_solicitada,
       cantidad_recibida,
       total
FROM sig_pedido
WHERE activo_sync = 1
ORDER BY id_origen`

	return queryOracleRows(
		ctx,
		querier,
		query,
		"pedidos",

		func(
			rows *sql.Rows,
		) (
			Order,
			error,
		) {
			var item Order
			var expectedDate sql.NullTime

			err := rows.Scan(
				&item.OriginID,
				&item.CompanyID,
				&item.Reason,
				&item.Status,
				&item.OrderDate,
				&expectedDate,
				&item.RequestedQuantity,
				&item.ReceivedQuantity,
				&item.Total,
			)

			item.ExpectedDate =
				nullTimePointer(
					expectedDate,
				)

			return item,
				err
		},
	)
}

func loadDeliveries(
	ctx context.Context,
	querier oracleQuerier,
) (
	[]Delivery,
	error,
) {
	const query = `
SELECT id_origen,
       pedido_id_origen,
       proveedor_id_origen,
       fecha_entrega,
       estado,
       cantidad_recibida
FROM sig_entrega
WHERE activo_sync = 1
ORDER BY id_origen`

	return queryOracleRows(
		ctx,
		querier,
		query,
		"entregas",

		func(
			rows *sql.Rows,
		) (
			Delivery,
			error,
		) {
			var item Delivery
			var deliveryDate sql.NullTime

			err := rows.Scan(
				&item.OriginID,
				&item.OrderID,
				&item.SupplierID,
				&deliveryDate,
				&item.Status,
				&item.ReceivedQuantity,
			)

			item.DeliveryDate =
				nullTimePointer(
					deliveryDate,
				)

			return item,
				err
		},
	)
}

func loadLots(
	ctx context.Context,
	querier oracleQuerier,
) (
	[]Lot,
	error,
) {
	const query = `
SELECT id_origen,
       producto_id_origen,
       fecha_caducidad,
       stock,
       estado
FROM sig_lote
WHERE activo_sync = 1
ORDER BY id_origen`

	return queryOracleRows(
		ctx,
		querier,
		query,
		"lotes",

		func(
			rows *sql.Rows,
		) (
			Lot,
			error,
		) {
			var item Lot

			err := rows.Scan(
				&item.OriginID,
				&item.ProductID,
				&item.ExpirationDate,
				&item.Stock,
				&item.Status,
			)

			return item,
				err
		},
	)
}

func loadMovements(
	ctx context.Context,
	querier oracleQuerier,
) (
	[]InventoryMovement,
	error,
) {
	const query = `
SELECT id_origen,
       producto_id_origen,
       lote_id_origen,
       tipo,
       cantidad,
       stock_producto_anterior,
       stock_producto_posterior,
       stock_lote_anterior,
       stock_lote_posterior,
       documento_origen,
       usuario_origen,
       fecha_movimiento
FROM sig_movimiento_inv
WHERE activo_sync = 1
ORDER BY id_origen`

	return queryOracleRows(
		ctx,
		querier,
		query,
		"movimientos",

		func(
			rows *sql.Rows,
		) (
			InventoryMovement,
			error,
		) {
			var item InventoryMovement

			var lotID sql.NullInt64

			var productBefore sql.NullInt64
			var productAfter sql.NullInt64

			var lotBefore sql.NullInt64
			var lotAfter sql.NullInt64

			var document sql.NullString
			var user sql.NullString

			err := rows.Scan(
				&item.OriginID,
				&item.ProductID,
				&lotID,
				&item.Type,
				&item.Quantity,
				&productBefore,
				&productAfter,
				&lotBefore,
				&lotAfter,
				&document,
				&user,
				&item.Date,
			)

			item.LotID =
				nullInt64Pointer(
					lotID,
				)

			item.ProductStockBefore =
				nullInt64Pointer(
					productBefore,
				)

			item.ProductStockAfter =
				nullInt64Pointer(
					productAfter,
				)

			item.LotStockBefore =
				nullInt64Pointer(
					lotBefore,
				)

			item.LotStockAfter =
				nullInt64Pointer(
					lotAfter,
				)

			item.SourceDocument =
				nullStringValue(
					document,
				)

			item.SourceUser =
				nullStringValue(
					user,
				)

			return item,
				err
		},
	)
}

func nullStringValue(
	value sql.NullString,
) string {
	if !value.Valid {
		return ""
	}

	return value.String
}

func nullInt64Pointer(
	value sql.NullInt64,
) *int64 {
	if !value.Valid {
		return nil
	}

	result := value.Int64

	return &result
}

func nullTimePointer(
	value sql.NullTime,
) *time.Time {
	if !value.Valid {
		return nil
	}

	result := value.Time

	return &result
}
