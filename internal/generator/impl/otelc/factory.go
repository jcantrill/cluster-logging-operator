package otelc

import (
	"sort"

	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	"github.com/openshift/cluster-logging-operator/internal/factory"
	"github.com/openshift/cluster-logging-operator/internal/generator/impl/otelc/api"
	"github.com/openshift/cluster-logging-operator/internal/generator/impl/otelc/input"
	"github.com/openshift/cluster-logging-operator/internal/generator/impl/otelc/output"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/adapters" //TODO move out of vector
	"github.com/openshift/cluster-logging-operator/internal/utils"
	corev1 "k8s.io/api/core/v1"
)

func Conf(secrets map[string]*corev1.Secret, clfspec obs.ClusterLogForwarderSpec, namespace, forwarderName string, resNames factory.ForwarderResourceNames, op utils.Options) (config *api.Config) {

	config = api.NewConfig()

	inputsToReceivers := map[string]api.Receivers{}
	for _, i := range clfspec.Inputs {
		a := adapters.NewInput(i)
		receivers := input.New(a, resNames, secrets, op)
		inputsToReceivers[i.Name] = receivers
		config.AddReceivers(receivers)
	}

	outputsToExporters := map[string]api.Exporters{}
	for _, spec := range clfspec.Outputs {
		o := adapters.NewOutput(spec)
		// TODO fix inputs to the exporters
		exporters := output.New(o, clfspec.Inputs, secrets, op)
		outputsToExporters[spec.Name] = exporters
		config.AddExporters(exporters)
	}

	for _, p := range clfspec.Pipelines {

		pipeline := api.NewLogPipeline(p.Name)
		for _, inputRefs := range p.InputRefs {
			receivers := inputsToReceivers[inputRefs]
			for _, rec := range receivers {
				pipeline.Receivers = append(pipeline.Receivers, rec.ID())
			}
		}
		for _, outputRefs := range p.OutputRefs {
			exporters := outputsToExporters[outputRefs]
			for _, ex := range exporters {
				pipeline.Exporters = append(pipeline.Exporters, ex.ID())
			}
		}
		sort.Strings(pipeline.Receivers)
		sort.Strings(pipeline.Exporters)
		config.Service.AddPipeline(pipeline)
	}

	return config
}
