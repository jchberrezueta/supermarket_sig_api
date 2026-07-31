package iot

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

// lockOracleDevice bloquea el dispositivo dentro de la transacción.
//
// Todas las lecturas de un mismo dispositivo deben pasar por este bloqueo.
// Así, dos solicitudes simultáneas no pueden comprobar y crear incidentes
// al mismo tiempo.
func lockOracleDevice(
	ctx context.Context,
	tx *sql.Tx,
	deviceCode string,
) error {
	const query = `
SELECT codigo
FROM sig_dispositivo_iot
WHERE codigo = :1
FOR UPDATE`

	var storedDeviceCode string

	err := tx.QueryRowContext(
		ctx,
		query,
		deviceCode,
	).Scan(
		&storedDeviceCode,
	)

	if errors.Is(
		err,
		sql.ErrNoRows,
	) {
		return fmt.Errorf(
			"el dispositivo IoT %q no existe en Oracle",
			deviceCode,
		)
	}

	if err != nil {
		return fmt.Errorf(
			"no se pudo bloquear el dispositivo IoT %q en Oracle: %w",
			deviceCode,
			err,
		)
	}

	return nil
}

// hasActiveIncidentForDevice comprueba si el dispositivo ya tiene
// un incidente que todavía no ha terminado su ciclo BPM.
func hasActiveIncidentForDevice(
	ctx context.Context,
	tx *sql.Tx,
	deviceCode string,
) (
	bool,
	error,
) {
	const query = `
SELECT COUNT(*)
FROM (
    SELECT incidente.id
    FROM sig_incidente incidente

    INNER JOIN sig_alerta alerta
            ON alerta.id = incidente.id_alerta

    INNER JOIN sig_lectura_iot lectura
            ON lectura.id = alerta.id_lectura

    WHERE lectura.codigo_dispositivo = :1
      AND incidente.estado IN (
          'abierto',
          'reconocido',
          'en_tratamiento',
          'resuelto'
      )
      AND ROWNUM = 1
)`

	var total int

	err := tx.QueryRowContext(
		ctx,
		query,
		deviceCode,
	).Scan(
		&total,
	)

	if err != nil {
		return false,
			fmt.Errorf(
				"no se pudo comprobar si el dispositivo %q tiene un incidente activo: %w",
				deviceCode,
				err,
			)
	}

	return total > 0,
		nil
}
