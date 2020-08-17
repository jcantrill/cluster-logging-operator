package k8shandler

import (
	"fmt"

	configv1 "github.com/openshift/api/config/v1"
	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/pkg/constants"
	"github.com/openshift/cluster-logging-operator/pkg/k8shandler/collector"
	"github.com/openshift/cluster-logging-operator/pkg/logger"
	"github.com/openshift/cluster-logging-operator/pkg/utils"
	core "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/util/retry"
)

const (
	//LogForwardingTopologyAnnotation describes the topology to use for log collection.  The absence of the
	//annotation or a recognized value results in the default topology
	LogForwardingTopologyAnnotation = "clusterlogging.openshift.io/logforwardingTopology"

	//EnableTechPreviewLogForwardingTopologyAnnotation setting the value to 'true' will cause the operator to evalute
	//the topology to be used in log forwarding
	EnableTechPreviewLogForwardingTopologyAnnotation = "clusterlogging.openshift.io/enableTechPreviewTopology"

	//LogForwardingDualEdgeNormalizationTopology deploys multiple containers to each node to collect and normalize log messagges
	LogForwardingDualEdgeNormalizationTopology = "dualEdgeNormalization"

	//LogForwardingEdgeNormalizationTopology is the default (legacy) topology to deploy a single container to each node to collect and normalize log messages.
	LogForwardingEdgeNormalizationTopology = "edgeNormalization"

	//LogForwardingCentralNormalizationTopology deploys a single container to each node to collect and forward messages
	//to a centralized log normalizer
	LogForwardingCentralNormalizationTopology = "centralNormalization"

	collectorServiceAccountName = "logCollector"

	collectorName        = "fluentd"
	componentName        = "fluentd"
	loggingComponentName = "fluentd"
)

func NewTopology(clusterRequest *ClusterLoggingRequest, proxyConfig *configv1.Proxy, pipelineSpec logging.ClusterLogForwarderSpec, trustedCABundleCM *core.ConfigMap) LogForwardingTopology {
	topology := LogForwardingEdgeNormalizationTopology
	switch topology {
	default:
		return EdgeNormalizationTopology{
			clusterRequest.cluster,
			clusterRequest,
			proxyConfig,
			pipelineSpec,
			trustedCABundleCM,
		}
	}
}

func (clusterRequest *ClusterLoggingRequest) ReconcileLogForwardingTopology(proxyConfig *configv1.Proxy) (err error) {
	cluster := clusterRequest.cluster

	caTrustBundle := &v1.ConfigMap{}
	// Create or update cluster proxy trusted CA bundle.
	if proxyConfig != nil {
		caTrustBundle, err = clusterRequest.createOrGetTrustedCABundleConfigMap(constants.FluentdTrustedCAName)
		if err != nil {
			return
		}
	}
	topology := NewTopology(clusterRequest, proxyConfig, clusterRequest.ForwarderSpec, caTrustBundle)
	normalizerConfig, err := topology.generateNormalizerConfig()
	if err != nil {
		return err
	}

	logger.Debugf("Generated normalizer config: %s", normalizerConfig)
	normalizerConfHash, err := utils.CalculateMD5Hash(normalizerConfig)
	if err != nil {
		logger.Errorf("unable to calculate MD5 hash. E: %s", err.Error())
		return err
	}
	collectorConfig, err := topology.generateCollectorConfig()
	if err != nil {
		return err
	}
	logger.Debugf("Generated collector config: %s", collectorConfig)
	collectorConfHash, err := utils.CalculateMD5Hash(collectorConfig)
	if err != nil {
		logger.Errorf("unable to calculate MD5 hash. E: %s", err.Error())
		return err
	}

	if err = clusterRequest.reconcileNormalizer(cluster, topology.newNormalizerPodSpec(), normalizerConfHash, collectorConfHash); err != nil {
		logger.Errorf("Error reconciling normalizer: %v", err)
	}
	if err = reconcileCollector(topology.newCollectorPodSpec()); err != nil {
		logger.Errorf("Error reconciling collector: %v", err)
	}
	return err
}

func reconcileCollector(podSpec *core.PodSpec) error {
	if podSpec == nil {
		return nil
	}
	return nil
}

