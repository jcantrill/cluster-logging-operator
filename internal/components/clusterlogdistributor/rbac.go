package clusterlogdistributor

import (
	"fmt"

	obsv1beta1 "github.com/openshift/cluster-logging-operator/api/observability/v1beta1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/factory"
	"github.com/openshift/cluster-logging-operator/internal/reconcile"
	obsruntime "github.com/openshift/cluster-logging-operator/internal/runtime"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func ServiceAccount(k8sClient client.Client, cld obsv1beta1.ClusterLogDistributor) (err error) {
	name := cldObjectName(cld)
	sa := factory.NewServiceAccount(constants.OpenshiftNS, name, cldLabelName, name, cldLabelComponent)
	sa, err = reconcile.ServiceAccount(k8sClient, sa)
	if err != nil && !errors.IsAlreadyExists(err) {
		return err
	}

	//// TODO determine bindings needed for cld
	subject := rbacv1.Subject{
		Kind:      "ServiceAccount",
		Namespace: constants.OpenshiftNS,
		Name:      sa.Name,
	}
	roleRef := rbacv1.RoleRef{
		APIGroup: "rbac.authorization.k8s.io",
		Kind:     "ClusterRole",
		Name:     "collect-application-logs",
	}
	clusterrolebinding := obsruntime.NewClusterRoleBinding(fmt.Sprintf("%s-collect-application-logs", name), roleRef, subject)
	err = reconcile.ClusterRoleBinding(k8sClient, clusterrolebinding.Name, func() *rbacv1.ClusterRoleBinding {
		return clusterrolebinding
	})
	if !errors.IsAlreadyExists(err) {
		return err
	}
	return nil
}

func cldObjectName(cld obsv1beta1.ClusterLogDistributor) string {
	return fmt.Sprintf("cld-%s", cld.Name)
}
