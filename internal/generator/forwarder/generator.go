package forwarder

import (
	internalobs "github.com/openshift/cluster-logging-operator/internal/api/observability"
	"github.com/openshift/cluster-logging-operator/internal/factory"
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/conf"

	"github.com/openshift/cluster-logging-operator/internal/generator/helpers"
	corev1 "k8s.io/api/core/v1"
)

type ConfigGenerator struct {
	g      framework.Generator
	conf   func(secrets map[string]*corev1.Secret, clfspec internalobs.Forwarder, namespace, forwarderName string, resNames factory.ForwarderResourceNames, op framework.Options) []framework.Section
	format func(conf string) string
}

func New() *ConfigGenerator {
	g := &ConfigGenerator{
		format: helpers.FormatVectorToml,
		conf:   conf.Conf,
	}
	return g
}

func (cg *ConfigGenerator) GenerateConf(secrets map[string]*corev1.Secret, clfspec internalobs.Forwarder, namespace, forwarderName string, resNames factory.ForwarderResourceNames, op framework.Options) (string, error) {
	sections := cg.conf(secrets, clfspec, namespace, forwarderName, resNames, op)
	conf, err := cg.g.GenerateConf(framework.MergeSections(sections)...)
	return cg.format(conf), err
}