func (clusterRequest *ClusterLoggingRequest) reconcileNormalizer(cluster *logging.ClusterLogging, podSpec *core.PodSpec, pipelineConfHash, collectorConfHash string) error {
	if podSpec == nil {
		return nil
	}
	daemonset := NewDaemonSet(collectorName, cluster.Namespace, loggingComponentName, componentName, *podSpec)
	daemonset.Spec.Template.Spec.Containers[0].Env = updateEnvVar(v1.EnvVar{Name: "FLUENT_CONF_HASH", Value: pipelineConfHash}, daemonset.Spec.Template.Spec.Containers[0].Env)
	daemonset.Spec.Template.Spec.Containers[1].Env = updateEnvVar(v1.EnvVar{Name: "CONF_HASH", Value: collectorConfHash}, daemonset.Spec.Template.Spec.Containers[1].Env)

	annotations, err := clusterRequest.getFluentdAnnotations(daemonset)
	if err != nil {
		return err
	}

	daemonset.Spec.Template.Annotations = annotations

	uid := getServiceAccountLogCollectorUID()
	if len(uid) == 0 {
		// There's no uid for logcollector serviceaccount; setting ClusterLogging for the ownerReference.
		utils.AddOwnerRefToObject(daemonset, utils.AsOwner(cluster))
	} else {
		// There's a uid for logcollector serviceaccount; setting the ServiceAccount for the ownerReference with blockOwnerDeletion.
		utils.AddOwnerRefToObject(daemonset, NewLogCollectorServiceAccountRef(uid))
	}

	err = clusterRequest.Create(daemonset)
	if err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("Failure creating Fluentd Daemonset %v", err)
	}

	if clusterRequest.isManaged() {
		retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
			return clusterRequest.updateFluentdDaemonsetIfRequired(daemonset)
		})
		if retryErr != nil {
			return retryErr
		}
	}

	return nil
}

type LogForwardingTopology interface {
	// GetProxyConfig() *configv1.Proxy
	// GetClusterLogForwarderSpec() logging.ClusterLogForwarderSpec
	// NewCollectorContainer(resources *core.ResourceRequirements) core.Container
	// NewNormalizerContainer(resources *core.ResourceRequirements, outputs []logging.OutputSpec, proxyConfig *configv1.Proxy) core.Container

	newNormalizerPodSpec() *core.PodSpec
	newCollectorPodSpec() *core.PodSpec
	newContainers(trustedCAVolumeMount core.VolumeMount) []v1.Container
	generateCollectorConfig() (string, error)
	generateNormalizerConfig() (string, error)
}

//EdgeNormalizationTopology creates a topology where collection and normalization are located on the nodes and utilize the
//same containers.  This is the default (legacy) topology
type EdgeNormalizationTopology struct {
	cluster           *logging.ClusterLogging
	clusterRequest    *ClusterLoggingRequest
	ProxyConfig       *configv1.Proxy
	pipelineSpec      logging.ClusterLogForwarderSpec
	TrustedCABundleCM *core.ConfigMap
}

func (topology EdgeNormalizationTopology) newCollectorPodSpec() *core.PodSpec {
	return nil
}
func (topology EdgeNormalizationTopology) generateCollectorConfig() (string, error) {
	return topology.clusterRequest.generateCollectorConfig()
}
func (topology EdgeNormalizationTopology) generateNormalizerConfig() (string, error) {
	return collector.GenerateConfig()
}

func (topology EdgeNormalizationTopology) newContainers(trustedCAVolumeMount core.VolumeMount) []v1.Container {
	cluster := topology.cluster
	collectionSpec := logging.CollectionSpec{}
	if cluster.Spec.Collection != nil {
		collectionSpec = *cluster.Spec.Collection
	}
	resources := collectionSpec.Logs.FluentdSpec.Resources
	if resources == nil {
		resources = &v1.ResourceRequirements{
			Limits: v1.ResourceList{v1.ResourceMemory: defaultFluentdMemory},
			Requests: v1.ResourceList{
				v1.ResourceMemory: defaultFluentdMemory,
				v1.ResourceCPU:    defaultFluentdCpuRequest,
			},
		}
	}
	container := newNormalizerContainer(resources, topology.pipelineSpec.Outputs, topology.ProxyConfig, trustedCAVolumeMount)
	return []v1.Container{container}
}

