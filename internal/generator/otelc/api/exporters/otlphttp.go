package exporters

import (
	"github.com/openshift/cluster-logging-operator/internal/generator/otelc/api/types"
)

// OTLPHTTP represents the OpenTelemetry Collector OTLP HTTP exporter configuration
// This is the recommended way to send logs to Loki v3+ using native OTLP ingestion
// See: https://grafana.com/docs/loki/latest/send-data/otel/
// See: https://github.com/open-telemetry/opentelemetry-collector/tree/main/exporter/otlphttpexporter
type OTLPHTTP struct {
	// exporterType is used internally to identify the exporter type
	// It is not serialized to YAML
	exporterType types.ExporterType

	// Endpoint is the target URL for the OTLP HTTP exporter
	// For Loki v3+, use: http://loki:3100/otlp
	Endpoint string `yaml:"endpoint"`

	// TLS configuration
	TLS *types.TLSClientConfig `yaml:"tls,omitempty"`

	// Timeout for HTTP requests (default: "30s")
	Timeout string `yaml:"timeout,omitempty"`

	// Headers to include in HTTP requests
	Headers map[string]string `yaml:"headers,omitempty"`

	// Compression for HTTP requests (default: "gzip")
	// Options: "gzip", "none"
	Compression string `yaml:"compression,omitempty"`

	// Encoding specifies the encoding format for OTLP (default: "proto")
	// Options: "proto", "json"
	Encoding string `yaml:"encoding,omitempty"`

	// ReadBufferSize for HTTP client (bytes)
	ReadBufferSize int `yaml:"read_buffer_size,omitempty"`

	// WriteBufferSize for HTTP client (bytes)
	WriteBufferSize int `yaml:"write_buffer_size,omitempty"`

	// MaxIdleConns for HTTP client connection pool
	MaxIdleConns int `yaml:"max_idle_conns,omitempty"`

	// MaxIdleConnsPerHost for HTTP client connection pool
	MaxIdleConnsPerHost int `yaml:"max_idle_conns_per_host,omitempty"`

	// MaxConnsPerHost limits the total number of connections per host
	MaxConnsPerHost int `yaml:"max_conns_per_host,omitempty"`

	// IdleConnTimeout is the maximum amount of time an idle connection will remain idle
	IdleConnTimeout string `yaml:"idle_conn_timeout,omitempty"`

	// Queue configuration for sending data
	SendingQueue *types.QueueSettings `yaml:"sending_queue,omitempty"`

	// Retry configuration for failed requests
	RetryOnFailure *types.RetrySettings `yaml:"retry_on_failure,omitempty"`
}

func (e *OTLPHTTP) ExporterType() types.ExporterType {
	return types.ExporterTypeOTLPHTTP
}

// NewOTLPHTTP creates a new OTLPHTTP exporter with the given endpoint
// For Loki v3+, use the OTLP endpoint: http://loki:3100/otlp
func NewOTLPHTTP(endpoint string) *OTLPHTTP {
	return &OTLPHTTP{
		exporterType: types.ExporterTypeOTLPHTTP,
		Endpoint:     endpoint,
	}
}
