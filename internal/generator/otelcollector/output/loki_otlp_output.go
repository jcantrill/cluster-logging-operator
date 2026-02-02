package output

import "github.com/openshift/cluster-logging-operator/internal/generator/otelcollector/api/exporters"

type LokiOtlp struct {
	exporters.OtlpHttp
}

func NewLokiOtlp(name, endpoint string) *LokiOtlp {
	return &LokiOtlp{
		OtlpHttp: *exporters.NewOtlpHttp(name, endpoint),
	}
}
