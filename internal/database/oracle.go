package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net"
	"net/url"

	"supermarket-sig-api/internal/config"

	_ "github.com/sijms/go-ora/v2"
)

// OpenOracle abre y verifica el pool de conexiones Oracle.
func OpenOracle(
	ctx context.Context,
	cfg config.DatabaseConfig,
) (*sql.DB, error) {
	if !cfg.Enabled {
		return nil, errors.New(
			"la conexión Oracle está deshabilitada",
		)
	}

	if cfg.Host == "" {
		return nil, errors.New(
			"DB_HOST es obligatorio",
		)
	}

	if cfg.Port == "" {
		return nil, errors.New(
			"DB_PORT es obligatorio",
		)
	}

	if cfg.Service == "" {
		return nil, errors.New(
			"DB_SERVICE es obligatorio",
		)
	}

	if cfg.User == "" {
		return nil, errors.New(
			"DB_USER es obligatorio",
		)
	}

	if cfg.Password == "" ||
		cfg.Password == "CAMBIAR" {
		return nil, errors.New(
			"DB_PASSWORD no está configurada",
		)
	}

	dsn := oracleDSN(cfg)

	db, err := sql.Open(
		"oracle",
		dsn,
	)

	if err != nil {
		return nil, fmt.Errorf(
			"no se pudo crear el pool Oracle: %w",
			err,
		)
	}

	db.SetMaxOpenConns(
		cfg.MaxOpenConns,
	)

	db.SetMaxIdleConns(
		cfg.MaxIdleConns,
	)

	db.SetConnMaxLifetime(
		cfg.ConnMaxLifetime,
	)

	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()

		return nil, fmt.Errorf(
			"Oracle no respondió: %w",
			err,
		)
	}

	return db, nil
}

func oracleDSN(
	cfg config.DatabaseConfig,
) string {
	connectionURL := &url.URL{
		Scheme: "oracle",
		User: url.UserPassword(
			cfg.User,
			cfg.Password,
		),
		Host: net.JoinHostPort(
			cfg.Host,
			cfg.Port,
		),
		Path: "/" + cfg.Service,
	}

	return connectionURL.String()
}
