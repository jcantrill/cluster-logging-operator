package forwarder

import (
	"fmt"

	yaml "gopkg.in/yaml.v2"

	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	loggingv1alpha1 "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1alpha1"
	"github.com/openshift/cluster-logging-operator/pkg/k8shandler"
	log "github.com/openshift/cluster-logging-operator/pkg/logger"
)

const (
	//these are fixed at the moment
	logCollectorType         = logging.LogCollectionTypeFluentd
	includeLegacySyslog      = false
	useOldRemoteSyslogPlugin = false
)

func Generate(clfYaml string, includeDefaultLogStore, includeLegacyForward bool) (string, error) {

	forwarder := &loggingv1alpha1.LogForwarding{}
	if clfYaml != "" {
		err := yaml.Unmarshal([]byte(clfYaml), forwarder)
		if err != nil {
			return "", fmt.Errorf("Error Unmarshalling %q: %v", clfYaml, err)
		}
	}
	log.Debugf("Unmarshalled", "forwarder", forwarder)
	clRequest := &k8shandler.ClusterLoggingRequest{
		ForwardingRequest: forwarder,
		Cluster: &logging.ClusterLogging{
			Spec: logging.ClusterLoggingSpec{
				Collection: &logging.CollectionSpec{
					Logs: logging.LogCollectionSpec{
						Type: logging.LogCollectionTypeFluentd,
					},
				},
			},
		},
	}
	if includeDefaultLogStore {
		clRequest.Cluster.Spec.LogStore = &logging.LogStoreSpec{
			Type: logging.LogStoreTypeElasticsearch,
		}
	}

	generatedConfig, err := clRequest.GenerateCollectorConfig(
		func() bool { return includeLegacyForward },
		func() bool { return includeLegacySyslog },
		func() bool { return useOldRemoteSyslogPlugin })
	if err != nil {
		return "", fmt.Errorf("Unable to generate log configuration: %v", err)
	}
	return generatedConfig, nil
}
