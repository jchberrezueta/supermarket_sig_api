package handlers

import (
	"net/http"
	"strconv"
	"strings"
)

// InventoryOverview devuelve el resumen general de inventario.
func (handler *ManagementHandler) InventoryOverview(
	w http.ResponseWriter,
	r *http.Request,
) {
	days, ok := parseExpirationDays(
		w,
		r,
	)

	if !ok {
		return
	}

	result, err :=
		handler.service.InventoryOverview(
			r.Context(),
			days,
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

// CriticalStock devuelve productos agotados o con stock bajo.
func (handler *ManagementHandler) CriticalStock(
	w http.ResponseWriter,
	r *http.Request,
) {
	limit, ok := parseInventoryLimit(
		w,
		r,
	)

	if !ok {
		return
	}

	result, err :=
		handler.service.CriticalStock(
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

// ExpiringLots devuelve lotes caducados o próximos a caducar.
func (handler *ManagementHandler) ExpiringLots(
	w http.ResponseWriter,
	r *http.Request,
) {
	days, ok := parseExpirationDays(
		w,
		r,
	)

	if !ok {
		return
	}

	limit, ok := parseInventoryLimit(
		w,
		r,
	)

	if !ok {
		return
	}

	result, err :=
		handler.service.ExpiringLots(
			r.Context(),
			days,
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

// InventoryMovements devuelve la trazabilidad del inventario.
func (handler *ManagementHandler) InventoryMovements(
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

	limit, ok := parseInventoryLimit(
		w,
		r,
	)

	if !ok {
		return
	}

	movementType := strings.TrimSpace(
		r.URL.Query().Get(
			"tipo",
		),
	)

	result, err :=
		handler.service.InventoryMovements(
			r.Context(),
			period,
			movementType,
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

func parseExpirationDays(
	w http.ResponseWriter,
	r *http.Request,
) (
	int,
	bool,
) {
	value := strings.TrimSpace(
		r.URL.Query().Get(
			"dias",
		),
	)

	if value == "" {
		return 30, true
	}

	days, err := strconv.Atoi(
		value,
	)

	if err != nil ||
		days < 1 ||
		days > 365 {
		WriteError(
			w,
			http.StatusBadRequest,
			"invalid_expiration_days",
			"El número de días debe estar entre 1 y 365.",
		)

		return 0, false
	}

	return days, true
}

func parseInventoryLimit(
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
			"invalid_inventory_limit",
			"El límite debe ser un número entre 1 y 100.",
		)

		return 0, false
	}

	return limit, true
}
