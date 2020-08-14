package k8shandler

import (
	"fmt"

	configv1 "github.com/openshift/api/config/v1"
	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/pkg/constants"
	"github.com/openshift/cluster-logging-operator/pkg/k8shandler/collector"
	"github.com/openshift/cluster-logging-operator/pkg/utils"
	core "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
)

const (
	//LogForwardingTopologyAnnotation describes the topology to use for log collection.  The absence of the
	//annotation or a recognized value results in the default topology
	LogForwardingTopologyAnnotation = "clusterlogging.openshift.io/logforwardingTopology"

	//EnableTechPreviewLogForwardingTopologyAnnotation setting the value to 'true' will cause the operator to evalute
	//the topology to be used in log forwarding
	EnableTechPreviewLogForwardingTopologyAnnotation = "clusterlogging.openshift.io/enableTechPreviewTopology"

	//DualEdgeNormalizationTopology deploys multiple containers to each node to collect and normalize log messagges
	DualEdgeNormalizationTopology = "dualEdgeNormalization"

	//EdgeNormalizationTopology is the default (legacy) topology to deploy a single container to each node to collect and normalize log messages.
	EdgeNormalizationTopology = "edgeNormalization"

	//CentralNormalizationTopology deploys a single container to each node to collect and forward messages
	//to a centralized log normalizer
	CentralNormalizationTopology = "centralNormalization"

	collectorServiceAccountName = "logCollector"
)

func ReconcileCollectionAndNormalization() {

}

func ReconcileCollector() {

}

func ReconcileNormalizer() {

}

type TopologyBuilder interface {
	// GetProxyConfig() *configv1.Proxy
	GetClusterLogForwarderSpec() logging.ClusterLogForwarderSpec
	NewCollectorContainer(resources *core.ResourceRequirements, nodeSelector map[string]string, tolerations []core.Toleration) core.Container
	NewNormalizerContainer(resources *core.ResourceRequirements, nodeSelector map[string]string, tolerations []core.Toleration) core.Container

	NewCollectorPodSpec() v1.PodSpec
}

type EdgeNormalizationTopologyBuilder struct {
	cluster           *logging.ClusterLogging
	ProxyConfig       *configv1.Proxy
	pipelineSpec      logging.ClusterLogForwarderSpec
	TrustedCABundleCM *core.ConfigMap
}

func (topology *EdgeNormalizationTopologyBuilder) NewCollectorPodSpec() core.PodSpec {
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
	fluentdContainer := topology.NewNormalizerContainer(resources, collectionSpec.Logs.FluentdSpec.NodeSelector, collectionSpec.Logs.FluentdSpec.Tolerations)
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
	addTrustedCAVolume := false
	// If trusted CA bundle ConfigMap exists and its hash value is non-zero, mount the bundle.
	if topology.TrustedCABundleCM != nil && hasTrustedCABundle(topology.TrustedCABundleCM) {
		addTrustedCAVolume = true
		fluentdContainer.VolumeMounts = append(fluentdContainer.VolumeMounts,
			v1.VolumeMount{
				Name:      constants.FluentdTrustedCAName,
				ReadOnly:  true,
				MountPath: constants.TrustedCABundleMountDir,
			})
	}
	fluentdPodSpec := NewPodSpec(
		collectorServiceAccountName,
		[]v1.Container{fluentdContainer},
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
			fluentdPodSpec.Volumes = append(fluentdPodSpec.Volumes, v1.Volume{Name: target.Name, VolumeSource: v1.VolumeSource{Secret: &v1.SecretVolumeSource{SecretName: target.Secret.Name}}})
		}
	}

	if addTrustedCAVolume {
		fluentdPodSpec.Volumes = append(fluentdPodSpec.Volumes,
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

	fluentdPodSpec.PriorityClassName = clusterLoggingPriorityClassName
	// Shorten the termination grace period from the default 30 sec to 10 sec.
	fluentdPodSpec.TerminationGracePeriodSeconds = utils.GetInt64(10)

	if topology.pipelineSpec.HasDefaultOutput() && cluster.Spec.LogStore != nil {
		fluentdPodSpec.InitContainers = []v1.Container{
			newFluentdInitContainer(cluster),
		}
	} else {
		fluentdPodSpec.InitContainers = []v1.Container{}
	}

	return fluentdPodSpec
}

func (topology *EdgeNormalizationTopologyBuilder) NewCollectorContainer(resources *core.ResourceRequirements, nodeSelector map[string]string, tolerations []core.Toleration) core.Container {

}

func (topology *EdgeNormalizationTopologyBuilder) NewNormalizerContainer(resources *core.ResourceRequirements, nodeSelector map[string]string) core.Container {
	fluentdContainer := NewContainer("fluentd", "fluentd", v1.PullIfNotPresent, *resources)
	fluentdContainer.Ports = []v1.ContainerPort{
		{
			Name:          metricsPortName,
			ContainerPort: metricsPort,
			Protocol:      v1.ProtocolTCP,
		},
	}

	fluentdContainer.Env = []v1.EnvVar{
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
	proxyEnv := utils.SetProxyEnvVars(topology.ProxyConfig)
	fluentdContainer.Env = append(fluentdContainer.Env, proxyEnv...)
	fluentdContainer.VolumeMounts = []v1.VolumeMount{
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
	}
	for _, target := range topology.ClusterLogForwarderSpec.Outputs {
		if target.Secret != nil && target.Secret.Name != "" {
			path := fmt.Sprintf("/var/run/ocp-collector/secrets/%s", target.Secret.Name)
			fluentdContainer.VolumeMounts = append(fluentdContainer.VolumeMounts, v1.VolumeMount{Name: target.Name, MountPath: path})
		}
	}

	fluentdContainer.SecurityContext = &v1.SecurityContext{
		Privileged: utils.GetBool(true),
	}
	return fluentdContainer
}
