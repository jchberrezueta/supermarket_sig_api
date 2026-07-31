package config

import (
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// DatabaseConfig contiene la configuración de Oracle.
type DatabaseConfig struct {
	Enabled         bool
	Host            string
	Port            string
	Service         string
	User            string
	Password        string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

// IoTConfig contiene la configuración del dispositivo y
// los parámetros permitidos de cadena de frío.
type IoTConfig struct {
	DeviceKey      string
	DeviceCode     string
	TemperatureMin float64
	TemperatureMax float64
}

// AuthConfig contiene la configuración para validar
// los tokens administrativos emitidos por NestJS.
type AuthConfig struct {
	JWTSecret    string
	AllowedRoles []string
}

// IntegrationConfig contiene la configuración de integración con el ERP.
type IntegrationConfig struct {
	SyncKey          string
	ERPBaseURL       string
	Timeout          time.Duration
	AutoSyncEnabled  bool
	AutoSyncInterval time.Duration
	InitialSyncDelay time.Duration
}

// Config contiene la configuración general de la API SIG.
type Config struct {
	AppEnvironment string
	Port           string
	CORSOrigins    []string
	Database       DatabaseConfig
	IoT            IoTConfig
	Auth           AuthConfig
	Integration    IntegrationConfig
}

// Load carga las variables desde .env y desde el entorno.
func Load() Config {
	_ = godotenv.Load()

	lifetimeMinutes := getEnvInt(
		"DB_CONNECTION_MAX_LIFETIME_MINUTES",
		30,
	)

	erpTimeoutSeconds := getEnvInt(
		"ERP_API_TIMEOUT_SECONDS",
		20,
	)

	autoSyncIntervalSeconds := getEnvInt(
		"ERP_AUTO_SYNC_INTERVAL_SECONDS",
		60,
	)

	initialSyncDelaySeconds := getEnvInt(
		"ERP_INITIAL_SYNC_DELAY_SECONDS",
		5,
	)

	return Config{
		AppEnvironment: getEnv(
			"APP_ENV",
			"development",
		),
		Port: getEnv(
			"API_PORT",
			"8080",
		),
		CORSOrigins: splitCSV(
			getEnv(
				"CORS_ORIGINS",
				"http://localhost:4200",
			),
		),
		Database: DatabaseConfig{
			Enabled: getEnvBool(
				"DB_ENABLED",
				false,
			),
			Host: getEnv(
				"DB_HOST",
				"127.0.0.1",
			),
			Port: getEnv(
				"DB_PORT",
				"1522",
			),
			Service: getEnv(
				"DB_SERVICE",
				"FREEPDB1",
			),
			User: getEnv(
				"DB_USER",
				"SUPERMARKET",
			),
			Password: getEnv(
				"DB_PASSWORD",
				"",
			),
			MaxOpenConns: getEnvInt(
				"DB_MAX_OPEN_CONNECTIONS",
				10,
			),
			MaxIdleConns: getEnvInt(
				"DB_MAX_IDLE_CONNECTIONS",
				5,
			),
			ConnMaxLifetime: time.Duration(
				lifetimeMinutes,
			) * time.Minute,
		},
		IoT: IoTConfig{
			DeviceKey: getEnv(
				"DEVICE_KEY",
				"",
			),
			DeviceCode: getEnv(
				"IOT_DEVICE_CODE",
				"ESP32-BODEGA-01",
			),
			TemperatureMin: getEnvFloat(
				"IOT_TEMPERATURE_MIN",
				2,
			),
			TemperatureMax: getEnvFloat(
				"IOT_TEMPERATURE_MAX",
				8,
			),
		},
		Auth: AuthConfig{
			JWTSecret: getEnv(
				"JWT_SECRET",
				"",
			),
			AllowedRoles: splitCSV(
				getEnv(
					"SIG_ALLOWED_ROLES",
					"padmin,pgerente",
				),
			),
		},
		Integration: IntegrationConfig{
			SyncKey: getEnv(
				"SIG_SYNC_KEY",
				"",
			),
			ERPBaseURL: getEnv(
				"ERP_API_BASE_URL",
				"http://localhost:3001/api",
			),
			Timeout: time.Duration(
				erpTimeoutSeconds,
			) * time.Second,
			AutoSyncEnabled: getEnvBool(
				"ERP_AUTO_SYNC_ENABLED",
				false,
			),
			AutoSyncInterval: time.Duration(
				autoSyncIntervalSeconds,
			) * time.Second,
			InitialSyncDelay: time.Duration(
				initialSyncDelaySeconds,
			) * time.Second,
		},
	}
}

func getEnv(
	key string,
	fallback string,
) string {
	value := strings.TrimSpace(
		os.Getenv(key),
	)

	if value == "" {
		return fallback
	}

	return value
}

func getEnvBool(
	key string,
	fallback bool,
) bool {
	value := strings.ToLower(
		strings.TrimSpace(
			os.Getenv(key),
		),
	)

	switch value {
	case "1", "true", "yes", "si", "sí":
		return true

	case "0", "false", "no":
		return false

	default:
		return fallback
	}
}

func getEnvInt(
	key string,
	fallback int,
) int {
	value := strings.TrimSpace(
		os.Getenv(key),
	)

	if value == "" {
		return fallback
	}

	number, err := strconv.Atoi(value)

	if err != nil || number < 0 {
		return fallback
	}

	return number
}

func getEnvFloat(
	key string,
	fallback float64,
) float64 {
	value := strings.TrimSpace(
		os.Getenv(key),
	)

	if value == "" {
		return fallback
	}

	number, err := strconv.ParseFloat(
		value,
		64,
	)

	if err != nil {
		return fallback
	}

	return number
}

func splitCSV(value string) []string {
	parts := strings.Split(
		value,
		",",
	)

	result := make(
		[]string,
		0,
		len(parts),
	)

	for _, part := range parts {
		item := strings.TrimSpace(part)

		if item != "" {
			result = append(
				result,
				item,
			)
		}
	}

	if len(result) == 0 {
		return []string{
			"http://localhost:4200",
		}
	}

	return result
}
