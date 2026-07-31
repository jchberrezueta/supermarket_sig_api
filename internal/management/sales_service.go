package management

import (
	"context"
	"fmt"
	"sort"

	"supermarket-sig-api/internal/erpdata"
)

// SalesOverview calcula los indicadores comerciales del período.
func (service *Service) SalesOverview(
	ctx context.Context,
	period SalesPeriod,
) (
	SalesOverview,
	error,
) {
	snapshot, exists, err :=
		service.erpService.CurrentSnapshot(
			ctx,
		)

	if err != nil {
		return SalesOverview{},
			fmt.Errorf(
				"no se pudo consultar la información empresarial: %w",
				err,
			)
	}

	result := SalesOverview{
		Period: period,
	}

	if !exists {
		return result, nil
	}

	completedSaleIDs := make(
		map[int64]struct{},
	)

	for _, sale := range snapshot.Sales {
		if !saleWithinPeriod(
			sale,
			period,
		) {
			continue
		}

		status := normalize(
			sale.Status,
		)

		if isCancelledStatus(status) {
			result.Indicators.CancelledSales++

			continue
		}

		if !isCompletedStatus(status) {
			continue
		}

		completedSaleIDs[sale.OriginID] =
			struct{}{}

		result.Indicators.CompletedSales++
		result.Indicators.Total += sale.Total

		switch normalize(sale.Channel) {
		case "pos":
			result.Indicators.POS.Transactions++
			result.Indicators.POS.Total += sale.Total

		case "movil", "móvil":
			result.Indicators.Mobile.Transactions++
			result.Indicators.Mobile.Total += sale.Total
		}
	}

	for _, detail := range snapshot.SaleDetails {
		if _, exists :=
			completedSaleIDs[detail.SaleID]; !exists {
			continue
		}

		result.Indicators.UnitsSold +=
			detail.Quantity
	}

	if result.Indicators.CompletedSales > 0 {
		result.Indicators.AverageTicket =
			result.Indicators.Total /
				float64(
					result.Indicators.CompletedSales,
				)
	}

	result.Indicators.Total =
		roundTwoDecimals(
			result.Indicators.Total,
		)

	result.Indicators.AverageTicket =
		roundTwoDecimals(
			result.Indicators.AverageTicket,
		)

	result.Indicators.POS.Total =
		roundTwoDecimals(
			result.Indicators.POS.Total,
		)

	result.Indicators.Mobile.Total =
		roundTwoDecimals(
			result.Indicators.Mobile.Total,
		)

	return result, nil
}

// SalesTrend agrupa las ventas completadas por día.
func (service *Service) SalesTrend(
	ctx context.Context,
	period SalesPeriod,
) (
	SalesTrend,
	error,
) {
	snapshot, exists, err :=
		service.erpService.CurrentSnapshot(
			ctx,
		)

	if err != nil {
		return SalesTrend{},
			fmt.Errorf(
				"no se pudo consultar la información empresarial: %w",
				err,
			)
	}

	result := SalesTrend{
		Period: period,

		Items: make(
			[]DailySalesPoint,
			0,
		),
	}

	if !exists {
		return result, nil
	}

	daily := make(
		map[string]DailySalesPoint,
	)

	saleDates := make(
		map[int64]string,
	)

	for _, sale := range snapshot.Sales {
		if !saleWithinPeriod(
			sale,
			period,
		) {
			continue
		}

		if !isCompletedStatus(
			normalize(sale.Status),
		) {
			continue
		}

		date := sale.Date.Format(
			"2006-01-02",
		)

		saleDates[sale.OriginID] =
			date

		point := daily[date]

		point.Date = date
		point.Transactions++
		point.Total += sale.Total

		daily[date] = point
	}

	for _, detail := range snapshot.SaleDetails {
		date, exists :=
			saleDates[detail.SaleID]

		if !exists {
			continue
		}

		point := daily[date]
		point.Units += detail.Quantity

		daily[date] = point
	}

	dates := make(
		[]string,
		0,
		len(daily),
	)

	for date := range daily {
		dates = append(
			dates,
			date,
		)
	}

	sort.Strings(
		dates,
	)

	for _, date := range dates {
		point := daily[date]

		point.Total =
			roundTwoDecimals(
				point.Total,
			)

		result.Items = append(
			result.Items,
			point,
		)
	}

	result.Total = len(
		result.Items,
	)

	return result, nil
}

