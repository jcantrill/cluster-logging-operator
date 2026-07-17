package forwarder

import (
	"fmt"

	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/factory"
	"github.com/openshift/cluster-logging-operator/internal/generator/impl/otelc"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/conf"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	"github.com/openshift/cluster-logging-operator/internal/utils/toml"

	corev1 "k8s.io/api/core/v1"
)

type ConfigGenerator struct {
	collectorType string
}

func New(collectorType string) *ConfigGenerator {
	g := &ConfigGenerator{
		collectorType: collectorType,
	}
	return g
}

func (cg *ConfigGenerator) GenerateConf(secrets map[string]*corev1.Secret, clfspec obs.ClusterLogForwarderSpec, namespace, forwarderName string, resNames factory.ForwarderResourceNames, op utils.Options) (string, error) {
	var config interface{}
	switch cg.collectorType {
	case constants.ComponentNameOtelc:
		config = conf.Conf(secrets, clfspec, namespace, forwarderName, resNames, op)
	case constants.VectorName:
		config = otelc.NewConfig(secrets, clfspec, namespace, forwarderName, resNames, op)
	default:
		return "", fmt.Errorf("unknown collector type %s", cg.collectorType)
	}
	return toml.Marshal(config)
}
