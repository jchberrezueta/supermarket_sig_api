package iot

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

type oracleRowScanner interface {
	Scan(
		dest ...any,
	) error
}

// Latest obtiene la lectura más reciente de un dispositivo.
func (repository *OracleRepository) Latest(
	ctx context.Context,
	deviceCode string,
) (
	Reading,
	bool,
	error,
) {
	if err := repository.validate(); err != nil {
		return Reading{},
			false,
			err
	}

	deviceCode = strings.TrimSpace(
		deviceCode,
	)

	if deviceCode == "" {
		deviceCode = repository.deviceCode
	}

	const query = `
SELECT id,
       codigo_dispositivo,
       temperatura,
       humedad,
       estado,
       fecha_lectura,
       fecha_registro
FROM (
    SELECT id,
           codigo_dispositivo,
           temperatura,
           humedad,
           estado,
           fecha_lectura,
           fecha_registro
    FROM sig_lectura_iot
    WHERE codigo_dispositivo = :1
    ORDER BY fecha_lectura DESC,
             id DESC
)
WHERE ROWNUM = 1`

	row := repository.db.QueryRowContext(
		ctx,
		query,
		deviceCode,
	)

	reading, err := scanOracleReading(
		row,
	)

	if errors.Is(
		err,
		sql.ErrNoRows,
	) {
		return Reading{},
			false,
			nil
	}

	if err != nil {
		return Reading{},
			false,
			fmt.Errorf(
				"no se pudo consultar la última lectura IoT en Oracle: %w",
				err,
			)
	}

	return reading,
		true,
		nil
}

// ListReadings devuelve las lecturas más recientes primero.
func (repository *OracleRepository) ListReadings(
	ctx context.Context,
	deviceCode string,
	limit int,
) (
	[]Reading,
	error,
) {
	if err := repository.validate(); err != nil {
		return nil, err
	}

	deviceCode = strings.TrimSpace(
		deviceCode,
	)

	if deviceCode == "" {
		deviceCode = repository.deviceCode
	}

	const query = `
SELECT id,
       codigo_dispositivo,
       temperatura,
       humedad,
       estado,
       fecha_lectura,
       fecha_registro
FROM (
    SELECT id,
           codigo_dispositivo,
           temperatura,
           humedad,
           estado,
           fecha_lectura,
           fecha_registro
    FROM sig_lectura_iot
    WHERE codigo_dispositivo = :1
    ORDER BY fecha_lectura DESC,
             id DESC
)
WHERE ROWNUM <= :2`

	rows, err := repository.db.QueryContext(
		ctx,
		query,
		deviceCode,
		limit,
	)

	if err != nil {
		return nil,
			fmt.Errorf(
				"no se pudo consultar el historial IoT en Oracle: %w",
				err,
			)
	}

	defer rows.Close()

	readings := make(
		[]Reading,
		0,
	)

	for rows.Next() {
		reading, err := scanOracleReading(
			rows,
		)

		if err != nil {
			return nil,
				fmt.Errorf(
					"no se pudo leer una lectura IoT desde Oracle: %w",
					err,
				)
		}

		readings = append(
			readings,
			reading,
		)
	}

	if err := rows.Err(); err != nil {
		return nil,
			fmt.Errorf(
				"falló la consulta del historial IoT en Oracle: %w",
				err,
			)
	}

	return readings,
		nil
}

func scanOracleReading(
	scanner oracleRowScanner,
) (
	Reading,
	error,
) {
	var reading Reading
	var humidity sql.NullFloat64

	err := scanner.Scan(
		&reading.ID,
		&reading.DeviceCode,
		&reading.Temperature,
		&humidity,
		&reading.Status,
		&reading.ReadAt,
		&reading.CreatedAt,
	)

	if err != nil {
		return Reading{}, err
	}

	if humidity.Valid {
		value := humidity.Float64
		reading.Humidity = &value
	}

	return reading,
		nil
}
