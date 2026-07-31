package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"supermarket-sig-api/internal/httpapi/handlers"
	"supermarket-sig-api/internal/iot"
)

func TestIncidentWorkflowRoutesRequireAdminJWT(
	t *testing.T,
) {
	tests := []struct {
		name                   string
		includeToken           bool
		profile                string
		expectedStatus         int
		expectedErrorCode      string
		expectedIncidentStatus string
	}{
		{
			name: "rechaza solicitud sin token",

			includeToken: false,

			expectedStatus: http.StatusUnauthorized,

			expectedErrorCode: "invalid_access_token",

			expectedIncidentStatus: "abierto",
		},
		{
			name: "rechaza perfil no autorizado",

			includeToken: true,

			profile: "pcliente",

			expectedStatus: http.StatusForbidden,

			expectedErrorCode: "insufficient_role",

			expectedIncidentStatus: "abierto",
		},
		{
			name: "permite perfil padmin",

			includeToken: true,

			profile: "padmin",

			expectedStatus: http.StatusOK,

			expectedIncidentStatus: "reconocido",
		},
		{
			name: "permite perfil pgerente",

			includeToken: true,

			profile: "pgerente",

			expectedStatus: http.StatusOK,

			expectedIncidentStatus: "reconocido",
		},
	}

	for _, test := range tests {
		t.Run(
			test.name,
			func(t *testing.T) {
				router :=
					NewRouter(
						iotTestConfig(),
						nil,
					)

				registerOutOfRangeReading(
					t,
					router,
				)

				request :=
					httptest.NewRequest(
						http.MethodPatch,
						"/api/sig/iot/incidentes/1/reconocer",
						strings.NewReader(
							`{
								"responsable": "Gerente de bodega"
							}`,
						),
					)

				request.Header.Set(
					"Content-Type",
					"application/json",
				)

				if test.includeToken {
					token :=
						signAdminTestToken(
							t,
							"secreto-pruebas-jwt",
							validAdminTestClaims(
								test.profile,
							),
						)

					request.Header.Set(
						"Authorization",
						"Bearer "+token,
					)
				}

				response :=
					httptest.NewRecorder()

				router.ServeHTTP(
					response,
					request,
				)

				if response.Code !=
					test.expectedStatus {
					t.Fatalf(
						"se esperaba HTTP %d y se obtuvo %d; cuerpo: %s",
						test.expectedStatus,
						response.Code,
						response.Body.String(),
					)
				}

				if test.expectedErrorCode != "" {
					var errorBody handlers.ErrorResponse

					if err :=
						json.NewDecoder(
							response.Body,
						).Decode(
							&errorBody,
						); err != nil {
						t.Fatalf(
							"no se pudo interpretar la respuesta de error: %v",
							err,
						)
					}

					if errorBody.Success {
						t.Fatal(
							"se esperaba success=false",
						)
					}

					if errorBody.Error.Code !=
						test.expectedErrorCode {
						t.Fatalf(
							"código de error inesperado: se esperaba %q y se obtuvo %q",
							test.expectedErrorCode,
							errorBody.Error.Code,
						)
					}
				}

				detailRequest :=
					httptest.NewRequest(
						http.MethodGet,
						"/api/sig/iot/incidentes/1",
						nil,
					)

				detailResponse :=
					httptest.NewRecorder()

				router.ServeHTTP(
					detailResponse,
					detailRequest,
				)

				if detailResponse.Code !=
					http.StatusOK {
					t.Fatalf(
						"no se pudo consultar el incidente: HTTP %d; cuerpo: %s",
						detailResponse.Code,
						detailResponse.Body.String(),
					)
				}

				var detailBody struct {
					Success bool               `json:"success"`
					Data    iot.IncidentDetail `json:"data"`
				}

				if err :=
					json.NewDecoder(
						detailResponse.Body,
					).Decode(
						&detailBody,
					); err != nil {
					t.Fatalf(
						"no se pudo interpretar el detalle del incidente: %v",
						err,
					)
				}

				if !detailBody.Success {
					t.Fatal(
						"se esperaba success=true al consultar el incidente",
					)
				}

				if detailBody.Data.Incident.Status !=
					test.expectedIncidentStatus {
					t.Fatalf(
						"estado del incidente inesperado: se esperaba %q y se obtuvo %q",
						test.expectedIncidentStatus,
						detailBody.Data.Incident.Status,
					)
				}

				if test.expectedIncidentStatus ==
					"reconocido" {
					if detailBody.Data.Incident.Responsible !=
						"Gerente de bodega" {
						t.Fatalf(
							"responsable inesperado: %q",
							detailBody.Data.Incident.Responsible,
						)
					}

					if detailBody.Data.Incident.RecognizedAt ==
						nil {
						t.Fatal(
							"falta la fecha de reconocimiento",
						)
					}
				}
			},
		)
	}
}
