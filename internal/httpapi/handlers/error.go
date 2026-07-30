package handlers

import (
	"net/http"
)

// ErrorDetail contiene la información de un error controlado.
type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// ErrorResponse representa una respuesta fallida de la API.
type ErrorResponse struct {
	Success bool        `json:"success"`
	Error   ErrorDetail `json:"error"`
}

// WriteError escribe un error controlado en formato JSON.
func WriteError(
	w http.ResponseWriter,
	status int,
	code string,
	message string,
) {
	WriteJSON(
		w,
		status,
		ErrorResponse{
			Success: false,
			Error: ErrorDetail{
				Code:    code,
				Message: message,
			},
		},
	)
}

// NotFound responde cuando la ruta solicitada no existe.
func NotFound(
	w http.ResponseWriter,
	_ *http.Request,
) {
	WriteError(
		w,
		http.StatusNotFound,
		"route_not_found",
		"La ruta solicitada no existe.",
	)
}

// MethodNotAllowed responde cuando la ruta existe,
// pero no acepta el método HTTP utilizado.
func MethodNotAllowed(
	w http.ResponseWriter,
	_ *http.Request,
) {
	WriteError(
		w,
		http.StatusMethodNotAllowed,
		"method_not_allowed",
		"El método HTTP no está permitido para esta ruta.",
	)
}
