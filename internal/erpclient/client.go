package erpclient

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"supermarket-sig-api/internal/erpdata"
)

const maxSnapshotBytes int64 = 20 * 1024 * 1024

// ErrNotConfigured indica que no existe una configuración
// suficiente para conectarse al ERP.
var ErrNotConfigured = errors.New(
	"el cliente ERP no está configurado",
)

// UpstreamError representa una respuesta HTTP no exitosa del ERP.
type UpstreamError struct {
	StatusCode int
	Message    string
}

func (err *UpstreamError) Error() string {
	if strings.TrimSpace(err.Message) == "" {
		return fmt.Sprintf(
			"el ERP respondió con estado HTTP %d",
			err.StatusCode,
		)
	}

	return fmt.Sprintf(
		"el ERP respondió con estado HTTP %d: %s",
		err.StatusCode,
		err.Message,
	)
}

// Client consulta el contrato Snapshot v1 producido por NestJS.
type Client struct {
	endpoint   string
	syncKey    string
	httpClient *http.Client
}

// NewClient crea el cliente REST hacia el ERP.
func NewClient(
	baseURL string,
	syncKey string,
	timeout time.Duration,
) *Client {
	baseURL = strings.TrimRight(
		strings.TrimSpace(baseURL),
		"/",
	)

	if timeout <= 0 {
		timeout = 20 * time.Second
	}

	endpoint := ""

	if baseURL != "" {
		endpoint =
			baseURL +
				"/integracion/sig/snapshot"
	}

	return &Client{
		endpoint: endpoint,

		syncKey: strings.TrimSpace(
			syncKey,
		),

		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// FetchSnapshot obtiene el snapshot empresarial desde NestJS.
func (client *Client) FetchSnapshot(
	ctx context.Context,
) (
	erpdata.Snapshot,
	error,
) {
	if client == nil ||
		client.endpoint == "" ||
		client.syncKey == "" {
		return erpdata.Snapshot{},
			ErrNotConfigured
	}

	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		client.endpoint,
		nil,
	)

	if err != nil {
		return erpdata.Snapshot{},
			fmt.Errorf(
				"no se pudo construir la solicitud al ERP: %w",
				err,
			)
	}

	request.Header.Set(
		"Accept",
		"application/json",
	)

	request.Header.Set(
		"X-SIG-Key",
		client.syncKey,
	)

	request.Header.Set(
		"User-Agent",
		"supermarket-sig-api/1.0",
	)

	response, err :=
		client.httpClient.Do(
			request,
		)

	if err != nil {
		return erpdata.Snapshot{},
			fmt.Errorf(
				"no se pudo conectar con el ERP: %w",
				err,
			)
	}

	defer response.Body.Close()

	if response.StatusCode !=
		http.StatusOK {
		body, _ := io.ReadAll(
			io.LimitReader(
				response.Body,
				4096,
			),
		)

		return erpdata.Snapshot{},
			&UpstreamError{
				StatusCode: response.StatusCode,

				Message: strings.TrimSpace(
					string(body),
				),
			}
	}

	decoder := json.NewDecoder(
		io.LimitReader(
			response.Body,
			maxSnapshotBytes,
		),
	)

	decoder.DisallowUnknownFields()

	var snapshot erpdata.Snapshot

	if err := decoder.Decode(
		&snapshot,
	); err != nil {
		return erpdata.Snapshot{},
			fmt.Errorf(
				"el ERP devolvió un snapshot JSON inválido: %w",
				err,
			)
	}

	if err := decoder.Decode(
		&struct{}{},
	); err != io.EOF {
		return erpdata.Snapshot{},
			errors.New(
				"el ERP devolvió más de un objeto JSON",
			)
	}

	return snapshot, nil
}
