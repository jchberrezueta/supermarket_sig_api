package erpdata

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type oracleMergeColumn struct {
	name string
	cast string
}

type oracleMergeSpec struct {
	table   string
	key     string
	columns []oracleMergeColumn
}

func insertSynchronization(
	ctx context.Context,
	tx *sql.Tx,
	snapshot Snapshot,
	result ImportResult,
	importedAt time.Time,
) (
	int64,
	error,
) {
	const statement = `
INSERT INTO sig_sincronizacion (
    version_contrato,
    fuente,
    modo,
    estado,
    fecha_generacion,
    fecha_inicio,
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
) VALUES (
    :1,
    :2,
    :3,
    'iniciada',
    :4,
    :5,
    :6,
    :7,
    :8,
    :9,
    :10,
    :11,
    :12,
    :13,
    :14,
    :15,
    :16
)
RETURNING id INTO :17`

	var syncID int64

	_, err := tx.ExecContext(
		ctx,
		statement,
		snapshot.ContractVersion,
		snapshot.Source,
		result.Mode,
		snapshot.GeneratedAt,
		importedAt,
		result.Counts.Categories,
		result.Counts.Companies,
		result.Counts.Suppliers,
		result.Counts.Products,
		result.Counts.Customers,
		result.Counts.Sales,
		result.Counts.SaleDetails,
		result.Counts.Orders,
		result.Counts.Deliveries,
		result.Counts.Lots,
		result.Counts.Movements,
		sql.Out{
			Dest: &syncID,
		},
	)

	if err != nil {
		return 0,
			fmt.Errorf(
				"no se pudo registrar el inicio de la sincronización Oracle: %w",
				err,
			)
	}

	return syncID,
		nil
}

func deactivatePreviousSnapshot(
	ctx context.Context,
	tx *sql.Tx,
	syncID int64,
	importedAt time.Time,
) error {
	tables := []string{
		"sig_movimiento_inv",
		"sig_lote",
		"sig_entrega",
		"sig_pedido",
		"sig_detalle_venta",
		"sig_venta",
		"sig_cliente",
		"sig_producto",
		"sig_proveedor",
		"sig_empresa",
		"sig_categoria",
	}

	for _, table := range tables {
		statement := fmt.Sprintf(
			`UPDATE %s
SET activo_sync = 0,
    id_sync_ultima = :1,
    sincronizado_en = :2
WHERE activo_sync = 1`,
			table,
		)

		if _, err := tx.ExecContext(
			ctx,
			statement,
			syncID,
			importedAt,
		); err != nil {
			return fmt.Errorf(
				"no se pudo desactivar el snapshot anterior en %s: %w",
				table,
				err,
			)
		}
	}

	return nil
}

