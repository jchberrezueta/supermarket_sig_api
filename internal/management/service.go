package management

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"supermarket-sig-api/internal/erpdata"
	"supermarket-sig-api/internal/iot"
)

// Service calcula indicadores y recomendaciones gerenciales.
type Service struct {
	erpService *erpdata.Service
	iotService *iot.Service
}

// NewService crea el servicio gerencial.
func NewService(
	erpService *erpdata.Service,
	iotService *iot.Service,
) *Service {
	return &Service{
		erpService: erpService,
		iotService: iotService,
	}
}

// ExecutiveSummary calcula el resumen principal del SIG.
func (service *Service) ExecutiveSummary(
	ctx context.Context,
) (
	ExecutiveSummary,
	error,
) {
	state, err := service.erpService.State(
		ctx,
	)

	if err != nil {
		return ExecutiveSummary{},
			fmt.Errorf(
				"no se pudo consultar el estado de integración: %w",
				err,
			)
	}

	summary := ExecutiveSummary{
		HasERPData:          state.HasData,
		LastSynchronization: state.LastImport,

		Recommendations: make(
			[]Recommendation,
			0,
		),

		UpdatedAt: time.Now(),
	}

	snapshot, exists, err :=
		service.erpService.CurrentSnapshot(
			ctx,
		)

	if err != nil {
		return ExecutiveSummary{},
			fmt.Errorf(
				"no se pudo consultar la información empresarial: %w",
				err,
			)
	}

	if exists {
		service.calculateSales(
			&summary,
			snapshot,
		)

		service.calculateInventory(
			&summary,
			snapshot,
		)

		service.buildERPRecommendations(
			&summary,
		)
	} else {
		summary.Recommendations = append(
			summary.Recommendations,
			Recommendation{
				Code:     "ERP_SIN_DATOS",
				Module:   "integracion",
				Priority: "alta",
				Title:    "Información empresarial no disponible",
				Message:  "Ejecute la sincronización con el ERP para generar los indicadores gerenciales.",
			},
		)
	}

	iotSummary, err :=
		service.iotService.Summary(
			ctx,
			"",
		)

	if err != nil {
		return ExecutiveSummary{},
			fmt.Errorf(
				"no se pudo consultar el resumen IoT: %w",
				err,
			)
	}

	service.applyIoTSummary(
		&summary,
		iotSummary,
	)

	return summary, nil
}

func (service *Service) calculateSales(
	summary *ExecutiveSummary,
	snapshot erpdata.Snapshot,
) {
	completedSaleIDs := make(
		map[int64]struct{},
	)

	for _, sale := range snapshot.Sales {
		status := normalize(
			sale.Status,
		)

		if isCancelledStatus(status) {
			summary.Sales.CancelledSales++

			continue
		}

		if !isCompletedStatus(status) {
			continue
		}

		completedSaleIDs[sale.OriginID] =
			struct{}{}

		summary.Sales.CompletedSales++
		summary.Sales.Total += sale.Total

		switch normalize(sale.Channel) {
		case "pos":
			summary.Sales.POS.Transactions++
			summary.Sales.POS.Total += sale.Total

		case "movil", "móvil":
			summary.Sales.Mobile.Transactions++
			summary.Sales.Mobile.Total += sale.Total
		}
	}

	for _, detail := range snapshot.SaleDetails {
		if _, exists :=
			completedSaleIDs[detail.SaleID]; !exists {
			continue
		}

		summary.Sales.UnitsSold +=
			detail.Quantity
	}

	if summary.Sales.CompletedSales > 0 {
		summary.Sales.AverageTicket =
			summary.Sales.Total /
				float64(
					summary.Sales.CompletedSales,
				)
	}

	summary.Sales.Total =
		roundTwoDecimals(
			summary.Sales.Total,
		)

	summary.Sales.AverageTicket =
		roundTwoDecimals(
			summary.Sales.AverageTicket,
		)

	summary.Sales.POS.Total =
		roundTwoDecimals(
			summary.Sales.POS.Total,
		)

	summary.Sales.Mobile.Total =
		roundTwoDecimals(
			summary.Sales.Mobile.Total,
		)
}

func (service *Service) calculateInventory(
	summary *ExecutiveSummary,
	snapshot erpdata.Snapshot,
) {
	now := time.Now()
	expirationLimit := now.AddDate(
		0,
		0,
		30,
	)

	summary.Inventory.Products =
		len(snapshot.Products)

	for _, product := range snapshot.Products {
		summary.Inventory.AvailableUnits +=
			int(product.Stock)

		switch {
		case product.Stock == 0:
			summary.Inventory.OutOfStock++

		case product.Stock <= product.MinimumStock:
			summary.Inventory.LowStock++
		}
	}

	for _, lot := range snapshot.Lots {
		if lot.Stock <= 0 {
			continue
		}

		switch {
		case lot.ExpirationDate.Before(now):
			summary.Inventory.ExpiredLots++

		case !lot.ExpirationDate.After(
			expirationLimit,
		):
			summary.Inventory.ExpiringLots++
		}
	}
}

