package output

import (
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/generator/impl/otelc/api"
	"github.com/openshift/cluster-logging-operator/internal/generator/impl/otelc/output/lokistack"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/adapters"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	corev1 "k8s.io/api/core/v1"
)

func New(o *adapters.Output, inputs []obs.InputSpec, secrets map[string]*corev1.Secret, op utils.Options) (exporters api.Exporters) {

	exporters = api.Exporters{}

	switch o.Type {
	case obs.OutputTypeLokiStack:
		exporters.Merge(lokistack.New(o.Name, o.OutputSpec, inputs, secrets))
	}
	return exporters
}
