package webhooks

import (
	"context"
	"encoding/json"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	internalruntime "github.com/openshift/cluster-logging-operator/internal/runtime"
	"net/http" // Required for AdmissionResponse status codes
	"sigs.k8s.io/controller-runtime/pkg/webhook"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	// For JSON Patch operations
	jsonpatch "github.com/evanphx/json-patch"
)

// PodMutator implements webhook.Handler for mutating Pods.
// It should be registered with the controller-runtime Manager.
type PodMutator struct {
	Client  client.Client
	Scheme  *runtime.Scheme
	decoder admission.Decoder
}

// +kubebuilder:webhook:path=/mutate-v1-pod,mutating=true,failurePolicy=fail,groups="",resources=pods,verbs=create;update,versions=v1,name=mpod.clusterlogging.openshift.io,sideEffects=NoneOnDryRun

// Handle implements admission.Handler.
// It receives an admission request and returns an admission response.
func (pm *PodMutator) Handle(ctx context.Context, req admission.Request) admission.Response {
	klog.Info("Received admission request for Pod mutation")

	// Decode the Pod object from the AdmissionRequest
	pod := &corev1.Pod{}
	err := pm.decoder.Decode(req, pod)
	if err != nil {
		klog.Errorf("Failed to decode Pod object: %v", err)
		return admission.Errored(http.StatusBadRequest, err)
	}

	isClusterLoggingPod := internalruntime.Labels(pod).Includes(
		internalruntime.ObjectLabels{
			constants.LabelK8sPartOf:    constants.ClusterLogging,
			constants.LabelK8sManagedBy: constants.ClusterLoggingOperator},
	)

	requiresStorage := true

	if !isClusterLoggingPod || !requiresStorage {
		klog.Infof("Pod %s/%s in namespace %s is not a cluster-logging-operator component, skipping mutation.",
			pod.Name, pod.Namespace, req.Namespace)
		return admission.Allowed("No mutation needed for this pod.")
	}

	klog.Infof("Identified cluster-logging-operator pod %s/%s in namespace %s for mutation.",
		pod.Name, pod.Namespace, req.Namespace)

	// Create a copy of the original pod to apply changes to
	originalPodJSON, err := json.Marshal(pod)
	if err != nil {
		klog.Errorf("Failed to marshal original pod: %v", err)
		return admission.Errored(http.StatusInternalServerError, err)
	}

	// --- Apply your desired mutations here ---

	// Marshal the modified pod back to JSON
	modifiedPodJSON, err := json.Marshal(pod)
	if err != nil {
		klog.Errorf("Failed to marshal modified pod: %v", err)
		return admission.Errored(http.StatusInternalServerError, err)
	}

	// Generate the JSON Patch
	patch, err := jsonpatch.CreateMergePatch(originalPodJSON, modifiedPodJSON)
	if err != nil {
		klog.Errorf("Failed to create JSON patch: %v", err)
		return admission.Errored(http.StatusInternalServerError, err)
	}

	klog.Infof("Generated patch for pod %s/%s: %s", pod.Namespace, pod.Name, string(patch))

	return admission.PatchResponseFromRaw(req.Object.Raw, patch)
}

// InjectDecoder injects the decoder.
// This is required by the controller-runtime webhook framework.
func (pm *PodMutator) InjectDecoder(d admission.Decoder) error {
	pm.decoder = d
	return nil
}

// SetupWebhookWithManager sets up the webhook with the Manager.
func (pm *PodMutator) SetupWebhookWithManager(mgr ctrl.Manager) error {
	// The path in the webhook builder must match the path in the +kubebuilder:webhook marker
	mgr.GetWebhookServer().Register("/mutate-v1-pod", &webhook.Admission{Handler: pm})
	return nil
}
