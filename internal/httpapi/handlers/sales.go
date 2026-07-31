package handlers

import (
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"supermarket-sig-api/internal/management"
)

// SalesOverview devuelve los indicadores comerciales.
func (handler *ManagementHandler) SalesOverview(
	w http.ResponseWriter,
	r *http.Request,
) {
	period, ok := parseSalesPeriod(
		w,
		r,
	)

	if !ok {
		return
	}

	result, err :=
		handler.service.SalesOverview(
			r.Context(),
			period,
		)

	if err != nil {
		handler.writeManagementError(
			w,
			err,
		)

		return
	}

	WriteJSON(
		w,
		http.StatusOK,
		successResponse{
			Success: true,
			Data:    result,
		},
	)
}

// SalesTrend devuelve las ventas agrupadas por día.
func (handler *ManagementHandler) SalesTrend(
	w http.ResponseWriter,
	r *http.Request,
) {
	period, ok := parseSalesPeriod(
		w,
		r,
	)

	if !ok {
		return
	}

	result, err :=
		handler.service.SalesTrend(
			r.Context(),
			period,
		)

	if err != nil {
		handler.writeManagementError(
			w,
			err,
		)

		return
	}

	WriteJSON(
		w,
		http.StatusOK,
		successResponse{
			Success: true,
			Data:    result,
		},
	)
}

// TopSellingProducts devuelve el ranking de productos.
func (handler *ManagementHandler) TopSellingProducts(
	w http.ResponseWriter,
	r *http.Request,
) {
	period, ok := parseSalesPeriod(
		w,
		r,
	)

	if !ok {
		return
	}

	limit, ok := parseSalesLimit(
		w,
		r,
	)

	if !ok {
		return
	}

	result, err :=
		handler.service.TopSellingProducts(
			r.Context(),
			period,
			limit,
		)

	if err != nil {
		handler.writeManagementError(
			w,
			err,
		)

		return
	}

	WriteJSON(
		w,
		http.StatusOK,
		successResponse{
			Success: true,
			Data:    result,
		},
	)
}

// SalesByCategory devuelve las ventas agrupadas por categoría.
func (handler *ManagementHandler) SalesByCategory(
	w http.ResponseWriter,
	r *http.Request,
) {
	period, ok := parseSalesPeriod(
		w,
		r,
	)

	if !ok {
		return
	}

	limit, ok := parseSalesLimit(
		w,
		r,
	)

	if !ok {
		return
	}

	result, err :=
		handler.service.SalesByCategory(
			r.Context(),
			period,
			limit,
		)

	if err != nil {
		handler.writeManagementError(
			w,
			err,
		)

		return
	}

	WriteJSON(
		w,
		http.StatusOK,
		successResponse{
			Success: true,
			Data:    result,
		},
	)
}

func parseSalesPeriod(
	w http.ResponseWriter,
	r *http.Request,
) (
	management.SalesPeriod,
	bool,
) {
	period := management.SalesPeriod{
		From: strings.TrimSpace(
			r.URL.Query().Get(
				"desde",
			),
		),

		To: strings.TrimSpace(
			r.URL.Query().Get(
				"hasta",
			),
		),
	}

	if period.From != "" {
		if _, err := time.Parse(
			"2006-01-02",
			period.From,
		); err != nil {
			WriteError(
				w,
				http.StatusBadRequest,
				"invalid_sales_period",
				"El parámetro desde debe usar el formato YYYY-MM-DD.",
			)

			return management.SalesPeriod{},
				false
		}
	}

	if period.To != "" {
		if _, err := time.Parse(
			"2006-01-02",
			period.To,
		); err != nil {
			WriteError(
				w,
				http.StatusBadRequest,
				"invalid_sales_period",
				"El parámetro hasta debe usar el formato YYYY-MM-DD.",
			)

			return management.SalesPeriod{},
				false
		}
	}

	if period.From != "" &&
		period.To != "" &&
		period.From > period.To {
		WriteError(
			w,
			http.StatusBadRequest,
			"invalid_sales_period",
			"La fecha desde no puede ser posterior a la fecha hasta.",
		)

		return management.SalesPeriod{},
			false
	}

	return period, true
}

func parseSalesLimit(
	w http.ResponseWriter,
	r *http.Request,
) (
	int,
	bool,
) {
	value := strings.TrimSpace(
		r.URL.Query().Get(
			"limite",
		),
	)

	if value == "" {
		return 10, true
	}

	limit, err := strconv.Atoi(
		value,
	)

	if err != nil ||
		limit < 1 ||
		limit > 50 {
		WriteError(
			w,
			http.StatusBadRequest,
			"invalid_sales_limit",
			"El límite debe ser un número entre 1 y 50.",
		)

		return 0, false
	}

	return limit, true
}

func (handler *ManagementHandler) writeManagementError(
	w http.ResponseWriter,
	err error,
) {
	log.Printf(
		"error del servicio gerencial: %v",
		err,
	)

	WriteError(
		w,
		http.StatusInternalServerError,
		"management_query_failed",
		"No se pudo procesar la consulta gerencial.",
	)
}
