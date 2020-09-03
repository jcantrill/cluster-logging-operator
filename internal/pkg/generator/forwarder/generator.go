package forwarder

import (
	"fmt"
	"os"

	yaml "gopkg.in/yaml.v2"

	logging "github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1"
	"github.com/openshift/cluster-logging-operator/pkg/generators/forwarding"
	"github.com/openshift/cluster-logging-operator/pkg/k8shandler"
	"github.com/openshift/cluster-logging-operator/pkg/logger"
)

const (
	//these are fixed at the moment
	logCollectorType         = logging.LogCollectionTypeFluentd
	includeLegacyForward     = false
	includeLegacySyslog      = false
	useOldRemoteSyslogPlugin = false
)

func Generate(clfYaml string, includeDefaultLogStore bool) (string, error) {

	generator, err := forwarding.NewConfigGenerator(
		logCollectorType,
		includeLegacyForward,
		includeLegacySyslog,
		useOldRemoteSyslogPlugin)
	if err != nil {
		return "", fmt.Errorf("Unable to create collector config generator: %v", err)
	}

	forwarder := &logging.ClusterLogForwarder{}
	err = yaml.Unmarshal([]byte(clfYaml), forwarder)
	if err != nil {
		logger.Errorf("Error Unmarshalling %q: %v", clfYaml, err)
		os.Exit(1)
	}
	logger.DebugObject("Unmarshalled %s", forwarder)
	clRequest := &k8shandler.ClusterLoggingRequest{
		ForwarderSpec: forwarder.Spec,
		Cluster: &logging.ClusterLogging{
			Spec: logging.ClusterLoggingSpec{},
		},
	}
	if includeDefaultLogStore {
		clRequest.Cluster.Spec.LogStore = &logging.LogStoreSpec{
			Type: logging.LogStoreTypeElasticsearch,
		}
	}
	spec, status := clRequest.NormalizeForwarder()
	logger.DebugObject("Normalization Status: %s", status)
	tunings := &logging.ForwarderSpec{}

	generatedConfig, err := generator.Generate(spec, tunings)
	if err != nil {
		return "", fmt.Errorf("Unable to generate log configuration: %v", err)
	}
	return generatedConfig, nil
}
