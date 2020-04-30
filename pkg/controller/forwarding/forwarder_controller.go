package forwarding

import (
	"context"
	"time"

	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/pkg/constants"
	"github.com/openshift/cluster-logging-operator/pkg/k8shandler"
	"github.com/openshift/cluster-logging-operator/pkg/logger"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// Add creates a new Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileForwarder{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("clusterlogforwarder-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource ClusterLogging
	err = c.Watch(&source.Kind{Type: &logging.ClusterLogForwarder{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileForwarder{}

// ReconcileForwarder reconciles a ClusterLogForwarder object
type ReconcileForwarder struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

var (
	reconcilePeriod = 30 * time.Second
	reconcileResult = reconcile.Result{RequeueAfter: reconcilePeriod}
)

// Reconcile reads that state of the cluster for a ClusterLogForwarder object and makes changes based on the state read
// and what is in the Logging.Spec
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileForwarder) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	logger.DebugObject("clusterlogforwarder-controller Reconciling: %v", request)
	// Fetch the ClusterLogForwarder instance
	instance := &logging.ClusterLogForwarder{}
	logger.Debug("clusterlogforwarder-controller fetching LF instance")
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		logger.Debugf("clusterlogforwarder-controller Error getting instance. It will be retried if other then 'NotFound': %v", err)
		if !errors.IsNotFound(err) {
			// Error reading the - requeue the request.
			return reconcileResult, err
		}

		// else the object is not found -- meaning it was removed so do clean up manually
		reconcileErr := k8shandler.ReconcileForClusterLogForwarder(instance, r.client)

		if reconcileErr != nil {
			logger.Debugf("clusterlogforwarder-controller returning, error: %v", reconcileErr)
		}
		return reconcile.Result{}, reconcileErr
	}

	logger.DebugObject("clusterlogforwarder-controller fetched LF instance: %v", instance)

	s := *logging.NewClusterLogForwarderStatus()
	instance.Status = s

	if instance.Name != constants.SingletonName {
		s.SetNewCondition(logging.ConditionReady, false, logging.ReasonInvalid,
			"Invalid name %q, singleton instance must be named %q",
			instance.Name, constants.SingletonName)
		return r.updateStatus(instance, nil)
	}

	logger.Debug("clusterlogforwarder-controller calling ClusterLogging reconciler...")
	reconcileErr := k8shandler.ReconcileForClusterLogForwarder(instance, r.client)

	if reconcileErr != nil {
		logger.Debugf("clusterlogforwarder-controller returning, error: %v", reconcileErr)
	}
	if instance.Status.IsReady() {
		s.SetNewCondition(logging.ConditionReady, true, "", "")
	} else {
		s.SetNewCondition(logging.ConditionDegraded, true, "There are one or more pipelines that are unavailable", "")
	}
	return r.updateStatus(instance, reconcileErr)
}

func (r *ReconcileForwarder) updateStatus(instance *logging.ClusterLogForwarder, reconcileErr error) (reconcile.Result, error) {
	logger.DebugObject("clusterlogforwarder-controller updating status of instance: %v, error: %v", instance, reconcileErr)
	if err := r.client.Status().Update(context.TODO(), instance); err != nil {
		logger.Debugf("clusterlogforwarder-controller error updating status: %v", err)
		return reconcileResult, reconcileErr
	}

	return reconcile.Result{}, reconcileErr
}
