package erpdata

import (
	"context"
	"log"
	"time"
)

const defaultAutoSyncInterval = 60 * time.Second

// AutoSyncWorker ejecuta la sincronización periódica con el ERP.
type AutoSyncWorker struct {
	service      *Service
	source       SnapshotSource
	initialDelay time.Duration
	interval     time.Duration
}

// NewAutoSyncWorker crea el trabajador de sincronización automática.
func NewAutoSyncWorker(
	service *Service,
	source SnapshotSource,
	initialDelay time.Duration,
	interval time.Duration,
) *AutoSyncWorker {
	if initialDelay < 0 {
		initialDelay = 0
	}

	if interval <= 0 {
		interval = defaultAutoSyncInterval
	}

	return &AutoSyncWorker{
		service:      service,
		source:       source,
		initialDelay: initialDelay,
		interval:     interval,
	}
}

// Run mantiene la sincronización activa hasta cancelar el contexto.
func (worker *AutoSyncWorker) Run(
	ctx context.Context,
) {
	if worker == nil ||
		worker.service == nil ||
		worker.source == nil {
		log.Println(
			"La sincronización automática no pudo iniciarse: configuración incompleta.",
		)

		return
	}

	initialTimer := time.NewTimer(
		worker.initialDelay,
	)

	defer initialTimer.Stop()

	select {
	case <-ctx.Done():
		return

	case <-initialTimer.C:
		worker.synchronize(
			ctx,
		)
	}

	ticker := time.NewTicker(
		worker.interval,
	)

	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println(
				"Sincronización automática detenida.",
			)

			return

		case <-ticker.C:
			worker.synchronize(
				ctx,
			)
		}
	}
}

func (worker *AutoSyncWorker) synchronize(
	ctx context.Context,
) {
	startedAt := time.Now()

	result, err :=
		worker.service.Synchronize(
			ctx,
			worker.source,
		)

	if err != nil {
		if ctx.Err() != nil {
			return
		}

		log.Printf(
			"La sincronización automática falló: %v",
			err,
		)

		return
	}

	log.Printf(
		"Sincronización automática completada en %s: productos=%d, ventas=%d, movimientos=%d.",
		time.Since(
			startedAt,
		).Round(
			time.Millisecond,
		),
		result.Counts.Products,
		result.Counts.Sales,
		result.Counts.Movements,
	)
}
