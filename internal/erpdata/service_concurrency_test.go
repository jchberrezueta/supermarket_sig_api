package erpdata

import (
	"context"
	"sync"
	"testing"
	"time"
)

type concurrencyTrackingRepository struct {
	mu sync.Mutex

	activeWrites    int
	maxActiveWrites int
	writeDelay      time.Duration
}

func (
	repository *concurrencyTrackingRepository,
) ReplaceSnapshot(
	_ context.Context,
	_ Snapshot,
	_ ImportResult,
) error {
	repository.mu.Lock()

	repository.activeWrites++

	if repository.activeWrites >
		repository.maxActiveWrites {
		repository.maxActiveWrites =
			repository.activeWrites
	}

	repository.mu.Unlock()

	time.Sleep(
		repository.writeDelay,
	)

	repository.mu.Lock()
	repository.activeWrites--
	repository.mu.Unlock()

	return nil
}

func (
	repository *concurrencyTrackingRepository,
) CurrentSnapshot(
	_ context.Context,
) (
	Snapshot,
	bool,
	error,
) {
	return Snapshot{},
		false,
		nil
}

func (
	repository *concurrencyTrackingRepository,
) State(
	_ context.Context,
) (
	IntegrationState,
	error,
) {
	return IntegrationState{},
		nil
}

func TestServiceSerializesSnapshotImports(
	t *testing.T,
) {
	repository :=
		&concurrencyTrackingRepository{
			writeDelay: 50 * time.Millisecond,
		}

	service := NewService(
		repository,
	)

	snapshot := Snapshot{
		ContractVersion: SnapshotContractVersion,
		Source:          SnapshotSourceERP,
		GeneratedAt:     time.Now(),
	}

	const workers = 4

	start := make(
		chan struct{},
	)

	errorsChannel := make(
		chan error,
		workers,
	)

	var waitGroup sync.WaitGroup

	for range workers {
		waitGroup.Add(1)

		go func() {
			defer waitGroup.Done()

			<-start

			_, err :=
				service.ImportSnapshot(
					context.Background(),
					snapshot,
				)

			errorsChannel <- err
		}()
	}

	close(start)

	waitGroup.Wait()
	close(errorsChannel)

	for err := range errorsChannel {
		if err != nil {
			t.Fatalf(
				"la importación concurrente falló: %v",
				err,
			)
		}
	}

	repository.mu.Lock()

	maxActiveWrites :=
		repository.maxActiveWrites

	repository.mu.Unlock()

	if maxActiveWrites != 1 {
		t.Fatalf(
			"se esperaba una sola escritura simultánea; se obtuvieron %d",
			maxActiveWrites,
		)
	}
}
