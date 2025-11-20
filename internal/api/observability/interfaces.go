package observability

import (
	obsv1 "github.com/openshift/cluster-logging-operator/api/observability/v1"
	obsv1beta1 "github.com/openshift/cluster-logging-operator/api/observability/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Forwarder is the logging workload which collects logs based upon defined inputs
// and wires them to outputs through filters using pipelines
type Forwarder interface {
	Name() string
	Namespace() string
	Inputs() (specs []Input)

	// TODO: remove me
	V1Inputs() (specs []obsv1.InputSpec)

	// TODO: refactor to return an interface
	Outputs() (specs []obsv1.OutputSpec)

	// TODO: refactor to return an interface
	Filters() (specs []obsv1.FilterSpec)

	// TODO: refactor to return an interface
	Pipelines() (specs []obsv1.PipelineSpec)

	Annotations() map[string]string
}

//type Output interface {
//	ComponentType() string
//	AsV1OutputSpec() obsv1.OutputSpec
//}
//
//type OutputSpec struct {
//	V1OutputSpec *obsv1.OutputSpec
//	obsv1beta1.
//}
//
//func (o *OutputSpec) AsV1OutputSpec() obsv1.OutputSpec {
//	if o.V1OutputSpec == nil {
//		return *o.V1OutputSpec
//	}
//	return obsv1.OutputSpec{}
//}
//func (o *OutputSpec) ComponentType() string {
//	if o.V1OutputSpec == nil {
//		return *o.V1OutputSpec.Type.String()
//	}
//	return obsv1.OutputSpec{}
//}

// Input
// TODO: Combine with the vector adapters?
type Input interface {
	LogType() string
	ComponentType() string
	Name() string
	PodSelector() *metav1.LabelSelector
	RateLimitPerContainer() *obsv1.LimitSpec

	//TODO: Delete me
	AsV1InputSpec() obsv1.InputSpec
	AsInputSpec() InputSpec
}

type ReceiverSpec interface {
	Address() string
}

//TODO: Move me

func ConvertInputsToV1InputSpecs(inputs []Input) (specs []obsv1.InputSpec) {
	for _, i := range inputs {
		specs = append(specs, i.AsV1InputSpec())
	}
	return specs
}

type VectorReceiverSpec struct {
	// ListenAddress is a host:port (i.e. 0.0.0.0:4000)
	ListenAddress string
}

func (v *VectorReceiverSpec) Address() string {
	return v.ListenAddress
}

type InputSpec struct {
	name                         string
	componentType                string
	V1InputSpec                  *obsv1.InputSpec
	V1Beta1LogForwarderInputSpec *obsv1beta1.LogForwarderInputSpec
	ReceiverSpec                 ReceiverSpec
}

func (i InputSpec) LogType() string {
	if i.V1InputSpec != nil {
		return i.V1InputSpec.Type.String()
	}
	return ""
}

func (i InputSpec) AsV1InputSpec() obsv1.InputSpec {
	if i.V1InputSpec != nil {
		return *i.V1InputSpec
	}
	return obsv1.InputSpec{}
}
func (i InputSpec) AsInputSpec() InputSpec {
	return i
}

func (i InputSpec) ComponentType() string {
	return i.componentType
}

func (i InputSpec) Name() string {
	if i.V1InputSpec != nil {
		return i.V1InputSpec.Name
	}
	return i.name
}

func (i InputSpec) PodSelector() *metav1.LabelSelector {
	switch {
	case i.V1InputSpec != nil && i.V1InputSpec.Type == obsv1.InputTypeApplication:
		return i.V1InputSpec.Application.Selector
	default:
		return nil
	}
}

func (i InputSpec) RateLimitPerContainer() *obsv1.LimitSpec {
	switch {
	case i.V1InputSpec != nil && i.V1InputSpec.Type == obsv1.InputTypeApplication && i.V1InputSpec.Application.Tuning != nil:
		return i.V1InputSpec.Application.Tuning.RateLimitPerContainer
	default:
		return nil
	}
}
