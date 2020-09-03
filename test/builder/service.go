package builder

import (
	"context"

	"github.com/openshift/cluster-logging-operator/test/helpers"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
)

type ServiceBuilder struct {
	kubeclient *kubernetes.Clientset
	service    *corev1.Service
}

func NewServiceBuilder(namespace, name string) *ServiceBuilder {
	kubeclient, _ := helpers.NewKubeClient()
	builder := &ServiceBuilder{
		kubeclient: kubeclient,
		service: &corev1.Service{
			TypeMeta: metav1.TypeMeta{
				Kind:       "Service",
				APIVersion: corev1.SchemeGroupVersion.String(),
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
				Labels:    map[string]string{},
			},
			Spec: corev1.ServiceSpec{
				Selector: map[string]string{},
			},
		},
	}
	return builder
}

func (builder *ServiceBuilder) WithSelector(selector map[string]string) *ServiceBuilder {
	builder.service.Spec.Selector = selector
	return builder
}

func (builder *ServiceBuilder) AddServicePort(port int32, targetPort int) *ServiceBuilder {
	builder.service.Spec.Ports = append(builder.service.Spec.Ports, corev1.ServicePort{
		Port:       port,
		TargetPort: intstr.FromInt(targetPort),
	})
	return builder
}

func (builder *ServiceBuilder) Create() (*corev1.Service, error) {
	return builder.kubeclient.CoreV1().Services(builder.service.Namespace).
		Create(context.TODO(), builder.service, metav1.CreateOptions{})
}