func mergeSnapshot(
	ctx context.Context,
	tx *sql.Tx,
	snapshot Snapshot,
	syncID int64,
	importedAt time.Time,
) error {
	categorySpec := oracleMergeSpec{
		table: "sig_categoria",
		key:   "id_origen",

		columns: []oracleMergeColumn{
			{
				name: "id_origen",
			},
			{
				name: "nombre",
			},
			{
				name: "descripcion",
				cast: "VARCHAR2(1000)",
			},
			{
				name: "id_sync_ultima",
			},
			{
				name: "sincronizado_en",
			},
		},
	}

	for _, item := range snapshot.Categories {
		if err := executeMerge(
			ctx,
			tx,
			categorySpec,
			item.OriginID,
			item.OriginID,
			item.Name,
			nullIfBlank(
				item.Description,
			),
			syncID,
			importedAt,
		); err != nil {
			return fmt.Errorf(
				"no se pudo sincronizar la categoría %d: %w",
				item.OriginID,
				err,
			)
		}
	}

	companySpec := oracleMergeSpec{
		table: "sig_empresa",
		key:   "id_origen",

		columns: []oracleMergeColumn{
			{
				name: "id_origen",
			},
			{
				name: "nombre",
			},
			{
				name: "responsable",
				cast: "VARCHAR2(200)",
			},
			{
				name: "telefono",
				cast: "VARCHAR2(50)",
			},
			{
				name: "correo",
				cast: "VARCHAR2(320)",
			},
			{
				name: "estado",
			},
			{
				name: "id_sync_ultima",
			},
			{
				name: "sincronizado_en",
			},
		},
	}

	for _, item := range snapshot.Companies {
		if err := executeMerge(
			ctx,
			tx,
			companySpec,
			item.OriginID,
			item.OriginID,
			item.Name,
			nullIfBlank(
				item.Responsible,
			),
			nullIfBlank(
				item.Phone,
			),
			nullIfBlank(
				item.Email,
			),
			item.Status,
			syncID,
			importedAt,
		); err != nil {
			return fmt.Errorf(
				"no se pudo sincronizar la empresa %d: %w",
				item.OriginID,
				err,
			)
		}
	}

	supplierSpec := oracleMergeSpec{
		table: "sig_proveedor",
		key:   "id_origen",

		columns: []oracleMergeColumn{
			{
				name: "id_origen",
			},
			{
				name: "empresa_id_origen",
			},
			{
				name: "identificacion",
				cast: "VARCHAR2(50)",
			},
			{
				name: "nombre",
			},
			{
				name: "telefono",
				cast: "VARCHAR2(50)",
			},
			{
				name: "correo",
				cast: "VARCHAR2(320)",
			},
			{
				name: "estado",
			},
			{
				name: "id_sync_ultima",
			},
			{
				name: "sincronizado_en",
			},
		},
	}

	for _, item := range snapshot.Suppliers {
		if err := executeMerge(
			ctx,
			tx,
			supplierSpec,
			item.OriginID,
			item.OriginID,
			item.CompanyID,
			nullIfBlank(
				item.Identification,
			),
			item.Name,
			nullIfBlank(
				item.Phone,
			),
			nullIfBlank(
				item.Email,
			),
			item.Status,
			syncID,
			importedAt,
		); err != nil {
			return fmt.Errorf(
				"no se pudo sincronizar el proveedor %d: %w",
				item.OriginID,
				err,
			)
		}
	}

	productSpec := oracleMergeSpec{
		table: "sig_producto",
		key:   "id_origen",

		columns: []oracleMergeColumn{
			{
				name: "id_origen",
			},
			{
				name: "categoria_id_origen",
			},
			{
				name: "codigo_barra",
				cast: "VARCHAR2(100)",
			},
			{
				name: "nombre",
			},
			{
				name: "stock_actual",
			},
			{
				name: "stock_minimo",
			},
			{
				name: "precio_venta",
			},
			{
				name: "estado",
			},
			{
				name: "id_sync_ultima",
			},
			{
				name: "sincronizado_en",
			},
		},
	}

	for _, item := range snapshot.Products {
		if err := executeMerge(
			ctx,
			tx,
			productSpec,
			item.OriginID,
			item.OriginID,
			item.CategoryID,
			nullIfBlank(
				item.Barcode,
			),
			item.Name,
			item.Stock,
			item.MinimumStock,
			item.SalePrice,
			item.Status,
			syncID,
			importedAt,
		); err != nil {
			return fmt.Errorf(
				"no se pudo sincronizar el producto %d: %w",
				item.OriginID,
				err,
			)
		}
	}

	customerSpec := oracleMergeSpec{
		table: "sig_cliente",
		key:   "id_origen",

		columns: []oracleMergeColumn{
			{
				name: "id_origen",
			},
			{
				name: "identificacion",
				cast: "VARCHAR2(50)",
			},
			{
				name: "nombre",
			},
			{
				name: "correo",
				cast: "VARCHAR2(320)",
			},
			{
				name: "telefono",
				cast: "VARCHAR2(50)",
			},
			{
				name: "id_sync_ultima",
			},
			{
				name: "sincronizado_en",
			},
		},
	}

	for _, item := range snapshot.Customers {
		if err := executeMerge(
			ctx,
			tx,
			customerSpec,
			item.OriginID,
			item.OriginID,
			nullIfBlank(
				item.Identification,
			),
			item.Name,
			nullIfBlank(
				item.Email,
			),
			nullIfBlank(
				item.Phone,
			),
			syncID,
			importedAt,
		); err != nil {
			return fmt.Errorf(
				"no se pudo sincronizar el cliente %d: %w",
				item.OriginID,
				err,
			)
		}
	}

	saleSpec := oracleMergeSpec{
		table: "sig_venta",
		key:   "id_origen",

		columns: []oracleMergeColumn{
			{
				name: "id_origen",
			},
			{
				name: "cliente_id_origen",
				cast: "NUMBER(19)",
			},
			{
				name: "numero_factura",
			},
			{
				name: "fecha_venta",
			},
			{
				name: "canal",
			},
			{
				name: "estado",
			},
			{
				name: "subtotal",
			},
			{
				name: "descuento",
			},
			{
				name: "iva",
			},
			{
				name: "total",
			},
			{
				name: "id_sync_ultima",
			},
			{
				name: "sincronizado_en",
			},
		},
	}

	for _, item := range snapshot.Sales {
		if err := executeMerge(
			ctx,
			tx,
			saleSpec,
			item.OriginID,
			item.OriginID,
			nullableInt64(
				item.CustomerID,
			),
			item.Invoice,
			item.Date,
			item.Channel,
			item.Status,
			item.Subtotal,
			item.Discount,
			item.Tax,
			item.Total,
			syncID,
			importedAt,
		); err != nil {
			return fmt.Errorf(
				"no se pudo sincronizar la venta %d: %w",
				item.OriginID,
				err,
			)
		}
	}

	detailSpec := oracleMergeSpec{
		table: "sig_detalle_venta",
		key:   "id_origen",

		columns: []oracleMergeColumn{
			{
				name: "id_origen",
			},
			{
				name: "venta_id_origen",
			},
			{
				name: "producto_id_origen",
			},
			{
				name: "cantidad",
			},
			{
				name: "precio_unitario",
			},
			{
				name: "subtotal",
			},
			{
				name: "descuento",
			},
			{
				name: "iva",
			},
			{
				name: "total",
			},
			{
				name: "id_sync_ultima",
			},
			{
				name: "sincronizado_en",
			},
		},
	}

	for _, item := range snapshot.SaleDetails {
		if err := executeMerge(
			ctx,
			tx,
			detailSpec,
			item.OriginID,
			item.OriginID,
			item.SaleID,
			item.ProductID,
			item.Quantity,
			item.UnitPrice,
			item.Subtotal,
			item.Discount,
			item.Tax,
			item.Total,
			syncID,
			importedAt,
		); err != nil {
			return fmt.Errorf(
				"no se pudo sincronizar el detalle de venta %d: %w",
				item.OriginID,
				err,
			)
		}
	}

	orderSpec := oracleMergeSpec{
		table: "sig_pedido",
		key:   "id_origen",

		columns: []oracleMergeColumn{
			{
				name: "id_origen",
			},
			{
				name: "empresa_id_origen",
			},
			{
				name: "motivo",
			},
			{
				name: "estado",
			},
			{
				name: "fecha_pedido",
			},
			{
				name: "fecha_esperada",
				cast: "TIMESTAMP WITH TIME ZONE",
			},
			{
				name: "cantidad_solicitada",
			},
			{
				name: "cantidad_recibida",
			},
			{
				name: "total",
			},
			{
				name: "id_sync_ultima",
			},
			{
				name: "sincronizado_en",
			},
		},
	}

	for _, item := range snapshot.Orders {
		if err := executeMerge(
			ctx,
			tx,
			orderSpec,
			item.OriginID,
			item.OriginID,
			item.CompanyID,
			item.Reason,
			item.Status,
			item.OrderDate,
			nullableTime(
				item.ExpectedDate,
			),
			item.RequestedQuantity,
			item.ReceivedQuantity,
			item.Total,
			syncID,
			importedAt,
		); err != nil {
			return fmt.Errorf(
				"no se pudo sincronizar el pedido %d: %w",
				item.OriginID,
				err,
			)
		}
	}

	deliverySpec := oracleMergeSpec{
		table: "sig_entrega",
		key:   "id_origen",

		columns: []oracleMergeColumn{
			{
				name: "id_origen",
			},
			{
				name: "pedido_id_origen",
			},
			{
				name: "proveedor_id_origen",
			},
			{
				name: "fecha_entrega",
				cast: "TIMESTAMP WITH TIME ZONE",
			},
			{
				name: "estado",
			},
			{
				name: "cantidad_recibida",
			},
			{
				name: "id_sync_ultima",
			},
			{
				name: "sincronizado_en",
			},
		},
	}

	for _, item := range snapshot.Deliveries {
		if err := executeMerge(
			ctx,
			tx,
			deliverySpec,
			item.OriginID,
			item.OriginID,
			item.OrderID,
			item.SupplierID,
			nullableTime(
				item.DeliveryDate,
			),
			item.Status,
			item.ReceivedQuantity,
			syncID,
			importedAt,
		); err != nil {
			return fmt.Errorf(
				"no se pudo sincronizar la entrega %d: %w",
				item.OriginID,
				err,
			)
		}
	}

	lotSpec := oracleMergeSpec{
		table: "sig_lote",
		key:   "id_origen",

		columns: []oracleMergeColumn{
			{
				name: "id_origen",
			},
			{
				name: "producto_id_origen",
			},
			{
				name: "fecha_caducidad",
			},
			{
				name: "stock",
			},
			{
				name: "estado",
			},
			{
				name: "id_sync_ultima",
			},
			{
				name: "sincronizado_en",
			},
		},
	}

	for _, item := range snapshot.Lots {
		if err := executeMerge(
			ctx,
			tx,
			lotSpec,
			item.OriginID,
			item.OriginID,
			item.ProductID,
			item.ExpirationDate,
			item.Stock,
			item.Status,
			syncID,
			importedAt,
		); err != nil {
			return fmt.Errorf(
				"no se pudo sincronizar el lote %d: %w",
				item.OriginID,
				err,
			)
		}
	}

	movementSpec := oracleMergeSpec{
		table: "sig_movimiento_inv",
		key:   "id_origen",

		columns: []oracleMergeColumn{
			{
				name: "id_origen",
			},
			{
				name: "producto_id_origen",
			},
			{
				name: "lote_id_origen",
				cast: "NUMBER(19)",
			},
			{
				name: "tipo",
			},
			{
				name: "cantidad",
			},
			{
				name: "stock_producto_anterior",
				cast: "NUMBER(19)",
			},
			{
				name: "stock_producto_posterior",
				cast: "NUMBER(19)",
			},
			{
				name: "stock_lote_anterior",
				cast: "NUMBER(19)",
			},
			{
				name: "stock_lote_posterior",
				cast: "NUMBER(19)",
			},
			{
				name: "documento_origen",
				cast: "VARCHAR2(200)",
			},
			{
				name: "usuario_origen",
				cast: "VARCHAR2(150)",
			},
			{
				name: "fecha_movimiento",
			},
			{
				name: "id_sync_ultima",
			},
			{
				name: "sincronizado_en",
			},
		},
	}

	for _, item := range snapshot.Movements {
		if err := executeMerge(
			ctx,
			tx,
			movementSpec,
			item.OriginID,
			item.OriginID,
			item.ProductID,
			nullableInt64(
				item.LotID,
			),
			item.Type,
			item.Quantity,
			nullableInt64(
				item.ProductStockBefore,
			),
			nullableInt64(
				item.ProductStockAfter,
			),
			nullableInt64(
				item.LotStockBefore,
			),
			nullableInt64(
				item.LotStockAfter,
			),
			nullIfBlank(
				item.SourceDocument,
			),
			nullIfBlank(
				item.SourceUser,
			),
			item.Date,
			syncID,
			importedAt,
		); err != nil {
			return fmt.Errorf(
				"no se pudo sincronizar el movimiento %d: %w",
				item.OriginID,
				err,
			)
		}
	}

	return nil
}

