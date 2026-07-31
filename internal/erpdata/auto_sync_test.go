package erpdata

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

type autoSyncTestSource struct {
	mu sync.Mutex

	calls    int
	failures int
	called   chan struct{}
}

func (
	source *autoSyncTestSource,
) FetchSnapshot(
	ctx context.Context,
) (
	Snapshot,
	error,
) {
	if err := ctx.Err(); err != nil {
		return Snapshot{},
			err
	}

	source.mu.Lock()

	source.calls++

	callNumber :=
		source.calls

	source.mu.Unlock()

	select {
	case source.called <- struct{}{}:

	default:
	}

	if callNumber <=
		source.failures {
		return Snapshot{},
			errors.New(
				"error temporal del ERP",
			)
	}

	return Snapshot{
			ContractVersion: SnapshotContractVersion,

			Source: SnapshotSourceERP,

			GeneratedAt: time.Now(),
		},
		nil
}

func (
	source *autoSyncTestSource,
) totalCalls() int {
	source.mu.Lock()
	defer source.mu.Unlock()

	return source.calls
}

func TestAutoSyncWorkerRetriesAndStops(
	t *testing.T,
) {
	repository :=
		NewMemoryRepository()

	service :=
		NewService(
			repository,
		)

	source :=
		&autoSyncTestSource{
			failures: 1,

			called: make(
				chan struct{},
				8,
			),
		}

	worker :=
		NewAutoSyncWorker(
			service,
			source,
			5*time.Millisecond,
			15*time.Millisecond,
		)

	ctx, cancel :=
		context.WithCancel(
			context.Background(),
		)

	done := make(
		chan struct{},
	)

	go func() {
		defer close(
			done,
		)

		worker.Run(
			ctx,
		)
	}()

	waitForAutoSyncCall(
		t,
		source.called,
	)

	waitForAutoSyncCall(
		t,
		source.called,
	)

	waitForAutoSyncData(
		t,
		service,
	)

	cancel()

	select {
	case <-done:

	case <-time.After(
		time.Second,
	):
		t.Fatal(
			"el trabajador no se detuvo después de cancelar el contexto",
		)
	}

	callsAfterStop :=
		source.totalCalls()

	time.Sleep(
		50 * time.Millisecond,
	)

	finalCalls :=
		source.totalCalls()

	if finalCalls != callsAfterStop {
		t.Fatalf(
			"el trabajador continuó ejecutándose después de detenerse: antes=%d, después=%d",
			callsAfterStop,
			finalCalls,
		)
	}

	if finalCalls < 2 {
		t.Fatalf(
			"se esperaban al menos dos intentos de sincronización; se obtuvieron %d",
			finalCalls,
		)
	}
}

func waitForAutoSyncCall(
	t *testing.T,
	called <-chan struct{},
) {
	t.Helper()

	select {
	case <-called:

	case <-time.After(
		time.Second,
	):
		t.Fatal(
			"no se ejecutó la sincronización automática esperada",
		)
	}
}

func waitForAutoSyncData(
	t *testing.T,
	service *Service,
) {
	t.Helper()

	deadline :=
		time.Now().Add(
			time.Second,
		)

	for time.Now().Before(
		deadline,
	) {
		state, err :=
			service.State(
				context.Background(),
			)

		if err != nil {
			t.Fatalf(
				"no se pudo consultar el estado: %v",
				err,
			)
		}

		if state.HasData {
			return
		}

		time.Sleep(
			5 * time.Millisecond,
		)
	}

	t.Fatal(
		"la sincronización exitosa no almacenó el snapshot",
	)
}