func (topology EdgeNormalizationTopology) newNormalizerPodSpec() *core.PodSpec {
	cluster := topology.cluster
	collectionSpec := logging.CollectionSpec{}
	if cluster.Spec.Collection != nil {
		collectionSpec = *cluster.Spec.Collection
	}

	addTrustedCAVolume := false
	trustedCAVolumeMount := core.VolumeMount{}
	// If trusted CA bundle ConfigMap exists and its hash value is non-zero, mount the bundle.
	if topology.TrustedCABundleCM != nil && hasTrustedCABundle(topology.TrustedCABundleCM) {
		addTrustedCAVolume = true
		trustedCAVolumeMount = v1.VolumeMount{
			Name:      constants.FluentdTrustedCAName,
			ReadOnly:  true,
			MountPath: constants.TrustedCABundleMountDir,
		}
	}
	tolerations := utils.AppendTolerations(
		collectionSpec.Logs.FluentdSpec.Tolerations,
		[]v1.Toleration{
			{
				Key:      "node-role.kubernetes.io/master",
				Operator: v1.TolerationOpExists,
				Effect:   v1.TaintEffectNoSchedule,
			},
			{
				Key:      "node.kubernetes.io/disk-pressure",
				Operator: v1.TolerationOpExists,
				Effect:   v1.TaintEffectNoSchedule,
			},
		},
	)

	podSpec := NewPodSpec(
		collectorServiceAccountName,
		topology.newContainers(trustedCAVolumeMount),
		[]v1.Volume{
			{Name: "runlogjournal", VolumeSource: v1.VolumeSource{HostPath: &v1.HostPathVolumeSource{Path: "/run/log/journal"}}},
			{Name: "varlog", VolumeSource: v1.VolumeSource{HostPath: &v1.HostPathVolumeSource{Path: "/var/log"}}},
			{Name: "varlibdockercontainers", VolumeSource: v1.VolumeSource{HostPath: &v1.HostPathVolumeSource{Path: "/var/lib/docker"}}},
			{Name: "config", VolumeSource: v1.VolumeSource{ConfigMap: &v1.ConfigMapVolumeSource{LocalObjectReference: v1.LocalObjectReference{Name: "fluentd"}}}},
			{Name: collector.CollectorConfName, VolumeSource: v1.VolumeSource{ConfigMap: &v1.ConfigMapVolumeSource{LocalObjectReference: v1.LocalObjectReference{Name: collector.CollectorName}}}},
			{Name: "secureforwardconfig", VolumeSource: v1.VolumeSource{ConfigMap: &v1.ConfigMapVolumeSource{LocalObjectReference: v1.LocalObjectReference{Name: "secure-forward"}, Optional: utils.GetBool(true)}}},
			{Name: "secureforwardcerts", VolumeSource: v1.VolumeSource{Secret: &v1.SecretVolumeSource{SecretName: "secure-forward", Optional: utils.GetBool(true)}}},
			{Name: "syslogconfig", VolumeSource: v1.VolumeSource{ConfigMap: &v1.ConfigMapVolumeSource{LocalObjectReference: v1.LocalObjectReference{Name: syslogName}, Optional: utils.GetBool(true)}}},
			{Name: "syslogcerts", VolumeSource: v1.VolumeSource{Secret: &v1.SecretVolumeSource{SecretName: syslogName, Optional: utils.GetBool(true)}}},
			{Name: "entrypoint", VolumeSource: v1.VolumeSource{ConfigMap: &v1.ConfigMapVolumeSource{LocalObjectReference: v1.LocalObjectReference{Name: "fluentd"}}}},
			{Name: "certs", VolumeSource: v1.VolumeSource{Secret: &v1.SecretVolumeSource{SecretName: "fluentd"}}},
			{Name: "localtime", VolumeSource: v1.VolumeSource{HostPath: &v1.HostPathVolumeSource{Path: "/etc/localtime"}}},
			{Name: "dockercfg", VolumeSource: v1.VolumeSource{HostPath: &v1.HostPathVolumeSource{Path: "/etc/sysconfig/docker"}}},
			{Name: "dockerdaemoncfg", VolumeSource: v1.VolumeSource{HostPath: &v1.HostPathVolumeSource{Path: "/etc/docker"}}},
			{Name: "filebufferstorage", VolumeSource: v1.VolumeSource{HostPath: &v1.HostPathVolumeSource{Path: "/var/lib/fluentd"}}},
			{Name: metricsVolumeName, VolumeSource: v1.VolumeSource{Secret: &v1.SecretVolumeSource{SecretName: "fluentd-metrics"}}},
		},
		collectionSpec.Logs.FluentdSpec.NodeSelector,
		tolerations,
	)
	for _, target := range topology.pipelineSpec.Outputs {
		if target.Secret != nil && target.Secret.Name != "" {
			podSpec.Volumes = append(podSpec.Volumes, v1.Volume{Name: target.Name, VolumeSource: v1.VolumeSource{Secret: &v1.SecretVolumeSource{SecretName: target.Secret.Name}}})
		}
	}

	if addTrustedCAVolume {
		podSpec.Volumes = append(podSpec.Volumes,
			v1.Volume{
				Name: constants.FluentdTrustedCAName,
				VolumeSource: v1.VolumeSource{
					ConfigMap: &v1.ConfigMapVolumeSource{
						LocalObjectReference: v1.LocalObjectReference{
							Name: constants.FluentdTrustedCAName,
						},
						Items: []v1.KeyToPath{
							{
								Key:  constants.TrustedCABundleKey,
								Path: constants.TrustedCABundleMountFile,
							},
						},
					},
				},
			})
	}

	podSpec.PriorityClassName = clusterLoggingPriorityClassName
	// Shorten the termination grace period from the default 30 sec to 10 sec.
	podSpec.TerminationGracePeriodSeconds = utils.GetInt64(10)

	if topology.pipelineSpec.HasDefaultOutput() && cluster.Spec.LogStore != nil {
		podSpec.InitContainers = []v1.Container{
			newFluentdInitContainer(cluster),
		}
	} else {
		podSpec.InitContainers = []v1.Container{}
	}

	return &podSpec
}