// TopSellingProducts calcula los productos más vendidos.
func (service *Service) TopSellingProducts(
	ctx context.Context,
	period SalesPeriod,
	limit int,
) (
	ProductSalesRanking,
	error,
) {
	snapshot, exists, err :=
		service.erpService.CurrentSnapshot(
			ctx,
		)

	if err != nil {
		return ProductSalesRanking{},
			fmt.Errorf(
				"no se pudo consultar la información empresarial: %w",
				err,
			)
	}

	result := ProductSalesRanking{
		Period: period,
		Limit:  limit,

		Items: make(
			[]ProductSalesItem,
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

	completedSaleIDs := make(
		map[int64]struct{},
	)

	for _, sale := range snapshot.Sales {
		if !saleWithinPeriod(
			sale,
			period,
		) {
			continue
		}

		if !isCompletedStatus(
			normalize(sale.Status),
		) {
			continue
		}

		completedSaleIDs[sale.OriginID] =
			struct{}{}
	}

	products := make(
		map[int64]ProductSalesItem,
	)

	for _, detail := range snapshot.SaleDetails {
		if _, exists :=
			completedSaleIDs[detail.SaleID]; !exists {
			continue
		}

		item := products[detail.ProductID]

		item.ProductID =
			detail.ProductID

		item.Name =
			productNames[detail.ProductID]

		item.Units +=
			detail.Quantity

		item.Revenue +=
			detail.Total

		products[detail.ProductID] =
			item
	}

	for _, item := range products {
		item.Revenue =
			roundTwoDecimals(
				item.Revenue,
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
			if result.Items[left].Units !=
				result.Items[right].Units {
				return result.Items[left].Units >
					result.Items[right].Units
			}

			if result.Items[left].Revenue !=
				result.Items[right].Revenue {
				return result.Items[left].Revenue >
					result.Items[right].Revenue
			}

			return result.Items[left].Name <
				result.Items[right].Name
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

func saleWithinPeriod(
	sale erpdata.Sale,
	period SalesPeriod,
) bool {
	date := sale.Date.Format(
		"2006-01-02",
	)

	if period.From != "" &&
		date < period.From {
		return false
	}

	if period.To != "" &&
		date > period.To {
		return false
	}

	return true
}

// SalesByCategory calcula las ventas agrupadas por categoría.
func (service *Service) SalesByCategory(
	ctx context.Context,
	period SalesPeriod,
	limit int,
) (
	CategorySalesRanking,
	error,
) {
	snapshot, exists, err :=
		service.erpService.CurrentSnapshot(
			ctx,
		)

	if err != nil {
		return CategorySalesRanking{},
			fmt.Errorf(
				"no se pudo consultar la información empresarial: %w",
				err,
			)
	}

	result := CategorySalesRanking{
		Period: period,
		Limit:  limit,
		Items: make(
			[]CategorySalesItem,
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

	productCategories := make(
		map[int64]int64,
	)

	for _, product := range snapshot.Products {
		productCategories[product.OriginID] =
			product.CategoryID
	}

	completedSaleIDs := make(
		map[int64]struct{},
	)

	for _, sale := range snapshot.Sales {
		if !saleWithinPeriod(
			sale,
			period,
		) {
			continue
		}

		if !isCompletedStatus(
			normalize(sale.Status),
		) {
			continue
		}

		completedSaleIDs[sale.OriginID] =
			struct{}{}
	}

	categories := make(
		map[int64]CategorySalesItem,
	)

	for _, detail := range snapshot.SaleDetails {
		if _, exists :=
			completedSaleIDs[detail.SaleID]; !exists {
			continue
		}

		categoryID, productExists :=
			productCategories[detail.ProductID]

		if !productExists {
			continue
		}

		item := categories[categoryID]

		item.CategoryID = categoryID
		item.Name = categoryNames[categoryID]
		item.Units += detail.Quantity
		item.Revenue += detail.Total

		categories[categoryID] = item
	}

	for _, item := range categories {
		item.Revenue =
			roundTwoDecimals(
				item.Revenue,
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

			if leftItem.Revenue !=
				rightItem.Revenue {
				return leftItem.Revenue >
					rightItem.Revenue
			}

			if leftItem.Units !=
				rightItem.Units {
				return leftItem.Units >
					rightItem.Units
			}

			return leftItem.Name <
				rightItem.Name
		},
	)

	result.Total = len(result.Items)

	if len(result.Items) > limit {
		result.Items =
			result.Items[:limit]
	}

	return result, nil
}
