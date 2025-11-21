package logforwarder

import (
	log "github.com/ViaQ/logerr/v2/log/static"
	internalobs "github.com/openshift/cluster-logging-operator/internal/api/observability"
	"github.com/openshift/cluster-logging-operator/internal/collector/vector"
	"github.com/openshift/cluster-logging-operator/internal/reconcile"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/internal/utils/comparators"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ReconcileConfig reconciles a collector config specifically for the collector defined by the factory
func ReconcileConfig(k8sClient client.Client, lf internalobs.LogForwarder, config, configName string) error {
	log.V(3).Info("Updating ConfigMap")
	configMap := runtime.NewConfigMap(
		lf.Namespace(),
		configName,
		map[string]string{
			vector.ConfigFile: config,
		},
		lf.CommonLabelsInitializer())
	lf.AddOwnerRefTo(configMap)
	return reconcile.Configmap(k8sClient, k8sClient, configMap, comparators.CompareLabels)
}
