package logforwarder

import (
	internalobs "github.com/openshift/cluster-logging-operator/internal/api/observability"
	"github.com/openshift/cluster-logging-operator/internal/collector"
	"github.com/openshift/cluster-logging-operator/internal/components/clusterlogdistributor"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/reconcile"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func ReconcileService(k8sClient client.Client, lf internalobs.LogForwarder) error {
	ports := []corev1.ServicePort{
		{
			Name:     "cld-source",
			Protocol: corev1.ProtocolTCP,
			Port:     clusterlogdistributor.DefaultLogForwarderPort,
			TargetPort: intstr.IntOrString{
				Type:   intstr.String,
				StrVal: "forward-logs",
			},
		},
		{
			Port:       collector.MetricsPort,
			TargetPort: intstr.FromString(collector.MetricsPortName),
			Name:       collector.MetricsPortName,
		},
	}
	svc := runtime.NewService(lf.Namespace(), lf.DeployedName(), lf.CommonLabelsInitializer())
	selector := runtime.Selectors(lf.Name(), internalobs.LogForwarderComponentName, clusterlogdistributor.ApplicationName)
	runtime.NewServiceBuilder(svc).WithSelector(selector).WithServicePort(ports)
	svc.Annotations = map[string]string{
		constants.AnnotationServingCertSecretName: lf.ResourceNames().SecretMetrics,
	}
	lf.AddOwnerRefTo(svc)
	return reconcile.Service(k8sClient, svc)
}
