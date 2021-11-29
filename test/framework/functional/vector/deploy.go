package vector

import (
	"github.com/ViaQ/logerr/log"
	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	"github.com/openshift/cluster-logging-operator/test/client"
)

type VectorCollector struct {
	*client.Test
}

func (c *VectorCollector) DeployConfigMapForConfig(name, config, clfYaml string) error {
	log.V(2).Info("Creating config configmap")
	configmap := runtime.NewConfigMap(c.NS.Name, name, map[string]string{})
	runtime.NewConfigMapBuilder(configmap).
		Add("vector.toml", config).
		Add("clfyaml", clfYaml)
	if err := c.Create(configmap); err != nil {
		return err
	}
	return nil
}

func (c *VectorCollector) BuildCollectorContainer(b *runtime.ContainerBuilder, nodeName string) *runtime.ContainerBuilder {
	return b.AddEnvVar("LOG_LEVEL", "debug").
		AddEnvVarFromFieldRef("POD_IP", "status.podIP").
		AddEnvVar("NODE_NAME", nodeName).
		AddVolumeMount("config", "/etc/vector", "", true)
}

func (c *VectorCollector) IsStarted(logs string) bool {
	return true
}


func (c *VectorCollector) Image() string {
	return utils.GetComponentImage(constants.VectorName)
}
