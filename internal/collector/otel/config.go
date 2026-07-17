package otel

import (
	log "github.com/ViaQ/logerr/v2/log/static"
	"github.com/openshift/cluster-logging-operator/internal/factory"
	"github.com/openshift/cluster-logging-operator/internal/reconcile"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	"github.com/openshift/cluster-logging-operator/internal/utils/comparators"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ReconcileCollectorConfig reconciles a collector config specifically for the collector defined by the factory
func ReconcileCollectorConfig(k8sClient client.Client, reader client.Reader, resourceNames factory.ForwarderResourceNames, namespace, collectorConfig string, owner metav1.OwnerReference, visitors ...func(o runtime.Object)) error {
	log.V(3).Info("Updating ConfigMap and Secrets")
	configMap := runtime.NewConfigMap(
		namespace,
		resourceNames.ConfigMap,
		map[string]string{
			ConfigFile: collectorConfig,
		},
		visitors...)

	utils.AddOwnerRefToObject(configMap, owner)
	return reconcile.Configmap(k8sClient, reader, configMap, comparators.CompareLabels)
}
