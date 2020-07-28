package collector

import (
	"fmt"
	"reflect"

	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/pkg/factory"
	api "github.com/openshift/cluster-logging-operator/pkg/k8shandler/clients"
	"github.com/openshift/cluster-logging-operator/pkg/logger"
	"github.com/openshift/cluster-logging-operator/pkg/utils"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/util/retry"
)

const (
	CollectorName     = "collector"
	CollectorConfName = "collector-conf"
)

func NewContainer() v1.Container {
	container := factory.NewContainer(
		CollectorName,
		CollectorName,
		v1.PullIfNotPresent,
		v1.ResourceRequirements{},
	)
	container.VolumeMounts = []v1.VolumeMount{
		{Name: "runlogjournal", MountPath: "/run/log/journal"},
		{Name: "varlog", MountPath: "/var/log"},
		{Name: CollectorConfName, MountPath: "/etc/fluent-bit"},
	}
	container.SecurityContext = &v1.SecurityContext{
		Privileged: utils.GetBool(true),
	}
	return container
}

//GenerateCollectorConfig creates configuration for the collector
func GenerateConfig() (string, error) {
	return string(utils.GetFileContents(utils.GetShareDir() + "/fluent-bit/fluent-bit.conf")), nil
}

func ReconcileConfigMap(apiClient api.ApiGateway, cluster *logging.ClusterLogging, namespace, config string) error {
	logger.Debug("Reconcile collector configmap...")
	configMap := factory.NewConfigMap(
		CollectorName,
		namespace,
		map[string]string{
			"fluent-bit.conf": config,
			"parsers.conf":    string(utils.GetFileContents(utils.GetShareDir() + "/fluent-bit/parsers.conf")),
		},
	)

	utils.AddOwnerRefToObject(configMap, utils.AsOwner(cluster))

	err := apiClient.Create(configMap)
	if err != nil && !errors.IsAlreadyExists(err) {
		return fmt.Errorf("Failure constructing collector configmap: %v", err)
	}
	retryErr := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		current := &v1.ConfigMap{}
		if err = apiClient.Get(configMap.Name, current); err != nil {
			if errors.IsNotFound(err) {
				logrus.Debugf("Returning nil. The configmap %q was not found even though create previously failed.  Was it culled?", configMap.Name)
				return nil
			}
			return fmt.Errorf("Failed to get %v configmap: %v", configMap.Name, err)
		}
		if reflect.DeepEqual(configMap.Data, current.Data) {
			return nil
		}
		current.Data = configMap.Data
		return apiClient.Update(current)
	})

	return retryErr
}
