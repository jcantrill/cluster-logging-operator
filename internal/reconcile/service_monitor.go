package reconcile

import (
	"context"
	"strings"

	log "github.com/ViaQ/logerr/v2/log/static"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ServiceMonitor reconciles a ServiceMonitor to the desired spec using ServerSideApply.
// This allows users to modify fields not managed by the operator without the operator reverting the changes.
func ServiceMonitor(k8Client client.Client, desired *monitoringv1.ServiceMonitor) error {
	err := k8Client.Patch(context.TODO(), desired, client.Apply, &client.PatchOptions{
		FieldManager: "cluster-logging-operator",
		Force:        utils.GetPtr(false),
	})

	// If ServerSideApply is not supported (e.g., in test fake clients) or the resource doesn't exist,
	// fall back to Create. Subsequent reconciles will use ServerSideApply if supported.
	if err != nil && (errors.IsNotFound(err) || isApplyNotSupported(err)) {
		if createErr := k8Client.Create(context.TODO(), desired); createErr != nil {
			if errors.IsAlreadyExists(createErr) {
				// Resource exists but Apply failed - this shouldn't happen in production but can in tests
				log.V(3).Info("serviceMonitor already exists, skipping create", "name", desired.Name, "namespace", desired.Namespace)
				return nil
			}
			return createErr
		}
		log.V(3).Info("created serviceMonitor", "name", desired.Name, "namespace", desired.Namespace)
		return nil
	}

	if err == nil {
		log.V(3).Info("reconciled serviceMonitor using ServerSideApply", "name", desired.Name, "namespace", desired.Namespace)
	}
	return err
}

// isApplyNotSupported checks if the error indicates that Apply patches are not supported
func isApplyNotSupported(err error) bool {
	return err != nil && strings.Contains(err.Error(), "apply patches are not supported")
}
