package management

import (
	"context"
	"fmt"
	"sort"
)

// SupplyOverview calcula los indicadores generales de abastecimiento.
func (service *Service) SupplyOverview(
	ctx context.Context,
) (
	SupplyOverview,
	error,
) {
	snapshot, exists, err :=
		service.erpService.CurrentSnapshot(
			ctx,
		)

	if err != nil {
		return SupplyOverview{},
			fmt.Errorf(
				"no se pudo consultar la información empresarial: %w",
				err,
			)
	}

	result := SupplyOverview{}

	if !exists {
		return result, nil
	}

	for _, order := range snapshot.Orders {
		result.TotalOrders++
		result.RequestedQuantity += order.RequestedQuantity
		result.ReceivedQuantity += order.ReceivedQuantity
		result.PurchaseTotal += order.Total

		status := normalize(order.Status)

		switch {
		case isCompletedStatus(status):
			result.CompletedOrders++

		case isCancelledStatus(status):
			result.CancelledOrders++

		default:
			result.PendingOrders++
		}
	}

	for _, delivery := range snapshot.Deliveries {
		result.TotalDeliveries++

		if isCompletedStatus(
			normalize(delivery.Status),
		) {
			result.CompletedDeliveries++
		} else {
			result.PendingDeliveries++
		}
	}

	ordersByID := make(
		map[int64]int,
	)

	for index, order := range snapshot.Orders {
		ordersByID[order.OriginID] = index
	}

	for _, delivery := range snapshot.Deliveries {
		if delivery.DeliveryDate == nil {
			continue
		}

		orderIndex, exists :=
			ordersByID[delivery.OrderID]

		if !exists {
			continue
		}

		order := snapshot.Orders[orderIndex]

		if order.ExpectedDate == nil {
			continue
		}

		if delivery.DeliveryDate.After(
			*order.ExpectedDate,
		) {
			result.LateDeliveries++
		} else {
			result.OnTimeDeliveries++
		}
	}

	if result.RequestedQuantity > 0 {
		result.FulfillmentPercentage =
			float64(result.ReceivedQuantity) *
				100 /
				float64(result.RequestedQuantity)
	}

	evaluatedDeliveries :=
		result.OnTimeDeliveries +
			result.LateDeliveries

	if evaluatedDeliveries > 0 {
		result.OnTimePercentage =
			float64(result.OnTimeDeliveries) *
				100 /
				float64(evaluatedDeliveries)
	}

	result.FulfillmentPercentage =
		roundTwoDecimals(
			result.FulfillmentPercentage,
		)

	result.OnTimePercentage =
		roundTwoDecimals(
			result.OnTimePercentage,
		)

	result.PurchaseTotal =
		roundTwoDecimals(
			result.PurchaseTotal,
		)

	return result, nil
}

