package observability

import (
	"fmt"

	obsv1 "github.com/openshift/cluster-logging-operator/api/observability/v1"
	obsv1beta1 "github.com/openshift/cluster-logging-operator/api/observability/v1beta1"
	"github.com/openshift/cluster-logging-operator/internal/factory"
	vectorapisources "github.com/openshift/cluster-logging-operator/internal/generator/vector/api"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	obsruntime "github.com/openshift/cluster-logging-operator/internal/runtime/observability"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	LogForwarderComponentName  = "logforwarder"
	vectorDefaultListenAddress = "0.0.0.0:6000"
	defaultInternalInputName   = "internal"
)

type LogForwarder struct {
	obsv1beta1.LogForwarder
}

func NewLogForwarder(lf obsv1beta1.LogForwarder) *LogForwarder {
	return &LogForwarder{lf}
}

func (lf *LogForwarder) DeployedName() string {
	return fmt.Sprintf("lf-%s", lf.LogForwarder.Name)
}

func (lf *LogForwarder) AddOwnerRefTo(object metav1.Object) {
	utils.AddOwnerRefToObject(object, utils.AsOwner(&lf.LogForwarder))
}

func (lf *LogForwarder) CommonLabelsInitializer() func(object runtime.Object) {
	return func(o runtime.Object) {
		runtime.SetCommonLabels(o, LogForwarderComponentName, lf.Name(), LogForwarderComponentName)
	}
}

func (lf *LogForwarder) AsClusterLogForwarder() obsv1.ClusterLogForwarder {

	return *obsruntime.NewClusterLogForwarder(lf.Namespace(), lf.DeployedName(), runtime.Initialize)
}

func (lf *LogForwarder) ResourceNames() factory.ForwarderResourceNames {
	return *factory.ResourceNames(lf.AsClusterLogForwarder())
}

func (lf *LogForwarder) Name() string {
	return lf.LogForwarder.Name
}
func (lf *LogForwarder) Namespace() string {
	return lf.LogForwarder.Namespace
}
func (lf *LogForwarder) Annotations() map[string]string {
	return lf.LogForwarder.Annotations
}

func (lf *LogForwarder) Inputs() (specs []Input) {
	return []Input{
		InputSpec{
			name:          defaultInternalInputName,
			componentType: vectorapisources.ComponentTypeVector,
			ReceiverSpec: &VectorReceiverSpec{
				ListenAddress: vectorDefaultListenAddress,
			},
		},
	}
}

// TODO: remove me
func (lf *LogForwarder) V1Inputs() (specs []obsv1.InputSpec) {
	return specs
}

func (lf *LogForwarder) Outputs() (specs []obsv1.OutputSpec) {
	return lf.LogForwarder.Spec.Outputs
}

func (lf *LogForwarder) Filters() (specs []obsv1.FilterSpec) {
	return specs
}

func (lf *LogForwarder) Pipelines() (specs []obsv1.PipelineSpec) {

	specs = append(specs, obsv1.PipelineSpec{
		Name:       defaultInternalInputName,
		InputRefs:  []string{defaultInternalInputName},
		OutputRefs: Outputs(lf.Spec.Outputs).Names(),
	})

	return specs
}

type Pipeline interface {
	InputRefs() []string
	OutputRefs() []string
	FilterRefs() []string
}

type PipelineSpec struct {
	V1PipelineSpec *obsv1.PipelineSpec
}
