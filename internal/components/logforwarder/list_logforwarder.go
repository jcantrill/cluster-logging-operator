package logforwarder

import (
	"context"

	obsv1beta1 "github.com/openshift/cluster-logging-operator/api/observability/v1beta1"
	internalobs "github.com/openshift/cluster-logging-operator/internal/api/observability"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func List(k8sClient client.Client, labels client.MatchingLabels) (res []internalobs.LogForwarder, err error) {
	list := &obsv1beta1.LogForwarderList{}
	if err = k8sClient.List(context.TODO(), list, labels); err != nil {
		return nil, err
	}
	for _, l := range list.Items {
		res = append(res, internalobs.LogForwarder{LogForwarder: l})
	}
	return res, nil
}
