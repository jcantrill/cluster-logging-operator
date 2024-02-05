package v2beta

import (
	"github.com/openshift/cluster-logging-operator/apis/logging/v2beta"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
)

// NewClusterLogForwarder returns a ClusterLogForwarder with default name and deployment namespace.
func NewClusterLogForwarder(namespace, name string, visitors ...func(*v2beta.ClusterLogForwarder)) *v2beta.ClusterLogForwarder {
	clf := &v2beta.ClusterLogForwarder{}
	runtime.Initialize(clf, "", name)
	clf.Spec.Namespace = namespace
	for _, v := range visitors {
		v(clf)
	}
	return clf
}
