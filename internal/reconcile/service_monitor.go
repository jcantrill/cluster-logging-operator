package reconcile

import (
	"context"

	log "github.com/ViaQ/logerr/v2/log/static"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ServiceMonitor ensures a ServiceMonitor exists with the desired initial configuration.
// After creation, the operator does not update the ServiceMonitor, allowing users to manually
// edit it without their changes being reverted.
func ServiceMonitor(k8Client client.Client, desired *monitoringv1.ServiceMonitor) error {
	err := k8Client.Create(context.TODO(), desired)
	if err != nil {
		if errors.IsAlreadyExists(err) {
			// ServiceMonitor already exists, don't update it - allow manual edits to persist
			log.V(3).Info("serviceMonitor already exists, skipping", "name", desired.Name, "namespace", desired.Namespace)
			return nil
		}
		return err
	}
	log.V(3).Info("created serviceMonitor", "name", desired.Name, "namespace", desired.Namespace)
	return nil
}