func executeMerge(
	ctx context.Context,
	tx *sql.Tx,
	spec oracleMergeSpec,
	originID int64,
	values ...any,
) error {
	if len(values) !=
		len(spec.columns) {
		return fmt.Errorf(
			"MERGE %s inválido para el registro %d: se esperaban %d valores y se recibieron %d",
			spec.table,
			originID,
			len(spec.columns),
			len(values),
		)
	}

	_, err := tx.ExecContext(
		ctx,
		buildMergeStatement(
			spec,
		),
		values...,
	)

	return err
}

func buildMergeStatement(
	spec oracleMergeSpec,
) string {
	selects := make(
		[]string,
		0,
		len(spec.columns),
	)

	updates := make(
		[]string,
		0,
		len(spec.columns),
	)

	inserts := make(
		[]string,
		0,
		len(spec.columns)+1,
	)

	insertValues := make(
		[]string,
		0,
		len(spec.columns)+1,
	)

	for index, column := range spec.columns {
		bind := fmt.Sprintf(
			":%d",
			index+1,
		)

		if column.cast != "" {
			bind = fmt.Sprintf(
				"CAST(%s AS %s)",
				bind,
				column.cast,
			)
		}

		selects = append(
			selects,
			fmt.Sprintf(
				"%s AS %s",
				bind,
				column.name,
			),
		)

		inserts = append(
			inserts,
			column.name,
		)

		insertValues = append(
			insertValues,
			"source."+column.name,
		)

		if column.name != spec.key {
			updates = append(
				updates,
				fmt.Sprintf(
					"target.%s = source.%s",
					column.name,
					column.name,
				),
			)
		}
	}

	updates = append(
		updates,
		"target.activo_sync = 1",
	)

	inserts = append(
		inserts,
		"activo_sync",
	)

	insertValues = append(
		insertValues,
		"1",
	)

	return fmt.Sprintf(
		`
MERGE INTO %s target
USING (
    SELECT %s
    FROM dual
) source
ON (target.%s = source.%s)
WHEN MATCHED THEN UPDATE SET
    %s
WHEN NOT MATCHED THEN INSERT (
    %s
) VALUES (
    %s
)`,
		spec.table,

		strings.Join(
			selects,
			",\n           ",
		),

		spec.key,
		spec.key,

		strings.Join(
			updates,
			",\n    ",
		),

		strings.Join(
			inserts,
			",\n    ",
		),

		strings.Join(
			insertValues,
			",\n    ",
		),
	)
}