// SupplierPerformance calcula el desempeño por empresa proveedora.
func (service *Service) SupplierPerformance(
	ctx context.Context,
	limit int,
) (
	SupplierPerformanceReport,
	error,
) {
	snapshot, exists, err :=
		service.erpService.CurrentSnapshot(
			ctx,
		)

	if err != nil {
		return SupplierPerformanceReport{},
			fmt.Errorf(
				"no se pudo consultar la información empresarial: %w",
				err,
			)
	}

	result := SupplierPerformanceReport{
		Limit: limit,
		Items: make(
			[]SupplierPerformanceItem,
			0,
		),
	}

	if !exists {
		return result, nil
	}

	items := make(
		map[int64]SupplierPerformanceItem,
	)

	for _, company := range snapshot.Companies {
		items[company.OriginID] =
			SupplierPerformanceItem{
				CompanyID: company.OriginID,
				Company:   company.Name,
			}
	}

	for _, supplier := range snapshot.Suppliers {
		item, companyExists :=
			items[supplier.CompanyID]

		if !companyExists {
			continue
		}

		item.Suppliers++

		items[supplier.CompanyID] =
			item
	}

	ordersByID := make(
		map[int64]int,
	)

	for index, order := range snapshot.Orders {
		ordersByID[order.OriginID] = index

		item, companyExists :=
			items[order.CompanyID]

		if !companyExists {
			continue
		}

		item.Orders++
		item.RequestedQuantity += order.RequestedQuantity
		item.ReceivedQuantity += order.ReceivedQuantity
		item.PurchaseTotal += order.Total

		status := normalize(order.Status)

		if isCompletedStatus(status) {
			item.CompletedOrders++
		} else if !isCancelledStatus(status) {
			item.PendingOrders++
		}

		items[order.CompanyID] =
			item
	}

	for _, delivery := range snapshot.Deliveries {
		orderIndex, orderExists :=
			ordersByID[delivery.OrderID]

		if !orderExists {
			continue
		}

		order := snapshot.Orders[orderIndex]

		item, companyExists :=
			items[order.CompanyID]

		if !companyExists {
			continue
		}

		item.Deliveries++

		if delivery.DeliveryDate != nil &&
			order.ExpectedDate != nil {
			if delivery.DeliveryDate.After(
				*order.ExpectedDate,
			) {
				item.LateDeliveries++
			} else {
				item.OnTimeDeliveries++
			}
		}

		items[order.CompanyID] =
			item
	}

	for _, item := range items {
		if item.Orders == 0 {
			continue
		}

		if item.RequestedQuantity > 0 {
			item.FulfillmentPercentage =
				float64(item.ReceivedQuantity) *
					100 /
					float64(item.RequestedQuantity)
		}

		evaluatedDeliveries :=
			item.OnTimeDeliveries +
				item.LateDeliveries

		if evaluatedDeliveries > 0 {
			item.OnTimePercentage =
				float64(item.OnTimeDeliveries) *
					100 /
					float64(evaluatedDeliveries)
		}

		item.FulfillmentPercentage =
			roundTwoDecimals(
				item.FulfillmentPercentage,
			)

		item.OnTimePercentage =
			roundTwoDecimals(
				item.OnTimePercentage,
			)

		item.PurchaseTotal =
			roundTwoDecimals(
				item.PurchaseTotal,
			)

		result.Items = append(
			result.Items,
			item,
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

			if leftItem.FulfillmentPercentage !=
				rightItem.FulfillmentPercentage {
				return leftItem.FulfillmentPercentage >
					rightItem.FulfillmentPercentage
			}

			if leftItem.OnTimePercentage !=
				rightItem.OnTimePercentage {
				return leftItem.OnTimePercentage >
					rightItem.OnTimePercentage
			}

			if leftItem.PurchaseTotal !=
				rightItem.PurchaseTotal {
				return leftItem.PurchaseTotal >
					rightItem.PurchaseTotal
			}

			return leftItem.Company <
				rightItem.Company
		},
	)

	result.Total = len(result.Items)

	if len(result.Items) > limit {
		result.Items =
			result.Items[:limit]
	}

	return result, nil
}

// SupplyOrders devuelve los pedidos más recientes.
func (service *Service) SupplyOrders(
	ctx context.Context,
	status string,
	limit int,
) (
	SupplyOrderReport,
	error,
) {
	snapshot, exists, err :=
		service.erpService.CurrentSnapshot(
			ctx,
		)

	if err != nil {
		return SupplyOrderReport{},
			fmt.Errorf(
				"no se pudo consultar la información empresarial: %w",
				err,
			)
	}

	status = normalize(status)

	result := SupplyOrderReport{
		Status: status,
		Limit:  limit,
		Items: make(
			[]SupplyOrderItem,
			0,
		),
	}

	if !exists {
		return result, nil
	}

	companyNames := make(
		map[int64]string,
	)

	for _, company := range snapshot.Companies {
		companyNames[company.OriginID] =
			company.Name
	}

	for _, order := range snapshot.Orders {
		if status != "" &&
			normalize(order.Status) != status {
			continue
		}

		fulfillmentPercentage := 0.0

		if order.RequestedQuantity > 0 {
			fulfillmentPercentage =
				float64(order.ReceivedQuantity) *
					100 /
					float64(order.RequestedQuantity)
		}

		result.Items = append(
			result.Items,
			SupplyOrderItem{
				OrderID:   order.OriginID,
				CompanyID: order.CompanyID,
				Company:   companyNames[order.CompanyID],
				Reason:    order.Reason,
				Status:    order.Status,
				OrderDate: order.OrderDate,

				ExpectedDate: order.ExpectedDate,

				RequestedQuantity: order.RequestedQuantity,

				ReceivedQuantity: order.ReceivedQuantity,

				FulfillmentPercentage: roundTwoDecimals(
					fulfillmentPercentage,
				),

				Total: roundTwoDecimals(
					order.Total,
				),
			},
		)
	}

	sort.Slice(
		result.Items,
		func(
			left int,
			right int,
		) bool {
			return result.Items[left].
				OrderDate.
				After(
					result.Items[right].
						OrderDate,
				)
		},
	)

	result.Total = len(result.Items)

	if len(result.Items) > limit {
		result.Items =
			result.Items[:limit]
	}

	return result, nil
}
