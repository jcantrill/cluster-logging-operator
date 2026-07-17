package daemonset

import (
	"context"

	appv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func Get(k8sClient client.Client, namespace, name string) (proto *appv1.DaemonSet, err error) {
	key := types.NamespacedName{Name: name, Namespace: namespace}
	proto = &appv1.DaemonSet{}
	if err := k8sClient.Get(context.TODO(), key, proto); err != nil {
		if errors.IsNotFound(err) {
			return nil, err
		}
		return proto, err
	}

	// Do not modify cached copy
	return proto.DeepCopy(), nil
}
