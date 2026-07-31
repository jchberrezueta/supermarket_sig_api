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
	"supermarket-sig-api/internal/erpdata"
	"supermarket-sig-api/internal/httpapi"
)

func main() {
	cfg := config.Load()

	db := initializeDatabase(
		cfg,
	)

	apiRuntime :=
		httpapi.NewRuntime(
			cfg,
			db,
		)

	appContext, stopApplication :=
		signal.NotifyContext(
			context.Background(),
			os.Interrupt,
			syscall.SIGTERM,
		)

	autoSyncDone :=
		startAutoSync(
			appContext,
			cfg,
			apiRuntime,
		)

	server := &http.Server{
		Addr: ":" + cfg.Port,

		Handler: apiRuntime.Handler,

		ReadHeaderTimeout: 5 * time.Second,

		ReadTimeout: 15 * time.Second,

		WriteTimeout: 30 * time.Second,

		IdleTimeout: 60 * time.Second,
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

	var executionError error

	select {
	case <-appContext.Done():
		log.Println(
			"Se recibió una señal de apagado. Cerrando API...",
		)

	case err := <-serverErrors:
		if !errors.Is(
			err,
			http.ErrServerClosed,
		) {
			executionError = err

			log.Printf(
				"No se pudo ejecutar la API: %v",
				err,
			)
		}
	}

	stopApplication()

	shutdownServer(
		server,
	)

	waitForAutoSync(
		autoSyncDone,
	)

	closeDatabase(
		db,
	)

	if executionError != nil {
		log.Println(
			"SuperMarket SIG API detenida debido a un error.",
		)

		os.Exit(
			1,
		)
	}

	log.Println(
		"SuperMarket SIG API detenida correctamente.",
	)
}

func startAutoSync(
	ctx context.Context,
	cfg config.Config,
	apiRuntime httpapi.Runtime,
) <-chan struct{} {
	done := make(
		chan struct{},
	)

	if !cfg.Integration.AutoSyncEnabled {
		close(
			done,
		)

		log.Println(
			"Sincronización automática deshabilitada.",
		)

		return done
	}

	worker :=
		erpdata.NewAutoSyncWorker(
			apiRuntime.ERPService,
			apiRuntime.ERPSource,
			cfg.Integration.InitialSyncDelay,
			cfg.Integration.AutoSyncInterval,
		)

	log.Printf(
		"Sincronización automática habilitada: retraso inicial=%s, intervalo=%s.",
		cfg.Integration.InitialSyncDelay,
		cfg.Integration.AutoSyncInterval,
	)

	go func() {
		defer close(
			done,
		)

		worker.Run(
			ctx,
		)
	}()

	return done
}

func waitForAutoSync(
	done <-chan struct{},
) {
	select {
	case <-done:
		return

	case <-time.After(
		10 * time.Second,
	):
		log.Println(
			"Se agotó el tiempo de espera para detener la sincronización automática.",
		)
	}
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
