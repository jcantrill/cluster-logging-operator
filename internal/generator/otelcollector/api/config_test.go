package api_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/openshift/cluster-logging-operator/internal/generator/otelcollector/api"
	"github.com/openshift/cluster-logging-operator/internal/generator/otelcollector/api/extensions"
	"github.com/openshift/cluster-logging-operator/internal/generator/otelcollector/output"
	"github.com/openshift/cluster-logging-operator/internal/generator/otelcollector/source"
	"github.com/openshift/cluster-logging-operator/test"
)

var _ = Describe("Config", func() {

	It("should work", func() {
		containerLogs := source.NewKubernetesLogs("", []string{"/var/log/pods/*/*/*.log"}, []string{}, 1)
		lokiReceiver := output.NewLokiOtlp("loki", "http://localhost/otlp")
		bearTokenAuth := extensions.NewBearerTokenAuth("", "/etc/secret/token")
		config := api.NewConfig().
			SetReceiver(containerLogs).
			SetExporter(lokiReceiver).
			SetExtension(bearTokenAuth).
			SetPipeline(api.NewLogsPipeline("").
				AddExporter(containerLogs).
				AddReciever(lokiReceiver))
		Expect(test.YAMLString(config)).To(MatchYAML(""))

	})
})
