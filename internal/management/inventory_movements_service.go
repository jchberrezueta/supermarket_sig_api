package management

import (
	"context"
	"fmt"
	"sort"
	"strings"
)

// InventoryMovements devuelve los movimientos de inventario.
func (service *Service) InventoryMovements(
	ctx context.Context,
	period SalesPeriod,
	movementType string,
	limit int,
) (
	InventoryMovementReport,
	error,
) {
	snapshot, exists, err :=
		service.erpService.CurrentSnapshot(
			ctx,
		)

	if err != nil {
		return InventoryMovementReport{},
			fmt.Errorf(
				"no se pudo consultar la información empresarial: %w",
				err,
			)
	}

	movementType = normalize(
		movementType,
	)

	result := InventoryMovementReport{
		Period: period,
		Type:   movementType,
		Limit:  limit,
		Items: make(
			[]InventoryMovementItem,
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

	for _, movement := range snapshot.Movements {
		movementDate := movement.Date.Format(
			"2006-01-02",
		)

		if period.From != "" &&
			movementDate < period.From {
			continue
		}

		if period.To != "" &&
			movementDate > period.To {
			continue
		}

		normalizedType := normalize(
			movement.Type,
		)

		if movementType != "" &&
			normalizedType != movementType {
			continue
		}

		result.Summary.TotalMovements++
		result.Summary.NetUnits += movement.Quantity

		switch {
		case movement.Quantity > 0:
			result.Summary.Entries++
			result.Summary.EntryUnits +=
				movement.Quantity

		case movement.Quantity < 0:
			result.Summary.Outputs++
			result.Summary.OutputUnits +=
				-movement.Quantity
		}

		result.Items = append(
			result.Items,
			InventoryMovementItem{
				MovementID: movement.OriginID,
				ProductID:  movement.ProductID,
				Product: strings.TrimSpace(
					productNames[movement.ProductID],
				),
				LotID:              movement.LotID,
				Type:               movement.Type,
				Quantity:           movement.Quantity,
				ProductStockBefore: movement.ProductStockBefore,
				ProductStockAfter:  movement.ProductStockAfter,
				LotStockBefore:     movement.LotStockBefore,
				LotStockAfter:      movement.LotStockAfter,
				SourceDocument:     movement.SourceDocument,
				SourceUser:         movement.SourceUser,
				Date:               movement.Date,
			},
		)
	}

	sort.Slice(
		result.Items,
		func(
			left int,
			right int,
		) bool {
			leftItem := result.Items[left]
			rightItem := result.Items[right]

			if !leftItem.Date.Equal(
				rightItem.Date,
			) {
				return leftItem.Date.After(
					rightItem.Date,
				)
			}

			return leftItem.MovementID >
				rightItem.MovementID
		},
	)

	result.Total = len(
		result.Items,
	)

	if len(result.Items) > limit {
		result.Items =
			result.Items[:limit]
	}

	return result, nil
}
