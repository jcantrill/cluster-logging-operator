package lokistack

import (
	"fmt"

	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/api/observability"
	"github.com/openshift/cluster-logging-operator/internal/generator/common/lokistack"
	"github.com/openshift/cluster-logging-operator/internal/generator/helpers"
	"github.com/openshift/cluster-logging-operator/internal/generator/impl/otelc/api"
	"github.com/openshift/cluster-logging-operator/internal/generator/impl/otelc/api/exporters"
	"github.com/openshift/cluster-logging-operator/internal/generator/impl/otelc/api/types"
	"github.com/openshift/cluster-logging-operator/internal/generator/impl/otelc/helpers/tls"
	"github.com/openshift/cluster-logging-operator/internal/utils"
)

// New creates OTLPHTTP exporters for LokiStack using OpenShift Logging tenancy model.
// It generates one exporter per tenant (application, infrastructure, audit) based on the input types.
//
// Returns:
//   - exporterIDs: map of tenant -> exporter ID
//   - exporters: map of exporter ID -> configured OTLPHTTP exporter
func New(id string, o obs.OutputSpec, inputSpecs []obs.InputSpec, secrets observability.Secrets, op utils.Options) (exportersMap api.Exporters) {
	if o.LokiStack == nil {
		panic("LokiStack output spec is nil")
	}

	exportersMap = make(api.Exporters)

	// Determine tenants based on input types
	tenants := observability.DetermineTenants(inputSpecs)

	// Create an exporter for each tenant
	for _, tenant := range tenants {
		exporterID := helpers.MakeID(id, tenant)

		// Generate OTLPHTTP exporter for this tenant
		exporter := generateExporterForTenant(exporterID, o, tenant, secrets, op)
		exportersMap[exporter.ID()] = exporter
	}

	return exportersMap
}

// generateExporterForTenant creates an OTLPHTTP exporter configured for a specific tenant.
func generateExporterForTenant(exporterID string, o obs.OutputSpec, tenant string, secrets observability.Secrets, op utils.Options) *exporters.OtlpHttp {
	// Build the LokiStack OTLP endpoint URL for this tenant
	url := lokistack.OtlpURL(o.LokiStack, tenant)
	if url == "" {
		panic(fmt.Sprintf("LokiStack output has no valid URL for tenant %s", tenant))
	}

	// Create OTLPHTTP exporter
	exporter := exporters.NewOtlpHttp(exporterID, url)

	// Configure authentication
	if o.LokiStack.Authentication != nil && o.LokiStack.Authentication.Token != nil && o.LokiStack.Authentication.Token.Secret != nil {
		secretKey := fmt.Sprintf("%s.%s", o.LokiStack.Authentication.Token.Secret.Name, o.LokiStack.Authentication.Token.Secret.Key)
		tokenSecret := secrets[secretKey]
		if tokenSecret != nil && len(tokenSecret.Data) > 0 {
			// In OpenTelemetry Collector, bearer token is set via headers
			if exporter.Headers == nil {
				exporter.Headers = make(map[string]string)
			}
			tokenValue := string(tokenSecret.Data[o.LokiStack.Authentication.Token.Secret.Key])
			exporter.Headers["Authorization"] = fmt.Sprintf("Bearer %s", tokenValue)
		}
	}

	// Configure TLS if specified
	if o.TLS != nil {
		exporter.TLS = tls.NewTlsClientConfig(o.TLS, op)
	}

	// Apply tuning configuration
	if o.LokiStack.Tuning != nil {
		applyLokiTuning(exporter, o.LokiStack.Tuning)
	}

	return exporter
}

// applyLokiTuning applies LokiStack tuning configuration to the OTLPHTTP exporter.
func applyLokiTuning(exporter *exporters.OtlpHttp, tuning *obs.LokiTuningSpec) {
	if tuning == nil {
		return
	}

	// Set compression
	if tuning.Compression != "" {
		// OTLP supports "gzip" or "none"
		// LokiStack tuning uses "gzip", "snappy", or "none"
		// Map "snappy" to "gzip" since OTLP doesn't support snappy
		switch tuning.Compression {
		case "gzip":
			exporter.Compression = "gzip"
		case "snappy":
			exporter.Compression = "gzip" // Fallback to gzip
		case "none":
			exporter.Compression = "none"
		default:
			exporter.Compression = "gzip" // Default
		}
	}

	// Apply delivery mode (batch vs simple)
	if tuning.DeliveryMode == obs.DeliveryModeAtLeastOnce {
		// Enable retry and queue for at-least-once delivery
		exporter.RetryOnFailure = &types.RetrySettings{
			Enabled:         true,
			InitialInterval: "5s",
			MaxInterval:     "30s",
			MaxElapsedTime:  "5m",
		}

		exporter.SendingQueue = &types.QueueSettings{
			Enabled:      true,
			NumConsumers: 10,
			QueueSize:    1000,
		}
	}

	// Set max record size if specified
	if tuning.MaxWrite != nil {
		// OTLP doesn't have a direct equivalent for max record size
		// This would typically be handled by the receiver's max_log_size
	}
}
