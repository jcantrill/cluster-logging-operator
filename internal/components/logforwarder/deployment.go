package logforwarder

import (
	_ "embed"

	log "github.com/ViaQ/logerr/v2/log/static"
	internalobs "github.com/openshift/cluster-logging-operator/internal/api/observability"
	"github.com/openshift/cluster-logging-operator/internal/components/clusterlogdistributor"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/reconcile"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	appsv1 "k8s.io/api/apps/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	//go:embed deployment.yaml
	deploymentYaml string
)

func ReconcileDeployment(k8sClient client.Client, lf internalobs.LogForwarder) (err error) {

	deployment := &appsv1.Deployment{}
	utils.MustUnmarshal(deploymentYaml, deployment)
	log.V(0).Info("Unmarshalled deployment", "obj", deployment)
	deployment.Namespace = lf.Namespace()
	deployment.Name = lf.DeployedName()
	deployment.Labels[constants.LabelK8sInstance] = lf.Name()
	deployment.Labels[constants.LabelK8sName] = clusterlogdistributor.ApplicationName

	//deployment := runtime.NewDeployment(lf.Namespace(), lf.DeployedName(), lf.CommonLabelsInitializer())
	selector := deployment.Labels
	delete(selector, constants.LabelK8sVersion)
	runtime.NewDeploymentBuilder(deployment).
		WithSelector(selector).
		WithTemplateLabels(deployment.Labels)

	for _, v := range deployment.Spec.Template.Spec.Volumes {
		if v.Name == "config" {
			v.ConfigMap.Name = lf.DeployedName() + "-config"
		}
		if v.Name == "metrics" {
			v.Secret.SecretName = lf.DeployedName() + "-metrics"
		}
	}
	lf.AddOwnerRefTo(deployment)
	log.V(0).Info("Reconciling deployment", "obj", deployment)
	if err = reconcile.Deployment(k8sClient, deployment); err != nil {
		log.Error(err, "Failed to reconcile logforwarder deployment", "obj", deployment)
	}
	return err
}

//func newPodSpec(lf internalobs.LogForwarder) (spec corev1.PodSpec) {
//	spec.ServiceAccountName = "default"
//	if lf.Spec.ServiceAccount != nil {
//		spec.ServiceAccountName = lf.Spec.ServiceAccount.Name
//	}
//	if lf.Spec.Forwarder != nil {
//		spec.Tolerations = lf.Spec.Forwarder.Tolerations
//		spec.NodeSelector = lf.Spec.Forwarder.NodeSelector
//		spec.Affinity = lf.Spec.Forwarder.Affinity
//	}
//	spec.Containers = append(spec.Containers, newContainer(lf))
//	return spec
//}
//
//func newContainer(lf internalobs.LogForwarder, secretVolumes, configmapVolumes []string) corev1.Container {
//	container := runtime.NewContainer(
//		internalobs.LogForwarderComponentName,
//		utils.GetComponentImage(constants.VectorName),
//		corev1.PullIfNotPresent, lf.LogForwarder.Spec.Forwarder.Resources)
//
//	container.Env = []corev1.EnvVar{
//		//{Name: "COLLECTOR_CONF_HASH", Value: f.ConfigHash},
//		{Name: "K8S_NODE_NAME", ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{APIVersion: "v1", FieldPath: "spec.nodeName"}}},
//		{Name: "NODE_IPV4", ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{APIVersion: "v1", FieldPath: "status.hostIP"}}},
//		//{Name: "OPENSHIFT_CLUSTER_ID", Value: clusterID},
//		{Name: "POD_IP", ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{APIVersion: "v1", FieldPath: "status.podIP"}}},
//		{Name: "POD_IPS", ValueFrom: &corev1.EnvVarSource{FieldRef: &corev1.ObjectFieldSelector{APIVersion: "v1", FieldPath: "status.podIPs"}}},
//	}
//	container.Env = append(container.Env, utils.GetProxyEnvVars()...)
//
//	//container.VolumeMounts = []corev1.VolumeMount{
//	//	{Name: metricsVolumeName, ReadOnly: true, MountPath: metricsVolumePath},
//	//	{Name: tmpVolumeName, MountPath: tmpPath},
//	//}
//
//	collector.AddVolumeMounts(container, secretVolumes, common.SecretBasePath)
//	collector.AddVolumeMounts(container, configmapVolumes, func(name string) string {
//		return common.ConfigMapBasePath(strings.TrimPrefix(name, "config-"))
//	})
//
//	//if outputs.NeedServiceAccountToken() {
//	//	AddVolumeMounts(container, []string{saTokenVolumeName}, func(name string) string {
//	//		// projected sa tokens are created in their own 'token' directory at this path
//	//		return constants.ServiceAccountSecretPath
//	//	})
//	//}
//	return *container
//}
