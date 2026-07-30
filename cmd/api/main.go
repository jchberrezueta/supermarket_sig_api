package main

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"supermarket-sig-api/internal/config"
	"supermarket-sig-api/internal/database"
	"supermarket-sig-api/internal/httpapi"
)

func main() {
	cfg := config.Load()

	db := initializeDatabase(cfg)

	server := &http.Server{
		Addr: ":" + cfg.Port,

		Handler: httpapi.NewRouter(
			cfg,
			db,
		),

		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	serverErrors := make(
		chan error,
		1,
	)

	go func() {
		log.Printf(
			"SuperMarket SIG API iniciada en http://localhost:%s",
			cfg.Port,
		)

		log.Printf(
			"Health: http://localhost:%s/api/sig/health",
			cfg.Port,
		)

		log.Printf(
			"DB Health: http://localhost:%s/api/sig/db-health",
			cfg.Port,
		)

		serverErrors <- server.ListenAndServe()
	}()

	shutdownSignal := make(
		chan os.Signal,
		1,
	)

	signal.Notify(
		shutdownSignal,
		os.Interrupt,
		syscall.SIGTERM,
	)

	select {
	case signal := <-shutdownSignal:
		log.Printf(
			"Se recibió la señal %s. Cerrando API...",
			signal,
		)

	case err := <-serverErrors:
		if !errors.Is(
			err,
			http.ErrServerClosed,
		) {
			log.Fatalf(
				"No se pudo ejecutar la API: %v",
				err,
			)
		}
	}

	shutdownServer(
		server,
	)

	closeDatabase(
		db,
	)

	log.Println(
		"SuperMarket SIG API detenida correctamente.",
	)
}

func initializeDatabase(
	cfg config.Config,
) *sql.DB {
	if !cfg.Database.Enabled {
		log.Println(
			"Oracle está deshabilitado para este entorno.",
		)

		return nil
	}

	ctx, cancel := context.WithTimeout(
		context.Background(),
		15*time.Second,
	)

	defer cancel()

	db, err := database.OpenOracle(
		ctx,
		cfg.Database,
	)

	if err != nil {
		log.Fatalf(
			"No se pudo iniciar Oracle: %v",
			err,
		)
	}

	log.Println(
		"Conexión Oracle establecida correctamente.",
	)

	return db
}

func shutdownServer(
	server *http.Server,
) {
	ctx, cancel := context.WithTimeout(
		context.Background(),
		10*time.Second,
	)

	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf(
			"El cierre controlado falló: %v",
			err,
		)

		_ = server.Close()
	}
}

func closeDatabase(
	db *sql.DB,
) {
	if db == nil {
		return
	}

	if err := db.Close(); err != nil {
		log.Printf(
			"No se pudo cerrar Oracle correctamente: %v",
			err,
		)

		return
	}

	log.Println(
		"Pool de conexiones Oracle cerrado.",
	)
}
