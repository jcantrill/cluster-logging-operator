package observability

import (
	"context"
	"fmt"
	"time"

	log "github.com/ViaQ/logerr/v2/log/static"
	"github.com/go-logr/logr"
	obsv1 "github.com/openshift/cluster-logging-operator/api/observability/v1"
	obsv1beta1 "github.com/openshift/cluster-logging-operator/api/observability/v1beta1"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/factory"
	"github.com/openshift/cluster-logging-operator/internal/reconcile"
	obsruntime "github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/internal/runtime/observability"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"

	//rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	cldDefaultLogLevel = 0
	cldLabelComponent  = "distributor"
	cldLabelName       = "clusterlogdistributor"
)

type ClusterLogDistributorReconciler struct {
	Log          logr.Logger
	Client       client.Client
	Scheme       *runtime.Scheme
	PollInterval time.Duration
	TimeOut      time.Duration
}

// SetupWithManager sets up the controller with the Manager.
func (r *ClusterLogDistributorReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&obsv1beta1.ClusterLogDistributor{}).
		Owns(&obsv1.ClusterLogForwarder{}).
		Complete(r)
}

func (r *ClusterLogDistributorReconciler) Reconcile(cxt context.Context, req ctrl.Request) (result ctrl.Result, err error) {
	r.Log.V(cldDefaultLogLevel).Info("reconciling...", "request", req)

	var cld *obsv1beta1.ClusterLogDistributor
	if cld, err = r.LoadClusterLogDistributor(req.Name); err != nil {
		return defaultRequeue, err
	}

	// load LogForwarders
	// match LogForwarders to ClusterLogDistributors
	//    update LF.Status with CLD information
	//

	if err = r.ReconcileClusterLogForwarderForDistributor(cxt, *cld); err != nil {
		return defaultRequeue, err
	}

	//set status/transfer some status from CLF

	return defaultRequeue, nil
}

func (r *ClusterLogDistributorReconciler) LoadClusterLogDistributor(name string) (*obsv1beta1.ClusterLogDistributor, error) {
	key := types.NamespacedName{Name: name}
	cld := &obsv1beta1.ClusterLogDistributor{}
	if err := r.Client.Get(context.TODO(), key, cld); err != nil {
		return nil, err
	}
	r.Log.V(cldDefaultLogLevel).Info("loaded", "clusterlogdistributor", cld)
	return cld, nil
}

func (r *ClusterLogDistributorReconciler) ReconcileClusterLogForwarderForDistributor(cxt context.Context, cld obsv1beta1.ClusterLogDistributor) (err error) {

	if err = ServiceAccount(r.Client, cld); err != nil {
		return err
	}

	clf := initClusterLogForwarder(cld)
	r.Log.V(cldDefaultLogLevel).Info("initialized", "clusterlogforwarder", clf)
	err = r.Client.Create(cxt, clf) //TODO refactor to createOrUpdate
	if err != nil {
		log.V(cldDefaultLogLevel).Error(err, "Error creating clusterlogforwarder", clf)
		return err
	}
	log.V(cldDefaultLogLevel).Info("created", "clusterlogforwarder", clf)
	return nil
}

func ServiceAccount(k8sClient client.Client, cld obsv1beta1.ClusterLogDistributor) (err error) {
	name := cldObjectName(cld)
	sa := factory.NewServiceAccount(constants.OpenshiftNS, name, cldLabelName, name, cldLabelComponent)
	sa, err = reconcile.ServiceAccount(k8sClient, sa)
	if err != nil && !errors.IsAlreadyExists(err) {
		return err
	}

	//// TODO determine bindings needed for cld
	subject := obsruntime.NewSubject("ServiceAccount", sa.Name)
	subject.Namespace = constants.OpenshiftNS
	roleRef := rbacv1.RoleRef{
		APIGroup: "rbac.authorization.k8s.io",
		Kind:     "ClusterRole",
		Name:     "collect-infrastructure-logs",
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

func initClusterLogForwarder(cld obsv1beta1.ClusterLogDistributor) *obsv1.ClusterLogForwarder {
	clfName := cldObjectName(cld)
	clf := observability.NewClusterLogForwarder(constants.OpenshiftNS, clfName, obsruntime.Initialize,
		func(clf *obsv1.ClusterLogForwarder) {
			obsruntime.SetCommonLabels(clf, cldLabelName, cldObjectName(cld), cldLabelComponent)
			clf.Spec = obsv1.ClusterLogForwarderSpec{
				Inputs:    []obsv1.InputSpec{},
				Pipelines: []obsv1.PipelineSpec{},
				Outputs:   []obsv1.OutputSpec{},
				ServiceAccount: obsv1.ServiceAccount{
					Name: clfName,
				},
			}
			for _, i := range cld.Spec.Inputs {
				pipeline := obsv1.PipelineSpec{
					Name:       i.Name,
					OutputRefs: []string{"invalidRef"},
				}
				input := obsv1.InputSpec{
					Name: i.Name,
					Type: obsv1.InputTypeApplication,
					Application: &obsv1.Application{
						Includes: i.Includes,
						Excludes: i.Excludes,
						Selector: i.Selector,
					},
				}
				pipeline.InputRefs = []string{input.Name}
				clf.Spec.Inputs = append(clf.Spec.Inputs, input)
				clf.Spec.Pipelines = append(clf.Spec.Pipelines, pipeline)
			}
		},
	)
	utils.AddOwnerRefToObject(clf, utils.AsOwner(&cld))
	return clf
}
