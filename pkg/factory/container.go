package factory

import (
	"github.com/openshift/cluster-logging-operator/pkg/utils"
	v1 "k8s.io/api/core/v1"
)

//NewContainer stubs an instance of a Container
func NewContainer(containerName string, imageName string, pullPolicy v1.PullPolicy, resources v1.ResourceRequirements) v1.Container {
	return v1.Container{
		Name:            containerName,
		Image:           utils.GetComponentImage(imageName),
		ImagePullPolicy: pullPolicy,
		Resources:       resources,
	}
}