func (service *Service) buildERPRecommendations(
	summary *ExecutiveSummary,
) {
	if summary.Sales.CompletedSales == 0 {
		summary.Recommendations = append(
			summary.Recommendations,
			Recommendation{
				Code:     "VENTAS_SIN_MOVIMIENTO",
				Module:   "ventas",
				Priority: "media",
				Title:    "No existen ventas completadas",
				Message:  "Revise el período sincronizado o la actividad comercial registrada.",
			},
		)
	}

	if summary.Inventory.OutOfStock > 0 {
		summary.Recommendations = append(
			summary.Recommendations,
			Recommendation{
				Code:     "INVENTARIO_PRODUCTOS_AGOTADOS",
				Module:   "inventario",
				Priority: "critica",
				Title:    "Productos agotados",
				Message: fmt.Sprintf(
					"Existen %d productos sin stock disponible.",
					summary.Inventory.OutOfStock,
				),
			},
		)
	}

	if summary.Inventory.LowStock > 0 {
		summary.Recommendations = append(
			summary.Recommendations,
			Recommendation{
				Code:     "INVENTARIO_STOCK_BAJO",
				Module:   "inventario",
				Priority: "alta",
				Title:    "Reposición necesaria",
				Message: fmt.Sprintf(
					"Existen %d productos en el nivel mínimo de stock.",
					summary.Inventory.LowStock,
				),
			},
		)
	}

	if summary.Inventory.ExpiringLots > 0 {
		summary.Recommendations = append(
			summary.Recommendations,
			Recommendation{
				Code:     "INVENTARIO_CADUCIDAD_PROXIMA",
				Module:   "inventario",
				Priority: "media",
				Title:    "Priorizar salida de lotes",
				Message: fmt.Sprintf(
					"Existen %d lotes con stock que caducarán durante los próximos 30 días.",
					summary.Inventory.ExpiringLots,
				),
			},
		)
	}

	if summary.Inventory.ExpiredLots > 0 {
		summary.Recommendations = append(
			summary.Recommendations,
			Recommendation{
				Code:     "INVENTARIO_LOTES_CADUCADOS",
				Module:   "inventario",
				Priority: "critica",
				Title:    "Retirar lotes caducados",
				Message: fmt.Sprintf(
					"Existen %d lotes caducados que todavía mantienen stock.",
					summary.Inventory.ExpiredLots,
				),
			},
		)
	}
}

func (service *Service) applyIoTSummary(
	summary *ExecutiveSummary,
	iotSummary iot.IoTSummary,
) {
	summary.Quality.OpenAlerts =
		iotSummary.Alerts.Open

	summary.Quality.RecognizedAlerts =
		iotSummary.Alerts.Recognized

	summary.Quality.ActiveIncidents =
		iotSummary.Incidents.Open +
			iotSummary.Incidents.Recognized +
			iotSummary.Incidents.InTreatment +
			iotSummary.Incidents.Resolved

	summary.Quality.ClosedIncidents =
		iotSummary.Incidents.Closed

	if iotSummary.LatestReading != nil {
		temperature :=
			iotSummary.LatestReading.Temperature

		readAt :=
			iotSummary.LatestReading.ReadAt

		summary.Quality.LatestTemperature =
			&temperature

		summary.Quality.LatestReadingStatus =
			iotSummary.LatestReading.Status

		summary.Quality.LatestReadingAt =
			&readAt
	}

	for _, recommendation := range iotSummary.Recommendations {
		summary.Recommendations = append(
			summary.Recommendations,
			Recommendation{
				Code:     recommendation.Code,
				Module:   "cadena_frio",
				Priority: recommendation.Priority,
				Title:    recommendation.Title,
				Message:  recommendation.Message,
			},
		)
	}

	if len(summary.Recommendations) == 0 {
		summary.Recommendations = append(
			summary.Recommendations,
			Recommendation{
				Code:     "SIG_OPERACION_ESTABLE",
				Module:   "general",
				Priority: "informativa",
				Title:    "Operación estable",
				Message:  "No se detectaron incidencias críticas en los datos analizados.",
			},
		)
	}
}

func normalize(
	value string,
) string {
	return strings.ToLower(
		strings.TrimSpace(
			value,
		),
	)
}

func isCompletedStatus(
	status string,
) bool {
	switch status {
	case "completado",
		"completada",
		"completo",
		"completa",
		"completed":
		return true

	default:
		return false
	}
}

func isCancelledStatus(
	status string,
) bool {
	switch status {
	case "cancelado",
		"cancelada",
		"anulado",
		"anulada":
		return true

	default:
		return false
	}
}

func roundTwoDecimals(
	value float64,
) float64 {
	return math.Round(
		value*100,
	) / 100
}
