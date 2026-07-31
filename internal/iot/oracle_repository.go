package iot

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

// OracleRepository almacena el proceso IoT y BPM en Oracle.
type OracleRepository struct {
	db *sql.DB

	deviceCode     string
	temperatureMin float64
	temperatureMax float64
}

// NewOracleRepository crea el repositorio IoT para Oracle.
func NewOracleRepository(
	db *sql.DB,
	deviceCode string,
	temperatureMin float64,
	temperatureMax float64,
) *OracleRepository {
	deviceCode = strings.TrimSpace(
		deviceCode,
	)

	if deviceCode == "" {
		deviceCode = "ESP32-BODEGA-01"
	}

	if temperatureMin >= temperatureMax {
		temperatureMin = 2
		temperatureMax = 8
	}

	return &OracleRepository{
		db:             db,
		deviceCode:     deviceCode,
		temperatureMin: temperatureMin,
		temperatureMax: temperatureMax,
	}
}

// EnsureDevice registra o actualiza el dispositivo configurado.
func (repository *OracleRepository) EnsureDevice(
	ctx context.Context,
) error {
	if err := repository.validate(); err != nil {
		return err
	}

	return repository.ensureDevice(
		ctx,
		repository.db,
	)
}

type oracleExecer interface {
	ExecContext(
		ctx context.Context,
		query string,
		args ...any,
	) (
		sql.Result,
		error,
	)
}

func (repository *OracleRepository) ensureDevice(
	ctx context.Context,
	execer oracleExecer,
) error {
	const statement = `
MERGE INTO sig_dispositivo_iot destino
USING (
    SELECT :1 codigo,
           :2 nombre,
           :3 ubicacion,
           :4 tipo_sensor,
           :5 temperatura_minima,
           :6 temperatura_maxima,
           :7 fecha_actualizacion
    FROM dual
) origen
ON (
    destino.codigo = origen.codigo
)
WHEN MATCHED THEN
    UPDATE SET
        destino.nombre = origen.nombre,
        destino.ubicacion = origen.ubicacion,
        destino.tipo_sensor = origen.tipo_sensor,
        destino.temperatura_minima = origen.temperatura_minima,
        destino.temperatura_maxima = origen.temperatura_maxima,
        destino.activo = 1,
        destino.actualizado_en = origen.fecha_actualizacion
WHEN NOT MATCHED THEN
    INSERT (
        codigo,
        nombre,
        ubicacion,
        tipo_sensor,
        temperatura_minima,
        temperatura_maxima,
        activo,
        creado_en,
        actualizado_en
    )
    VALUES (
        origen.codigo,
        origen.nombre,
        origen.ubicacion,
        origen.tipo_sensor,
        origen.temperatura_minima,
        origen.temperatura_maxima,
        1,
        origen.fecha_actualizacion,
        origen.fecha_actualizacion
    )`

	_, err := execer.ExecContext(
		ctx,
		statement,
		repository.deviceCode,
		"Sensor de cadena de frío",
		"Bodega de productos",
		"DHT22",
		repository.temperatureMin,
		repository.temperatureMax,
		time.Now(),
	)

	if err != nil {
		return fmt.Errorf(
			"no se pudo registrar el dispositivo IoT en Oracle: %w",
			err,
		)
	}

	return nil
}

func (repository *OracleRepository) validate() error {
	if repository == nil {
		return errors.New(
			"el repositorio IoT Oracle no está configurado",
		)
	}

	if repository.db == nil {
		return errors.New(
			"la conexión Oracle del repositorio IoT no está configurada",
		)
	}

	if strings.TrimSpace(
		repository.deviceCode,
	) == "" {
		return errors.New(
			"el código del dispositivo IoT no está configurado",
		)
	}

	if repository.temperatureMin >=
		repository.temperatureMax {
		return errors.New(
			"el rango de temperatura IoT no es válido",
		)
	}

	return nil
}
