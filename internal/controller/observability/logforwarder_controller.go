package observability

import (
	"context"
	"time"

	log "github.com/ViaQ/logerr/v2/log/static"
	"github.com/go-logr/logr"
	obsv1beta1 "github.com/openshift/cluster-logging-operator/api/observability/v1beta1"
	internalcontext "github.com/openshift/cluster-logging-operator/internal/api/context"
	internalobs "github.com/openshift/cluster-logging-operator/internal/api/observability"
	"github.com/openshift/cluster-logging-operator/internal/components/logforwarder"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

const (
	lfDefaultLogLevel = 0
)

type LogForwarderReconciler struct {
	Log    logr.Logger
	Scheme *runtime.Scheme

	PollInterval time.Duration

	TimeOut time.Duration

	ForwarderContext internalcontext.ForwarderContext
}

// SetupWithManager sets up the controller with the Manager.
func (r *LogForwarderReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&obsv1beta1.LogForwarder{}).
		Owns(&corev1.ConfigMap{}).
		Owns(&appsv1.Deployment{}).
		Owns(&networkingv1.NetworkPolicy{}, builder.WithPredicates(predicate.GenerationChangedPredicate{})).
		Owns(&corev1.Service{}).
		Complete(r)
}

func (r *LogForwarderReconciler) Reconcile(_ context.Context, req ctrl.Request) (result ctrl.Result, err error) {
	r.Log.V(lfDefaultLogLevel).Info("Reconcile LogForwarder", "req", req)
	cxt := r.ForwarderContext
	// load
	var lf *internalobs.LogForwarder
	if lf, err = r.fetchLogForwarder(r.ForwarderContext.Client, req.NamespacedName); err != nil {
		return defaultRequeue, err
	}
	r.Log.V(lfDefaultLogLevel).Info("Fetched", "logForwarder", lf)
	clf := lf.AsClusterLogForwarder()
	cxt.Forwarder = &clf

	if lf.DeletionTimestamp != nil {
		// Resource is being deleted, no further reconciliation
		return defaultRequeue, nil
	}
	// removeStaleStatus
	// init
	if cxt, err = initialize(cxt); err != nil {
		return defaultRequeue, nil
	}
	options := utils.Options{}

	// Set options to any options added during initialization of CLF
	if cxt.AdditionalContext != nil {
		options = cxt.AdditionalContext
	}
	// validate (if needed)
	// reconcile
	//   eval outputs for SA token
	//   trusted CA bundle
	//   AWS profiles configmap
	//   generate config
	var forwarderConfig string
	if forwarderConfig, err = GenerateConfig(r.ForwarderContext.Client, lf, lf.ResourceNames(), cxt.Secrets, options); err != nil {
		log.V(lfDefaultLogLevel).Error(err, "forwarder.GenerateConfig")
		return defaultRequeue, err
	}
	r.Log.V(lfDefaultLogLevel).Info("Generated config", "config", forwarderConfig)
	if err = logforwarder.ReconcileConfig(r.ForwarderContext.Client, *lf, forwarderConfig, lf.ResourceNames().ConfigMap); err != nil {
		return defaultRequeue, err
	}

	//   create deployment
	if err = logforwarder.ReconcileDeployment(r.ForwarderContext.Client, *lf); err != nil {
		return defaultRequeue, err
	}
	//   NP

	//   In service + metrics
	if err = logforwarder.ReconcileService(r.ForwarderContext.Client, *lf); err != nil {
		return defaultRequeue, err
	}

	return periodicRequeue, nil
}

func (r *LogForwarderReconciler) fetchLogForwarder(k8sClient client.Client, objKey types.NamespacedName) (internallf *internalobs.LogForwarder, err error) {
	r.Log.V(lfDefaultLogLevel).Info("Fetching", "objKey", objKey)
	lf := &obsv1beta1.LogForwarder{}
	if err = k8sClient.Get(context.TODO(), objKey, lf); err != nil {
		return nil, err
	}
	return internalobs.NewLogForwarder(*lf), err
}
