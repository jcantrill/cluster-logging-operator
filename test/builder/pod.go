package builder

import (
	"context"

	"github.com/openshift/cluster-logging-operator/test/helpers"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type PodBuilder struct {
	kubeclient *kubernetes.Clientset
	obj        *corev1.Pod
}

func NewPodBuilder(namespace, name string) *PodBuilder {
	kubeclient, _ := helpers.NewKubeClient()
	builder := &PodBuilder{
		kubeclient: kubeclient,
		obj: &corev1.Pod{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Pod",
				APIVersion: corev1.SchemeGroupVersion.String(),
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
				Labels:    map[string]string{},
			},
			Spec: corev1.PodSpec{},
		},
	}
	return builder
}

type ContainerBuilder struct {
	container  corev1.Container
	podBuilder *PodBuilder
}

func (builder *ContainerBuilder) End() *PodBuilder {
	builder.podBuilder.obj.Spec.Containers = append(builder.podBuilder.obj.Spec.Containers, builder.container)
	return builder.podBuilder
}

func (builder *ContainerBuilder) AddVolumeMount(name, path, subPath string, readonly bool) *ContainerBuilder {
	builder.container.VolumeMounts = append(builder.container.VolumeMounts, corev1.VolumeMount{
		Name:      name,
		ReadOnly:  readonly,
		MountPath: path,
		SubPath:   subPath,
	})
	return builder
}

func (builder *ContainerBuilder) AddEnvVar(name, value string) *ContainerBuilder {
	builder.container.Env = append(builder.container.Env, corev1.EnvVar{
		Name:  name,
		Value: value,
	})
	return builder
}

func (builder *PodBuilder) AddContainer(name, image string) *ContainerBuilder {
	containerBuilder := ContainerBuilder{
		container: corev1.Container{
			Name:  name,
			Image: image,
		},
		podBuilder: builder,
	}
	return &containerBuilder
}

func (builder *PodBuilder) AddConfigMapVolume(name, configMapName string) *PodBuilder {
	builder.obj.Spec.Volumes = append(builder.obj.Spec.Volumes, corev1.Volume{
		Name: name,
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: configMapName,
				},
			},
		},
	})
	return builder
}

func (builder *PodBuilder) WithLabels(labels map[string]string) *PodBuilder {
	builder.obj.Labels = labels
	return builder
}

func (builder *PodBuilder) Create() (*corev1.Pod, error) {
	return builder.kubeclient.CoreV1().Pods(builder.obj.Namespace).
		Create(context.TODO(), builder.obj, metav1.CreateOptions{})
}