func newNormalizerContainer(resources *core.ResourceRequirements, outputs []logging.OutputSpec, proxyConfig *configv1.Proxy, trustedCAVolumeMount core.VolumeMount) core.Container {
	container := NewContainer(fluentdName, fluentdName, v1.PullIfNotPresent, *resources)
	container.Ports = []v1.ContainerPort{
		{
			Name:          metricsPortName,
			ContainerPort: metricsPort,
			Protocol:      v1.ProtocolTCP,
		},
	}

	container.Env = []v1.EnvVar{
		{Name: "NODE_NAME", ValueFrom: &v1.EnvVarSource{FieldRef: &v1.ObjectFieldSelector{APIVersion: "v1", FieldPath: "spec.nodeName"}}},
		{Name: "MERGE_JSON_LOG", Value: "false"},
		{Name: "PRESERVE_JSON_LOG", Value: "true"},
		{Name: "K8S_HOST_URL", Value: "https://kubernetes.default.svc"},
		{Name: "METRICS_CERT", Value: "/etc/fluent/metrics/tls.crt"},
		{Name: "METRICS_KEY", Value: "/etc/fluent/metrics/tls.key"},
		{Name: "BUFFER_QUEUE_LIMIT", Value: "32"},
		{Name: "BUFFER_SIZE_LIMIT", Value: "8m"},
		{Name: "FILE_BUFFER_LIMIT", Value: "256Mi"},
		{Name: "FLUENTD_CPU_LIMIT", ValueFrom: &v1.EnvVarSource{ResourceFieldRef: &v1.ResourceFieldSelector{ContainerName: "fluentd", Resource: "limits.cpu"}}},
		{Name: "FLUENTD_MEMORY_LIMIT", ValueFrom: &v1.EnvVarSource{ResourceFieldRef: &v1.ResourceFieldSelector{ContainerName: "fluentd", Resource: "limits.memory"}}},
		{Name: "NODE_IPV4", ValueFrom: &v1.EnvVarSource{FieldRef: &v1.ObjectFieldSelector{APIVersion: "v1", FieldPath: "status.hostIP"}}},
		{Name: "CDM_KEEP_EMPTY_FIELDS", Value: "message"}, // by default, keep empty messages
	}
	proxyEnv := utils.SetProxyEnvVars(proxyConfig)
	container.Env = append(container.Env, proxyEnv...)
	container.VolumeMounts = []v1.VolumeMount{
		{Name: "runlogjournal", MountPath: "/run/log/journal"},
		{Name: "varlog", MountPath: "/var/log"},
		{Name: "varlibdockercontainers", ReadOnly: true, MountPath: "/var/lib/docker"},
		{Name: "config", ReadOnly: true, MountPath: "/etc/fluent/configs.d/user"},
		{Name: "secureforwardconfig", ReadOnly: true, MountPath: "/etc/fluent/configs.d/secure-forward"},
		{Name: "secureforwardcerts", ReadOnly: true, MountPath: "/etc/ocp-forward"},
		{Name: "syslogconfig", ReadOnly: true, MountPath: "/etc/fluent/configs.d/syslog"},
		{Name: "syslogcerts", ReadOnly: true, MountPath: "/etc/ocp-syslog"},
		{Name: "entrypoint", ReadOnly: true, MountPath: "/opt/app-root/src/run.sh", SubPath: "run.sh"},
		{Name: "certs", ReadOnly: true, MountPath: "/etc/fluent/keys"},
		{Name: "localtime", ReadOnly: true, MountPath: "/etc/localtime"},
		{Name: "dockercfg", ReadOnly: true, MountPath: "/etc/sysconfig/docker"},
		{Name: "dockerdaemoncfg", ReadOnly: true, MountPath: "/etc/docker"},
		{Name: "filebufferstorage", MountPath: "/var/lib/fluentd"},
		{Name: metricsVolumeName, MountPath: "/etc/fluent/metrics"},
		trustedCAVolumeMount,
	}
	for _, target := range outputs {
		if target.Secret != nil && target.Secret.Name != "" {
			path := fmt.Sprintf("/var/run/ocp-collector/secrets/%s", target.Secret.Name)
			container.VolumeMounts = append(container.VolumeMounts, v1.VolumeMount{Name: target.Name, MountPath: path})
		}
	}

	container.SecurityContext = &v1.SecurityContext{
		Privileged: utils.GetBool(true),
	}
	return container
}
