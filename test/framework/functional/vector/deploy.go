package vector

import (
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	"github.com/openshift/cluster-logging-operator/test/client"
)

type VectorCollector struct {
	*client.Test
}

func (c *VectorCollector) DeployConfigMapForConfig(name, config, clfYaml string) error {
	return nil
}

func (c *VectorCollector) BuildCollectorContainer(b *runtime.ContainerBuilder, nodeName string) *runtime.ContainerBuilder {
	return b
}

func (c *VectorCollector) IsStarted(logs string) bool {
	return true
}


func (c *VectorCollector) Image() string {
	return utils.GetComponentImage(constants.VectorName)
}
