package httpapi

import (
	"database/sql"
	"net/http"

	"supermarket-sig-api/internal/config"
	"supermarket-sig-api/internal/erpclient"
	"supermarket-sig-api/internal/erpdata"
	"supermarket-sig-api/internal/httpapi/handlers"
	"supermarket-sig-api/internal/iot"
	"supermarket-sig-api/internal/management"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

// Runtime contiene el router y las dependencias compartidas
// necesarias durante la ejecución de la API.
type Runtime struct {
	Handler    http.Handler
	ERPService *erpdata.Service
	ERPSource  erpdata.SnapshotSource
	IoTService *iot.Service
}

// NewRuntime construye las dependencias compartidas de la API.
func NewRuntime(
	cfg config.Config,
	db *sql.DB,
) Runtime {
	var erpRepository erpdata.Repository

	if cfg.Database.Enabled &&
		db != nil {
		erpRepository =
			erpdata.NewOracleRepository(
				db,
			)
	} else {
		erpRepository =
			erpdata.NewMemoryRepository()
	}

	erpService :=
		erpdata.NewService(
			erpRepository,
		)

	erpSource :=
		erpclient.NewClient(
			cfg.Integration.ERPBaseURL,
			cfg.Integration.SyncKey,
			cfg.Integration.Timeout,
		)

	var iotRepository iot.Repository

	if cfg.Database.Enabled &&
		db != nil {
		iotRepository =
			iot.NewOracleRepository(
				db,
				cfg.IoT.DeviceCode,
				cfg.IoT.TemperatureMin,
				cfg.IoT.TemperatureMax,
			)
	} else {
		iotRepository =
			iot.NewMemoryRepository()
	}

	iotService :=
		iot.NewService(
			iotRepository,
			cfg.IoT.DeviceCode,
			cfg.IoT.TemperatureMin,
			cfg.IoT.TemperatureMax,
		)

	return Runtime{
		Handler: newRouter(
			cfg,
			db,
			erpService,
			erpSource,
			iotService,
		),

		ERPService: erpService,
		ERPSource:  erpSource,
		IoTService: iotService,
	}
}

// NewRouter construye las rutas y middlewares de la API.
// Se conserva para pruebas y usos que solo necesitan el handler HTTP.
func NewRouter(
	cfg config.Config,
	db *sql.DB,
) http.Handler {
	return NewRuntime(
		cfg,
		db,
	).Handler
}

func newRouter(
	cfg config.Config,
	db *sql.DB,
	erpService *erpdata.Service,
	erpSource erpdata.SnapshotSource,
	iotService *iot.Service,
) http.Handler {
	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	router.Use(
		cors.Handler(
			cors.Options{
				AllowedOrigins: cfg.CORSOrigins,

				AllowedMethods: []string{
					http.MethodGet,
					http.MethodPost,
					http.MethodPut,
					http.MethodPatch,
					http.MethodDelete,
					http.MethodOptions,
				},

				AllowedHeaders: []string{
					"Accept",
					"Authorization",
					"Content-Type",
					"X-Device-Key",
					"X-SIG-Key",
				},

				ExposedHeaders: []string{
					"X-Request-ID",
				},

				AllowCredentials: false,
				MaxAge:           300,
			},
		),
	)

	router.NotFound(
		handlers.NotFound,
	)

	router.MethodNotAllowed(
		handlers.MethodNotAllowed,
	)

	iotHandler := handlers.NewIoTHandler(
		iotService,
		cfg.IoT.DeviceKey,
	)

	adminAuth :=
		NewAdminAuth(
			cfg.Auth.JWTSecret,
			cfg.Auth.AllowedRoles,
		)

	integrationHandler :=
		handlers.NewIntegrationHandler(
			erpService,
			erpSource,
			cfg.Integration.SyncKey,
		)

	managementService :=
		management.NewService(
			erpService,
			iotService,
		)

	managementHandler :=
		handlers.NewManagementHandler(
			managementService,
		)

	router.Get(
		"/health",
		handlers.Health(
			cfg.AppEnvironment,
		),
	)

	router.Get(
		"/db-health",
		handlers.DatabaseHealth(
			db,
			cfg.Database.Enabled,
		),
	)

	router.Route(
		"/api/sig",
		func(router chi.Router) {
			router.Get(
				"/health",
				handlers.Health(
					cfg.AppEnvironment,
				),
			)

			router.Get(
				"/db-health",
				handlers.DatabaseHealth(
					db,
					cfg.Database.Enabled,
				),
			)

			router.Route(
				"/integracion",
				func(router chi.Router) {
					router.Post(
						"/sincronizar",
						integrationHandler.Synchronize,
					)

					router.Post(
						"/snapshot",
						integrationHandler.ImportSnapshot,
					)

					router.Get(
						"/estado",
						integrationHandler.State,
					)
				},
			)

			router.Route(
				"/inventario",
				func(router chi.Router) {
					router.Get(
						"/resumen",
						managementHandler.InventoryOverview,
					)

					router.Get(
						"/stock-critico",
						managementHandler.CriticalStock,
					)

					router.Get(
						"/caducidad",
						managementHandler.ExpiringLots,
					)

					router.Get(
						"/movimientos",
						managementHandler.InventoryMovements,
					)
				},
			)

			router.Route(
				"/abastecimiento",
				func(router chi.Router) {
					router.Get(
						"/resumen",
						managementHandler.SupplyOverview,
					)

					router.Get(
						"/proveedores",
						managementHandler.SupplierPerformance,
					)

					router.Get(
						"/pedidos",
						managementHandler.SupplyOrders,
					)
				},
			)

			router.Route(
				"/ventas",
				func(router chi.Router) {
					router.Get(
						"/resumen",
						managementHandler.SalesOverview,
					)

					router.Get(
						"/tendencia",
						managementHandler.SalesTrend,
					)

					router.Get(
						"/productos",
						managementHandler.TopSellingProducts,
					)

					router.Get(
						"/categorias",
						managementHandler.SalesByCategory,
					)
				},
			)

			router.Get(
				"/resumen-ejecutivo",
				managementHandler.ExecutiveSummary,
			)

			router.Route(
				"/iot",
				func(router chi.Router) {
					router.Get(
						"/resumen",
						iotHandler.Summary,
					)

					router.Post(
						"/lecturas",
						iotHandler.RegisterReading,
					)

					router.Get(
						"/lecturas",
						iotHandler.ListReadings,
					)

					router.Get(
						"/lecturas/ultima",
						iotHandler.LatestReading,
					)

					router.Get(
						"/alertas",
						iotHandler.ListAlerts,
					)

					router.Get(
						"/incidentes",
						iotHandler.ListIncidents,
					)

					router.Get(
						"/incidentes/{incidentID}",
						iotHandler.IncidentDetail,
					)

					router.Group(
						func(router chi.Router) {
							router.Use(
								adminAuth.Middleware,
							)

							router.Patch(
								"/incidentes/{incidentID}/reconocer",
								iotHandler.RecognizeIncident,
							)

							router.Post(
								"/incidentes/{incidentID}/acciones",
								iotHandler.AddCorrectiveAction,
							)

							router.Patch(
								"/incidentes/{incidentID}/resolver",
								iotHandler.ResolveIncident,
							)

							router.Patch(
								"/incidentes/{incidentID}/cerrar",
								iotHandler.CloseIncident,
							)
						},
					)
				},
			)

			router.Get(
				"/auditoria",
				iotHandler.ListAudit,
			)
		},
	)

	return router
}
