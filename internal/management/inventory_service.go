package management

import (
	"context"
	"fmt"
	"sort"
	"time"
)

// InventoryOverview calcula los indicadores generales de inventario.
func (service *Service) InventoryOverview(
	ctx context.Context,
	expirationDays int,
) (
	InventoryOverview,
	error,
) {
	snapshot, exists, err :=
		service.erpService.CurrentSnapshot(
			ctx,
		)

	if err != nil {
		return InventoryOverview{},
			fmt.Errorf(
				"no se pudo consultar la información empresarial: %w",
				err,
			)
	}

	result := InventoryOverview{
		ExpirationDays: expirationDays,
	}

	if !exists {
		return result, nil
	}

	result.Products =
		len(snapshot.Products)

	for _, product := range snapshot.Products {
		result.AvailableUnits +=
			product.Stock

		switch {
		case product.Stock == 0:
			result.OutOfStock++

		case product.Stock <= product.MinimumStock:
			result.LowStock++
		}
	}

	today := beginningOfDay(
		time.Now(),
	)

	expirationLimit :=
		today.AddDate(
			0,
			0,
			expirationDays,
		)

	for _, lot := range snapshot.Lots {
		if lot.Stock <= 0 {
			continue
		}

		result.LotsWithStock++

		expirationDate :=
			beginningOfDay(
				lot.ExpirationDate,
			)

		switch {
		case expirationDate.Before(today):
			result.ExpiredLots++

		case !expirationDate.After(
			expirationLimit,
		):
			result.ExpiringLots++
		}
	}

	return result, nil
}

// CriticalStock devuelve los productos agotados o bajo el mínimo.
func (service *Service) CriticalStock(
	ctx context.Context,
	limit int,
) (
	CriticalStockReport,
	error,
) {
	snapshot, exists, err :=
		service.erpService.CurrentSnapshot(
			ctx,
		)

	if err != nil {
		return CriticalStockReport{},
			fmt.Errorf(
				"no se pudo consultar la información empresarial: %w",
				err,
			)
	}

	result := CriticalStockReport{
		Limit: limit,

		Items: make(
			[]CriticalStockItem,
			0,
		),
	}

	if !exists {
		return result, nil
	}

	categoryNames := make(
		map[int64]string,
	)

	for _, category := range snapshot.Categories {
		categoryNames[category.OriginID] =
			category.Name
	}

	for _, product := range snapshot.Products {
		if product.Stock >
			product.MinimumStock {
			continue
		}

		status := "stock_bajo"

		if product.Stock == 0 {
			status = "agotado"
		}

		result.Items = append(
			result.Items,
			CriticalStockItem{
				ProductID:    product.OriginID,
				CategoryID:   product.CategoryID,
				Category:     categoryNames[product.CategoryID],
				Name:         product.Name,
				CurrentStock: product.Stock,
				MinimumStock: product.MinimumStock,
				Status:       status,
			},
		)
	}

	sort.Slice(
		result.Items,
		func(
			left int,
			right int,
		) bool {
			leftItem :=
				result.Items[left]

			rightItem :=
				result.Items[right]

			if leftItem.Status !=
				rightItem.Status {
				return leftItem.Status ==
					"agotado"
			}

			if leftItem.CurrentStock !=
				rightItem.CurrentStock {
				return leftItem.CurrentStock <
					rightItem.CurrentStock
			}

			return leftItem.Name <
				rightItem.Name
		},
	)

	result.Total =
		len(result.Items)

	if len(result.Items) > limit {
		result.Items =
			result.Items[:limit]
	}

	return result, nil
}

// ExpiringLots devuelve los lotes caducados o próximos a caducar.
func (service *Service) ExpiringLots(
	ctx context.Context,
	days int,
	limit int,
) (
	ExpirationReport,
	error,
) {
	snapshot, exists, err :=
		service.erpService.CurrentSnapshot(
			ctx,
		)

	if err != nil {
		return ExpirationReport{},
			fmt.Errorf(
				"no se pudo consultar la información empresarial: %w",
				err,
			)
	}

	result := ExpirationReport{
		Days:  days,
		Limit: limit,

		Items: make(
			[]ExpirationLotItem,
			0,
		),
	}

	if !exists {
		return result, nil
	}

	productNames := make(
		map[int64]string,
	)

	for _, product := range snapshot.Products {
		productNames[product.OriginID] =
			product.Name
	}

	today := beginningOfDay(
		time.Now(),
	)

	expirationLimit :=
		today.AddDate(
			0,
			0,
			days,
		)

	for _, lot := range snapshot.Lots {
		if lot.Stock <= 0 {
			continue
		}

		expirationDate :=
			beginningOfDay(
				lot.ExpirationDate,
			)

		if expirationDate.After(
			expirationLimit,
		) {
			continue
		}

		daysRemaining :=
			int(
				expirationDate.Sub(
					today,
				).Hours() / 24,
			)

		status := "proximo_caducar"

		if expirationDate.Before(today) {
			status = "caducado"
		}

		result.Items = append(
			result.Items,
			ExpirationLotItem{
				LotID:          lot.OriginID,
				ProductID:      lot.ProductID,
				Product:        productNames[lot.ProductID],
				ExpirationDate: lot.ExpirationDate,
				DaysRemaining:  daysRemaining,
				Stock:          lot.Stock,
				Status:         status,
			},
		)
	}

	sort.Slice(
		result.Items,
		func(
			left int,
			right int,
		) bool {
			leftItem :=
				result.Items[left]

			rightItem :=
				result.Items[right]

			if !leftItem.ExpirationDate.Equal(
				rightItem.ExpirationDate,
			) {
				return leftItem.ExpirationDate.Before(
					rightItem.ExpirationDate,
				)
			}

			return leftItem.Product <
				rightItem.Product
		},
	)

	result.Total =
		len(result.Items)

	if len(result.Items) > limit {
		result.Items =
			result.Items[:limit]
	}

	return result, nil
}

func beginningOfDay(
	value time.Time,
) time.Time {
	return time.Date(
		value.Year(),
		value.Month(),
		value.Day(),
		0,
		0,
		0,
		0,
		value.Location(),
	)
}
