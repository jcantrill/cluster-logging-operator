package functional

import (
	log "github.com/ViaQ/logerr/v2/log/static"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"strings"
)

const (
	OTELReceiverConf = `
exporters:
  logging:
    verbosity: detailed
receivers:
  otlp:
    protocols:
      http:
        endpoint: localhost:8090
service:
  pipelines:
    logs:
      receivers: [otlp]
      exporters: [logging]
`
	OTELCollectorImage = "quay.io/openshift-logging/opentelemetry-collector:0.85.0-httpreqdump"
)

func (f *CollectorFunctionalFramework) AddOTELCollector(b *runtime.PodBuilder, outputName string) error {
	log.V(3).Info("Adding OTEL collector", "name", outputName)
	name := strings.ToLower(outputName)

	config := runtime.NewConfigMap(b.Pod.Namespace, name, map[string]string{
		"config.yaml": OTELReceiverConf,
	})
	log.V(2).Info("Creating configmap", "namespace", config.Namespace, "name", config.Name, "config.yaml", OTELReceiverConf)
	if err := f.Test.Client.Create(config); err != nil {
		return err
	}

	log.V(2).Info("Adding container", "name", name, "image", OTELCollectorImage)
	b.AddContainer(name, OTELCollectorImage).
		AddVolumeMount(config.Name, "/etc/functional", "", false).
		WithCmd([]string{"otelcol", "--config", "/etc/functional/config.yaml"}).
		End().
		AddConfigMapVolume(config.Name, config.Name)
	return nil
}
