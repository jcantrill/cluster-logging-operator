package v1

import (
	"fmt"
	"github.com/openshift/cluster-logging-operator/apis/logging/v2beta1"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/filter/openshift"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

// ConvertTo converts this CLF to the Hub version (v2beta1).
func (from *ClusterLogForwarder) ConvertTo(dstRaw conversion.Hub) error {
	to := dstRaw.(*v2beta1.ClusterLogForwarder)
	to.Annotations = from.Annotations
	to.Labels = from.Labels
	to.Spec.ServiceAccountName = from.Spec.ServiceAccountName
	for _, i := range from.Spec.Inputs {
		to.Spec.Inputs = append(to.Spec.Inputs, migrateInput(i))
	}
	for _, o := range from.Spec.Outputs {
		to.Spec.Outputs = append(to.Spec.Outputs, migrateOutput(o))
	}
	for _, f := range from.Spec.Filters {
		to.Spec.Filters = append(to.Spec.Filters, migrateFilter(f))
	}
	for i, p := range from.Spec.Pipelines {
		toPipeline, filters := migratePipeline(i, p)
		to.Spec.Filters = append(to.Spec.Filters, filters...)
		to.Spec.Pipelines = append(to.Spec.Pipelines, toPipeline)
	}
	return nil
}

func migratePipeline(i int, from PipelineSpec) (to v2beta1.PipelineSpec, filters []v2beta1.FilterSpec) {
	to = v2beta1.PipelineSpec{
		Name:       from.Name,
		InputRefs:  from.InputRefs,
		OutputRefs: from.OutputRefs,
		FilterRefs: from.FilterRefs,
	}
	if to.Name == "" {
		to.Name = fmt.Sprintf("pipeline_%d", i)
	}
	if len(from.Labels) > 0 {
		//&v2beta1.FilterSpec{Type: openshiftfilter.Labels}, Labels: from.Labels }
		//add filterref
	}
	if from.Parse == openshift.ParseTypeJson {
		to.FilterRefs = append(to.FilterRefs, openshift.ParseJson)
	}
	if from.DetectMultilineErrors {
		to.FilterRefs = append(to.FilterRefs, openshift.DetectMultilineException)
	}

	return to, filters
}

func migrateFilter(f FilterSpec) v2beta1.FilterSpec {
	return v2beta1.FilterSpec{}
}

func migrateOutput(o OutputSpec) v2beta1.OutputSpec {
	return v2beta1.OutputSpec{}
}

func migrateInput(from InputSpec) v2beta1.InputSpec {
	to := v2beta1.InputSpec{
		Name: from.Name,
	}
	switch {
	case from.Receiver != nil:
		to.Receiver = &v2beta1.ReceiverSpec{
			Type: from.Receiver.Type,
		}
		if from.Receiver.HTTP != nil {
			to.Receiver.HTTP = &v2beta1.HTTPReceiver{
				Port:   from.Receiver.HTTP.Port,
				Format: from.Receiver.HTTP.Format,
			}
		}
		if from.Receiver.Syslog != nil {
			to.Receiver.Syslog = &v2beta1.SyslogReceiver{
				Port: from.Receiver.Syslog.Port,
			}
		}
	case from.Infrastructure != nil:
		to.Infrastructure = &v2beta1.Infrastructure{
			Sources: from.Infrastructure.Sources,
		}
	case from.Audit != nil:
		to.Audit = &v2beta1.Audit{
			Sources: from.Audit.Sources,
		}
	case from.Application != nil:
		to.Application = &v2beta1.Application{
			Namespaces: &v2beta1.InclusionSpec{
				Include: from.Application.Namespaces,
				Exclude: from.Application.ExcludeNamespaces,
			},
		}
		if from.Application.Containers != nil {
			to.Application.Containers = &v2beta1.InclusionSpec{
				Include: from.Application.Containers.Include,
				Exclude: from.Application.Containers.Exclude,
			}
		}
	}
	return to
}

// ConvertFrom converts from the Hub version (v2beta1) to this version.
func (dst *ClusterLogForwarder) ConvertFrom(srcRaw conversion.Hub) error {
	//src := srcRaw.(*v2beta1.ClusterLogForwarder)
	return nil
}
