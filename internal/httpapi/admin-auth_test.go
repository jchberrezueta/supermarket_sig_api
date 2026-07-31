package httpapi

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestAdminAuthMiddleware(t *testing.T) {
	const secret = "secreto-pruebas-jwt"

	auth := NewAdminAuth(
		secret,
		[]string{
			"padmin",
			"pgerente",
		},
	)

	protectedHandler :=
		auth.Middleware(
			http.HandlerFunc(
				func(
					w http.ResponseWriter,
					r *http.Request,
				) {
					identity, ok :=
						AdminIdentityFromContext(
							r.Context(),
						)

					if !ok {
						http.Error(
							w,
							"identidad ausente",
							http.StatusInternalServerError,
						)

						return
					}

					w.Header().Set(
						"X-Test-Account-ID",
						"25",
					)

					w.Header().Set(
						"X-Test-Username",
						identity.Username,
					)

					w.Header().Set(
						"X-Test-Profile",
						identity.Profile,
					)

					w.WriteHeader(
						http.StatusNoContent,
					)
				},
			),
		)

	tests := []struct {
		name           string
		authorization  string
		expectedStatus int
		expectedRole   string
	}{
		{
			name:
				"sin cabecera Authorization",

			expectedStatus:
				http.StatusUnauthorized,
		},
		{
			name:
				"esquema distinto de Bearer",

			authorization:
				"Basic abc123",

			expectedStatus:
				http.StatusUnauthorized,
		},
		{
			name:
				"token firmado con otro secreto",

			authorization:
				"Bearer " +
					signAdminTestToken(
						t,
						"otro-secreto",
						validAdminTestClaims(
							"padmin",
						),
					),

			expectedStatus:
				http.StatusUnauthorized,
		},
		{
			name:
				"token vencido",

			authorization:
				"Bearer " +
					signAdminTestToken(
						t,
						secret,
						jwt.MapClaims{
							"sub":       25,
							"username":  "admin",
							"perfil":    "padmin",
							"tokenType": "admin",
							"exp": time.Now().
								Add(-time.Minute).
								Unix(),
						},
					),

			expectedStatus:
				http.StatusUnauthorized,
		},
		{
			name:
				"token de refresco",

			authorization:
				"Bearer " +
					signAdminTestToken(
						t,
						secret,
						jwt.MapClaims{
							"sub":       25,
							"username":  "admin",
							"perfil":    "padmin",
							"tokenType": "admin",
							"purpose":   "refresh",
							"exp": time.Now().
								Add(time.Hour).
								Unix(),
						},
					),

			expectedStatus:
				http.StatusUnauthorized,
		},
		{
			name:
				"tipo de token distinto de admin",

			authorization:
				"Bearer " +
					signAdminTestToken(
						t,
						secret,
						jwt.MapClaims{
							"sub":       25,
							"username":  "admin",
							"perfil":    "padmin",
							"tokenType": "mobile",
							"exp": time.Now().
								Add(time.Hour).
								Unix(),
						},
					),

			expectedStatus:
				http.StatusUnauthorized,
		},
		{
			name:
				"sub negativo",

			authorization:
				"Bearer " +
					signAdminTestToken(
						t,
						secret,
						jwt.MapClaims{
							"sub":       -1,
							"username":  "admin",
							"perfil":    "padmin",
							"tokenType": "admin",
							"exp": time.Now().
								Add(time.Hour).
								Unix(),
						},
					),

			expectedStatus:
				http.StatusUnauthorized,
		},
		{
			name:
				"perfil no autorizado",

			authorization:
				"Bearer " +
					signAdminTestToken(
						t,
						secret,
						validAdminTestClaims(
							"pcliente",
						),
					),

			expectedStatus:
				http.StatusForbidden,
		},
		{
			name:
				"padmin autorizado",

			authorization:
				"Bearer " +
					signAdminTestToken(
						t,
						secret,
						validAdminTestClaims(
							"padmin",
						),
					),

			expectedStatus:
				http.StatusNoContent,

			expectedRole:
				"padmin",
		},
		{
			name:
				"pgerente autorizado",

			authorization:
				"Bearer " +
					signAdminTestToken(
						t,
						secret,
						validAdminTestClaims(
							"pgerente",
						),
					),

			expectedStatus:
				http.StatusNoContent,

			expectedRole:
				"pgerente",
		},
	}

	for _, test := range tests {
		t.Run(
			test.name,
			func(t *testing.T) {
				request :=
					httptest.NewRequest(
						http.MethodPatch,
						"/api/sig/iot/incidentes/1/reconocer",
						nil,
					)

				if test.authorization != "" {
					request.Header.Set(
						"Authorization",
						test.authorization,
					)
				}

				response :=
					httptest.NewRecorder()

				protectedHandler.ServeHTTP(
					response,
					request,
				)

				if response.Code !=
					test.expectedStatus {
					t.Fatalf(
						"estado HTTP inesperado: se esperaba %d y se obtuvo %d; cuerpo: %s",
						test.expectedStatus,
						response.Code,
						response.Body.String(),
					)
				}

				if test.expectedRole == "" {
					return
				}

				if response.Header().Get(
					"X-Test-Account-ID",
				) != "25" {
					t.Fatalf(
						"identificador de cuenta inesperado: %q",
						response.Header().Get(
							"X-Test-Account-ID",
						),
					)
				}

				if response.Header().Get(
					"X-Test-Username",
				) != "admin" {
					t.Fatalf(
						"nombre de usuario inesperado: %q",
						response.Header().Get(
							"X-Test-Username",
						),
					)
				}

				if response.Header().Get(
					"X-Test-Profile",
				) != test.expectedRole {
					t.Fatalf(
						"perfil inesperado: %q",
						response.Header().Get(
							"X-Test-Profile",
						),
					)
				}
			},
		)
	}
}

func TestAdminAuthMiddlewareWithoutSecret(
	t *testing.T,
) {
	auth :=
		NewAdminAuth(
			"",
			[]string{
				"padmin",
			},
		)

	handler :=
		auth.Middleware(
			http.HandlerFunc(
				func(
					w http.ResponseWriter,
					_ *http.Request,
				) {
					w.WriteHeader(
						http.StatusNoContent,
					)
				},
			),
		)

	request :=
		httptest.NewRequest(
			http.MethodGet,
			"/api/sig/iot/incidentes",
			nil,
		)

	response :=
		httptest.NewRecorder()

	handler.ServeHTTP(
		response,
		request,
	)

	if response.Code !=
		http.StatusServiceUnavailable {
		t.Fatalf(
			"se esperaba HTTP %d y se obtuvo %d; cuerpo: %s",
			http.StatusServiceUnavailable,
			response.Code,
			response.Body.String(),
		)
	}
}

func validAdminTestClaims(
	profile string,
) jwt.MapClaims {
	return jwt.MapClaims{
		"sub":       25,
		"username":  "admin",
		"perfil":    profile,
		"tokenType": "admin",
		"exp": time.Now().
			Add(time.Hour).
			Unix(),
	}
}

func signAdminTestToken(
	t *testing.T,
	secret string,
	claims jwt.MapClaims,
) string {
	t.Helper()

	token :=
		jwt.NewWithClaims(
			jwt.SigningMethodHS256,
			claims,
		)

	signedToken, err :=
		token.SignedString(
			[]byte(
				secret,
			),
		)

	if err != nil {
		t.Fatalf(
			"no se pudo firmar el JWT de prueba: %v",
			err,
		)
	}

	return signedToken
}