package observability

import (
	"context"
	log "github.com/ViaQ/logerr/v2/log/static"
	obs "github.com/openshift/cluster-logging-operator/api/observability/v1"
	internalobs "github.com/openshift/cluster-logging-operator/internal/api/observability"
	"github.com/openshift/cluster-logging-operator/internal/auth"
	"github.com/openshift/cluster-logging-operator/internal/collector"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/factory"
	forwardergenerator "github.com/openshift/cluster-logging-operator/internal/generator/forwarder"
	"github.com/openshift/cluster-logging-operator/internal/generator/framework"
	generatorhelpers "github.com/openshift/cluster-logging-operator/internal/generator/helpers"
	"github.com/openshift/cluster-logging-operator/internal/generator/vector/helpers"
	"github.com/openshift/cluster-logging-operator/internal/metrics"
	"github.com/openshift/cluster-logging-operator/internal/reconcile"
	"github.com/openshift/cluster-logging-operator/internal/tls"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
)

func CreateOrUpdateCollector(k8Client client.Client, k8Reader client.Reader, spec obs.ClusterLogForwarder, clusterID string) (err error) {
	log.V(9).Info("Entering obs.CreateOrUpdateCollection")
	defer func() {
		log.V(9).Info("Leaving obs.CreateOrUpdateCollection")
	}()

	// TODO LOG-2620: containers violate PodSecurity ?

	if err = reconcile.SecurityContextConstraints(k8Client, auth.NewSCC()); err != nil {
		log.V(3).Error(err, "reconcile.SecurityContextConstraints")
		return err
	}

	ownerRef := utils.AsOwner(&spec)
	resourceNames := factory.ResourceNames(spec)

	// Add roles to ServiceAccount to allow the collector to read from the node
	if err = auth.ReconcileRBAC(noOpEventRecorder, k8Client, spec.Namespace, resourceNames, ownerRef); err != nil {
		log.V(3).Error(err, "auth.ReconcileRBAC")
		return
	}

	// TODO enable me and sync with podspec generation
	//if err := collector.ReconcileTrustedCABundleConfigMap(clusterRequest.EventRecorder, clusterRequest.Client, clusterRequest.Forwarder.Namespace, clusterRequest.ResourceNames.CaTrustBundle, clusterRequest.ResourceOwner); err != nil {
	//	log.Error(err, "collector.ReconcileTrustedCABundleConfigMap")
	//	return err
	//}

	secrets, err := LoadSecrets(k8Client, spec.Namespace, spec.Spec.Inputs, spec.Spec.Outputs)
	if err != nil {
		log.V(3).Error(err, "auth.LoadSecrets")
		return err
	}

	var collectorConfig string
	if collectorConfig, err = GenerateConfig(k8Client, spec, *resourceNames, secrets); err != nil {
		log.V(9).Error(err, "clusterRequest.generateCollectorConfig")
		return err
	}
	log.V(3).Info("Generated collector config", "config", collectorConfig)
	var collectorConfHash string
	collectorConfHash, err = utils.CalculateMD5Hash(collectorConfig)
	if err != nil {
		log.Error(err, "unable to calculate MD5 hash")
		log.V(9).Error(err, "Returning from unable to calculate MD5 hash")
		return
	}
	isDaemonset := true
	factory := collector.New(collectorConfHash, clusterID, spec.Spec.Collector, secrets, spec.Spec, resourceNames, isDaemonset, LogLevel(spec.Annotations))
	if err = factory.ReconcileCollectorConfig(noOpEventRecorder, k8Client, k8Reader, spec.Namespace, collectorConfig, ownerRef); err != nil {
		log.Error(err, "collector.ReconcileCollectorConfig")
		return
	}

	reconcileDeployment := factory.ReconcileDaemonset
	if !isDaemonset {
		reconcileDeployment = factory.ReconcileDeployment
	}
	if err := reconcileDeployment(noOpEventRecorder, k8Client, spec.Namespace, ownerRef); err != nil {
		log.Error(err, "Error reconciling the deployment of the collector")
		return err
	}

	if err := metrics.ReconcileServiceMonitor(noOpEventRecorder, k8Client, spec.Namespace, resourceNames.CommonName, constants.CollectorName, collector.MetricsPortName, ownerRef); err != nil {
		log.Error(err, "collector.ReconcileServiceMonitor")
		return err
	}

	return nil
}

func GenerateConfig(k8Client client.Client, spec obs.ClusterLogForwarder, resourceNames factory.ForwarderResourceNames, secrets helpers.Secrets) (config string, err error) {
	op := framework.Options{}
	tlsProfile, _ := tls.FetchAPIServerTlsProfile(k8Client)
	op[framework.ClusterTLSProfileSpec] = tls.GetClusterTLSProfileSpec(tlsProfile)
	//EvaluateAnnotationsForEnabledCapabilities(clusterRequest.Forwarder, op)
	g := forwardergenerator.New()
	generatedConfig, err := g.GenerateConf(secrets, spec.Spec, spec.Namespace, spec.Name, resourceNames, op)

	if err != nil {
		log.Error(err, "Unable to generate log configuration")
		return "", err
	}

	log.V(3).Info("ClusterLogForwarder generated config", generatedConfig)
	return generatedConfig, err
	return "", nil
}

func LoadSecrets(k8Client client.Client, namespace string, inputs internalobs.Inputs, outputs internalobs.Outputs) (secrets helpers.Secrets, err error) {
	for _, name := range inputs.SecretNames() {
		secret := &corev1.Secret{}
		if err = k8Client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: namespace}, secret); err == nil {
			secrets[name] = secret
		} else {
			return secrets, err
		}
	}
	return secrets, nil
}

// EvaluateAnnotationsForEnabledCapabilities populates generator options with capabilities enabled by the ClusterLogForwarder
func EvaluateAnnotationsForEnabledCapabilities(annotations map[string]string, options framework.Options) {
	if annotations == nil {
		return
	}
	for key, value := range annotations {
		switch key {
		case constants.AnnotationDebugOutput:
			if strings.ToLower(value) == "true" {
				options[generatorhelpers.EnableDebugOutput] = "true"
			}
		}
	}
}

func LogLevel(annotations map[string]string) string {
	if level, ok := annotations[constants.AnnotationVectorLogLevel]; ok {
		return level
	}
	return "warn"
}
