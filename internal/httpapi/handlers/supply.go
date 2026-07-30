package handlers

import (
	"net/http"
	"strconv"
	"strings"
)

// SupplyOverview devuelve el resumen de abastecimiento.
func (handler *ManagementHandler) SupplyOverview(
	w http.ResponseWriter,
	r *http.Request,
) {
	result, err :=
		handler.service.SupplyOverview(
			r.Context(),
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

// SupplierPerformance devuelve el desempeño de proveedores.
func (handler *ManagementHandler) SupplierPerformance(
	w http.ResponseWriter,
	r *http.Request,
) {
	limit, ok := parseSupplyLimit(
		w,
		r,
	)

	if !ok {
		return
	}

	result, err :=
		handler.service.SupplierPerformance(
			r.Context(),
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

// SupplyOrders devuelve los pedidos sincronizados.
func (handler *ManagementHandler) SupplyOrders(
	w http.ResponseWriter,
	r *http.Request,
) {
	limit, ok := parseSupplyLimit(
		w,
		r,
	)

	if !ok {
		return
	}

	status := strings.TrimSpace(
		r.URL.Query().Get(
			"estado",
		),
	)

	result, err :=
		handler.service.SupplyOrders(
			r.Context(),
			status,
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

func parseSupplyLimit(
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
		return 20, true
	}

	limit, err := strconv.Atoi(
		value,
	)

	if err != nil ||
		limit < 1 ||
		limit > 100 {
		WriteError(
			w,
			http.StatusBadRequest,
			"invalid_supply_limit",
			"El límite debe ser un número entre 1 y 100.",
		)

		return 0, false
	}

	return limit, true
}
