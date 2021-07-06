package pipeline

import (
	"github.com/openshift/cluster-logging-operator/internal/builders/config/fluentd"
	clffluentd "github.com/openshift/cluster-logging-operator/internal/builders/logforwarder/fluentd"
	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
)

type PipelineToOutputs struct {
	pipeline logging.PipelineSpec
}

func NewPipelineToOutputsBuilder(pipeline logging.PipelineSpec) *PipelineToOutputs{
	return &PipelineToOutputs{
		pipeline: pipeline,
	}
}
func(b *PipelineToOutputs) AsString() []string {
	if b.pipeline.OutputRefs == 1 {
		fluentd.Match()
	}else {

	}
	return fluentd.Label(clffluentd.FormatLabelName(b.pipeline.Name), b.outForward.AsList())
}

func(b *PipelineToOutputs) String() string {
	return ""
}