package logforwarder

import (
	internalobs "github.com/openshift/cluster-logging-operator/internal/api/observability"
	"github.com/openshift/cluster-logging-operator/internal/factory"
	"github.com/openshift/cluster-logging-operator/internal/reconcile"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func ReconcileService(k8sClient client.Client, lf internalobs.LogForwarder) error {
	ports := []corev1.ServicePort{
		{
			Name:     "cld-source",
			Protocol: corev1.ProtocolTCP,
			Port:     443,
			TargetPort: intstr.IntOrString{
				Type:   intstr.String,
				StrVal: "forward-logs",
			},
		},
	}
	svc := factory.NewService(lf.DeployedName(), lf.Namespace(), internalobs.LogForwarderComponentName, lf.Name(), ports, lf.CommonLabelsInitializer())
	lf.AddOwnerRefTo(svc)
	return reconcile.Service(k8sClient, svc)
}
