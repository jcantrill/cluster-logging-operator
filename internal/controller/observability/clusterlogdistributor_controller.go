package observability

import (
	"context"
	"time"

	"github.com/go-logr/logr"
	obsv1 "github.com/openshift/cluster-logging-operator/api/observability/v1"
	obsv1beta1 "github.com/openshift/cluster-logging-operator/api/observability/v1beta1"
	internalobs "github.com/openshift/cluster-logging-operator/internal/api/observability"
	"github.com/openshift/cluster-logging-operator/internal/components/clusterlogdistributor"
	"github.com/openshift/cluster-logging-operator/internal/components/logforwarder"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	cldDefaultLogLevel = 0
)

var (
	cldRequeue = ctrl.Result{
		RequeueAfter: time.Minute * 1,
	}
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
		return cldRequeue, err
	}

	// load LogForwarders
	var logforwarders []internalobs.LogForwarder
	if logforwarders, err = logforwarder.List(r.Client, map[string]string{clusterlogdistributor.LabelLogDisributorService: cld.Name}); err != nil {
		return cldRequeue, err
	}

	if err = clusterlogdistributor.ReconcileClusterLogForwarderForDistributor(r.Client, *cld, logforwarders); err != nil {
		return cldRequeue, err
	}
	// match LogForwarders to ClusterLogDistributors
	//    update LF.Status with CLD information
	//

	//set status/transfer some status from CLF

	return cldRequeue, nil
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
