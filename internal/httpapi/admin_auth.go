package httpapi

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"

	"supermarket-sig-api/internal/httpapi/handlers"

	"github.com/golang-jwt/jwt/v5"
)

// AdminIdentity representa al usuario administrativo
// autenticado mediante el JWT emitido por NestJS.
type AdminIdentity struct {
	AccountID int64
	Username  string
	Profile   string
}

type adminIdentityContextKey struct{}

// AdminIdentityFromContext obtiene la identidad administrativa
// validada por el middleware.
func AdminIdentityFromContext(
	ctx context.Context,
) (
	AdminIdentity,
	bool,
) {
	identity, ok :=
		ctx.Value(
			adminIdentityContextKey{},
		).(AdminIdentity)

	return identity,
		ok
}

// AdminAuth valida los tokens administrativos emitidos por NestJS.
type AdminAuth struct {
	secret       []byte
	allowedRoles map[string]struct{}
}

// NewAdminAuth crea el validador JWT administrativo.
func NewAdminAuth(
	secret string,
	allowedRoles []string,
) *AdminAuth {
	roles := make(
		map[string]struct{},
		len(allowedRoles),
	)

	for _, role := range allowedRoles {
		normalizedRole :=
			strings.TrimSpace(
				role,
			)

		if normalizedRole == "" {
			continue
		}

		roles[normalizedRole] =
			struct{}{}
	}

	return &AdminAuth{
		secret: []byte(
			strings.TrimSpace(
				secret,
			),
		),

		allowedRoles: roles,
	}
}

// Middleware exige un JWT administrativo válido y un perfil autorizado.
func (
	auth *AdminAuth,
) Middleware(
	next http.Handler,
) http.Handler {
	return http.HandlerFunc(
		func(
			w http.ResponseWriter,
			r *http.Request,
		) {
			if len(auth.secret) == 0 {
				handlers.WriteError(
					w,
					http.StatusServiceUnavailable,
					"admin_auth_not_configured",
					"La autenticación administrativa no está configurada.",
				)

				return
			}

			rawToken, ok :=
				extractBearerToken(
					r.Header.Get(
						"Authorization",
					),
				)

			if !ok {
				writeInvalidAccessToken(
					w,
				)

				return
			}

			claims :=
				jwt.MapClaims{}

			token, err :=
				jwt.ParseWithClaims(
					rawToken,
					claims,
					func(
						token *jwt.Token,
					) (
						any,
						error,
					) {
						if token.Method.Alg() !=
							jwt.SigningMethodHS256.Alg() {
							return nil,
								fmt.Errorf(
									"algoritmo JWT no permitido: %s",
									token.Method.Alg(),
								)
						}

						return auth.secret,
							nil
					},
					jwt.WithValidMethods(
						[]string{
							jwt.SigningMethodHS256.Alg(),
						},
					),
					jwt.WithExpirationRequired(),
					jwt.WithJSONNumber(),
				)

			if err != nil ||
				token == nil ||
				!token.Valid {
				writeInvalidAccessToken(
					w,
				)

				return
			}

			identity, ok :=
				adminIdentityFromClaims(
					claims,
				)

			if !ok {
				writeInvalidAccessToken(
					w,
				)

				return
			}

			if !auth.isRoleAllowed(
				identity.Profile,
			) {
				handlers.WriteError(
					w,
					http.StatusForbidden,
					"insufficient_role",
					"El usuario no tiene permisos para operar este recurso.",
				)

				return
			}

			ctx :=
				context.WithValue(
					r.Context(),
					adminIdentityContextKey{},
					identity,
				)

			next.ServeHTTP(
				w,
				r.WithContext(
					ctx,
				),
			)
		},
	)
}

func (
	auth *AdminAuth,
) isRoleAllowed(
	profile string,
) bool {
	_, allowed :=
		auth.allowedRoles[strings.TrimSpace(
			profile,
		)]

	return allowed
}

func extractBearerToken(
	authorizationHeader string,
) (
	string,
	bool,
) {
	parts :=
		strings.Fields(
			authorizationHeader,
		)

	if len(parts) != 2 ||
		!strings.EqualFold(
			parts[0],
			"Bearer",
		) {
		return "",
			false
	}

	token :=
		strings.TrimSpace(
			parts[1],
		)

	if token == "" {
		return "",
			false
	}

	return token,
		true
}

func adminIdentityFromClaims(
	claims jwt.MapClaims,
) (
	AdminIdentity,
	bool,
) {
	tokenType, ok :=
		requiredStringClaim(
			claims,
			"tokenType",
		)

	if !ok ||
		tokenType != "admin" {
		return AdminIdentity{},
			false
	}

	if hasInvalidPurpose(
		claims,
	) {
		return AdminIdentity{},
			false
	}

	accountID, ok :=
		nonNegativeIntegerClaim(
			claims["sub"],
		)

	if !ok {
		return AdminIdentity{},
			false
	}

	profile, ok :=
		requiredStringClaim(
			claims,
			"perfil",
		)

	if !ok {
		return AdminIdentity{},
			false
	}

	username, _ :=
		optionalStringClaim(
			claims,
			"username",
		)

	return AdminIdentity{
			AccountID: accountID,
			Username:  username,
			Profile:   profile,
		},
		true
}

func requiredStringClaim(
	claims jwt.MapClaims,
	key string,
) (
	string,
	bool,
) {
	value, exists :=
		claims[key]

	if !exists {
		return "",
			false
	}

	text, ok :=
		value.(string)

	if !ok {
		return "",
			false
	}

	text =
		strings.TrimSpace(
			text,
		)

	if text == "" {
		return "",
			false
	}

	return text,
		true
}

func optionalStringClaim(
	claims jwt.MapClaims,
	key string,
) (
	string,
	bool,
) {
	value, exists :=
		claims[key]

	if !exists ||
		value == nil {
		return "",
			false
	}

	text, ok :=
		value.(string)

	if !ok {
		return "",
			false
	}

	return strings.TrimSpace(
			text,
		),
		true
}

func hasInvalidPurpose(
	claims jwt.MapClaims,
) bool {
	value, exists :=
		claims["purpose"]

	if !exists ||
		value == nil {
		return false
	}

	purpose, ok :=
		value.(string)

	if !ok {
		return true
	}

	return strings.TrimSpace(
		purpose,
	) != ""
}

func nonNegativeIntegerClaim(
	value any,
) (
	int64,
	bool,
) {
	switch typedValue :=
		value.(type) {
	case json.Number:
		number, err :=
			strconv.ParseInt(
				typedValue.String(),
				10,
				64,
			)

		if err != nil ||
			number < 0 {
			return 0,
				false
		}

		return number,
			true

	case string:
		number, err :=
			strconv.ParseInt(
				strings.TrimSpace(
					typedValue,
				),
				10,
				64,
			)

		if err != nil ||
			number < 0 {
			return 0,
				false
		}

		return number,
			true

	case float64:
		if math.IsNaN(
			typedValue,
		) ||
			math.IsInf(
				typedValue,
				0,
			) ||
			typedValue < 0 ||
			math.Trunc(
				typedValue,
			) != typedValue ||
			typedValue > float64(
				1<<53-1,
			) {
			return 0,
				false
		}

		return int64(
				typedValue,
			),
			true

	default:
		return 0,
			false
	}
}

func writeInvalidAccessToken(
	w http.ResponseWriter,
) {
	handlers.WriteError(
		w,
		http.StatusUnauthorized,
		"invalid_access_token",
		"El token de acceso no es válido.",
	)
}
