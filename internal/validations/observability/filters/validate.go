package filters

import (
	internalcontext "github.com/openshift/cluster-logging-operator/internal/api/context"
	internalobs "github.com/openshift/cluster-logging-operator/internal/api/observability"
)

func Validate(context internalcontext.ForwarderContext) {
	filterMap := internalobs.Filters(context.Forwarder.Spec.Filters).Map()
	for _, filter := range filterMap {
		internalobs.SetCondition(&context.Forwarder.Status.FilterConditions, ValidateFilter(*filter))
	}
}
