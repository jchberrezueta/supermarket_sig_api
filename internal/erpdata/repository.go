package erpdata

import "context"

// Repository define el almacenamiento de datos sincronizados.
type Repository interface {
	ReplaceSnapshot(
		ctx context.Context,
		snapshot Snapshot,
		result ImportResult,
	) error

	CurrentSnapshot(
		ctx context.Context,
	) (
		Snapshot,
		bool,
		error,
	)

	State(
		ctx context.Context,
	) (
		IntegrationState,
		error,
	)
}
