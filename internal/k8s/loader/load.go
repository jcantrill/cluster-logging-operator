package loader

import (
	"context"
	logging "github.com/openshift/cluster-logging-operator/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/internal/migrations"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/internal/validations/clusterlogforwarder"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func FetchClusterLogging(k8sClient client.Client, namespace, name string, skipMigrations bool) (logging.ClusterLogging, error) {
	key := types.NamespacedName{Name: name, Namespace: namespace}
	proto := runtime.NewClusterLogging(namespace, name)
	if err := k8sClient.Get(context.TODO(), key, proto); err != nil {
		return logging.ClusterLogging{}, err
	}
	// Do not modify cached copy
	clusterLogging := *proto.DeepCopy()
	if skipMigrations {
		return clusterLogging, nil
	}
	// TODO Drop migration upon introduction of v2
	clusterLogging.Spec = migrations.MigrateCollectionSpec(clusterLogging.Spec)
	return clusterLogging, nil
}

// TODO Add in named conditions to return
func FetchClusterLogForwarder(k8sClient client.Client, namespace, name string, fetchClusterLogging func() logging.ClusterLogging) (logging.ClusterLogForwarder, error) {
	key := types.NamespacedName{Name: name, Namespace: namespace}
	proto := runtime.NewClusterLogForwarder(namespace, name)
	if err := k8sClient.Get(context.TODO(), key, proto); err != nil {
		// TODO Handle the case of CL without CLF and "generating" CLF
		//if !apierrors.IsNotFound(err) {
		//	log.Error(err, "Encountered unexpected error getting", "forwarder", nsname)
		//}
		//forwarder.Spec = logging.ClusterLogForwarderSpec{}
		return logging.ClusterLogForwarder{}, err
	}
	// Do not modify cached copy
	forwarder := *proto.DeepCopy()
	// TODO Drop migration upon introduction of v2
	extras := map[string]bool{}
	forwarder.Spec, extras = migrations.MigrateClusterLogForwarderSpec(forwarder.Spec, fetchClusterLogging().Spec.LogStore, extras)

	if err, _ := clusterlogforwarder.Validate(forwarder, k8sClient, extras); err != nil {
		return forwarder, err
	}

	return forwarder, nil
}
