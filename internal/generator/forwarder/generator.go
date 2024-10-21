package forwarder

import (
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/factory"
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/conf"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/openshift/cluster-logging-operator/internal/generator/helpers"
	corev1 "k8s.io/api/core/v1"
)

var (
	configModRegistry []func(client.Client, string, string, map[string]string, string) string
)

func AddConfigModifier(m func(client.Client, string, string, map[string]string, string) string) {
	configModRegistry = append(configModRegistry, m)
}

type ConfigGenerator struct {
	g           framework.Generator
	conf        func(secrets map[string]*corev1.Secret, clfspec obs.ClusterLogForwarderSpec, namespace, forwarderName string, resNames factory.ForwarderResourceNames, op framework.Options) []framework.Section
	format      func(conf string) string
	k8sClient   client.Client
	annotations map[string]string
}

func New(k8sClient client.Client, annotations map[string]string) *ConfigGenerator {
	g := &ConfigGenerator{
		format:      helpers.FormatVectorToml,
		conf:        conf.Conf,
		k8sClient:   k8sClient,
		annotations: annotations,
	}
	return g
}

func (cg *ConfigGenerator) GenerateConf(secrets map[string]*corev1.Secret, clfspec obs.ClusterLogForwarderSpec, namespace, forwarderName string, resNames factory.ForwarderResourceNames, op framework.Options) (string, error) {
	sections := cg.conf(secrets, clfspec, namespace, forwarderName, resNames, op)
	conf, err := cg.g.GenerateConf(framework.MergeSections(sections)...)
	for _, modifier := range configModRegistry {
		conf = modifier(cg.k8sClient, namespace, forwarderName, cg.annotations, conf)
	}
	return cg.format(conf), err
}
