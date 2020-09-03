package builder

import (
	"context"

	"github.com/openshift/cluster-logging-operator/test/helpers"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/client-go/kubernetes"
)

type ConfigMapBuilder struct {
	kubeclient *kubernetes.Clientset
	obj        *corev1.ConfigMap
}

func NewConfigMapBuilder(namespace, name string) *ConfigMapBuilder {
	kubeclient, _ := helpers.NewKubeClient()
	builder := &ConfigMapBuilder{
		kubeclient: kubeclient,
		obj: &corev1.ConfigMap{
			TypeMeta: metav1.TypeMeta{
				Kind:       "ConfigMap",
				APIVersion: corev1.SchemeGroupVersion.String(),
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
				Labels:    map[string]string{},
			},
			Data: map[string]string{},
		},
	}
	return builder
}
func (builder *ConfigMapBuilder) Add(key, value string) *ConfigMapBuilder {
	builder.obj.Data[key] = value
	return builder
}

func (builder *ConfigMapBuilder) Create() (*corev1.ConfigMap, error) {
	return builder.kubeclient.CoreV1().ConfigMaps(builder.obj.Namespace).
		Create(context.TODO(), builder.obj, metav1.CreateOptions{})
}
