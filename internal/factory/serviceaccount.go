package factory

import (
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	core "k8s.io/api/core/v1"
)

// NewServiceAccount stubs an instance of a ServiceAccount
func NewServiceAccount(namespace string, name string, applicationName, componentName, instanceName string) *core.ServiceAccount {
	sa := runtime.NewServiceAccount(namespace, name)
	runtime.SetCommonLabels(sa, applicationName, instanceName, componentName)
	return sa
}
