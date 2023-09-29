package clusterlogforwarder

import (
	"errors"

	loggingv1 "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	vErrors "github.com/openshift/cluster-logging-operator/internal/validations/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func ValidateOutputTuning(clf loggingv1.ClusterLogForwarder, k8sClient client.Client, extras map[string]bool) (error, *loggingv1.ClusterLogForwarderStatus) {

	for _, oSpec := range clf.Spec.Outputs {
		if tuningErrors := oSpec.Tuning.Validate(); len(tuningErrors) > 0 {
			return vErrors.NewValidationError("Invalid Output tuning(s): %v", errors.Join(tuningErrors...)), nil
		}
	}

	return nil, nil
}