func completeSynchronization(
	ctx context.Context,
	tx *sql.Tx,
	syncID int64,
	importedAt time.Time,
) error {
	const statement = `
UPDATE sig_sincronizacion
SET estado = 'completada',
    fecha_fin = :1,
    mensaje = :2
WHERE id = :3`

	result, err := tx.ExecContext(
		ctx,
		statement,
		importedAt,
		"Sincronización completada correctamente.",
		syncID,
	)

	if err != nil {
		return fmt.Errorf(
			"no se pudo completar la sincronización Oracle: %w",
			err,
		)
	}

	rowsAffected, err :=
		result.RowsAffected()

	if err != nil {
		return fmt.Errorf(
			"no se pudo verificar la sincronización Oracle: %w",
			err,
		)
	}

	if rowsAffected != 1 {
		return fmt.Errorf(
			"la sincronización Oracle %d no pudo marcarse como completada",
			syncID,
		)
	}

	return nil
}

func nullIfBlank(
	value string,
) any {
	if strings.TrimSpace(
		value,
	) == "" {
		return nil
	}

	return value
}

func nullableInt64(
	value *int64,
) any {
	if value == nil {
		return nil
	}

	return *value
}

func nullableTime(
	value *time.Time,
) any {
	if value == nil {
		return nil
	}

	return *value
}
