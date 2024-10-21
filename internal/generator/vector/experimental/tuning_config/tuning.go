//go:build experimental

package tuning_config

import (
	controllerobs "github.com/openshift/cluster-logging-operator/internal/controller/observability"
	forwardergenerator "github.com/openshift/cluster-logging-operator/internal/generator/forwarder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
)

const (
	enabledAnnotation = "logging.observability.openshift.io/experimental-forwarder-tuning"
)

func init() {
	forwardergenerator.AddConfigModifier(ModifyToml)
}

func ModifyToml(k8sClient client.Client, namespace, name string, annotations map[string]string, conf string) string {
	if value, found := annotations[enabledAnnotation]; found {
		if configMaps, err := controllerobs.FetchConfigMaps(k8sClient, namespace, value); err != nil && len(configMaps) > 0 {
			tunings := &ExperimentalCLFTuning{}
			if err = yaml.Unmarshal([]byte(configMaps[0].Data[name]), tunings); err != nil {
				toml := ParseToml(conf)
				return toml.Modify(*tunings).String()
			}
		}
	}
	return conf
}
