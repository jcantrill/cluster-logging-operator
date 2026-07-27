package otel

import (
	"path"

	"github.com/openshift/cluster-logging-operator/internal/collector/common"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/factory"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	corev1 "k8s.io/api/core/v1"
)

func CollectorVisitor(collectorContainer *corev1.Container, podSpec *corev1.PodSpec, resNames *factory.ForwarderResourceNames, namespace, logLevel string) {
	collectorContainer.Env = append(collectorContainer.Env,
		corev1.EnvVar{Name: "OTEL_LOG_LEVEL", Value: logLevel},
	)

	if collectorContainer.SecurityContext != nil && collectorContainer.SecurityContext.Capabilities != nil {
		collectorContainer.SecurityContext.Capabilities.Add = append(
			collectorContainer.SecurityContext.Capabilities.Add,
			"DAC_READ_SEARCH",
		)
	}

	dataPath := GetDataPath(namespace, resNames.ForwarderName)
	collectorContainer.VolumeMounts = append(collectorContainer.VolumeMounts,
		corev1.VolumeMount{Name: common.ConfigVolumeName, ReadOnly: true, MountPath: configPath},
		corev1.VolumeMount{Name: common.DataDir, ReadOnly: false, MountPath: dataPath},
	)

	collectorContainer.Image = utils.GetComponentImage(constants.ComponentNameOtelc)
	collectorContainer.Command = []string{"/otelcol-contrib"}
	collectorContainer.Args = []string{"--config=" + path.Join(configPath, ConfigFile)}

	podSpec.Volumes = append(podSpec.Volumes,
		corev1.Volume{Name: common.ConfigVolumeName, VolumeSource: corev1.VolumeSource{ConfigMap: &corev1.ConfigMapVolumeSource{LocalObjectReference: corev1.LocalObjectReference{Name: resNames.ConfigMap}}}},
		corev1.Volume{Name: common.DataDir, VolumeSource: corev1.VolumeSource{HostPath: &corev1.HostPathVolumeSource{Path: dataPath}}},
	)
}
