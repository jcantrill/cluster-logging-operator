package input_test

import (
	"fmt"

	"github.com/openshift/cluster-logging-operator/internal/generator/otelc/api"
	"github.com/openshift/cluster-logging-operator/internal/generator/otelc/api/receivers"
	"github.com/openshift/cluster-logging-operator/internal/generator/otelc/input"
	"gopkg.in/yaml.v3"
)

func ExampleNewCRIOOperators() {
	// Create a FileLog receiver for container logs
	fileLog := receivers.NewFileLog("", "/var/log/pods/*/*/*.log")
	fileLog.StartAt = "end"
	fileLog.IncludeFilePath = true

	// Add CRI-O parsing operators
	fileLog.Operators = input.NewCRIOOperators()

	// Create receivers collection
	receiversMap := api.Receivers{
		"file_log/containers": fileLog,
	}

	// Marshal to YAML
	data, err := yaml.Marshal(receiversMap)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(data))
	// Output:
	// file_log/containers:
	//     include:
	//         - /var/log/pods/*/*/*.log
	//     start_at: end
	//     include_file_path: true
	//     operators:
	//         - type: regex_parser
	//           id: parser-crio
	//           output: extract_metadata_from_filepath
	//           regex: ^(?P<time>[^ Z]+) (?P<stream>stdout|stderr) (?P<logtag>[^ ]*) ?(?P<log>.*)$
	//           timestamp:
	//             layout: 2006-01-02T15:04:05.999999999Z07:00
	//             layout_type: gotime
	//             parse_from: attributes.time
	//         - type: regex_parser
	//           id: extract_metadata_from_filepath
	//           output: move_log_to_body
	//           cache:
	//             size: 128
	//           parse_from: attributes["log.file.path"]
	//           regex: ^.*\/(?P<namespace>[^_]+)_(?P<pod_name>[^_]+)_(?P<uid>[^\/]+)\/(?P<container_name>[^\/]+)\/(?P<restart_count>\d+)\.log$
	//         - type: move
	//           id: move_log_to_body
	//           output: set_iostream
	//           from: attributes.log
	//           to: body
	//         - type: move
	//           id: set_iostream
	//           output: move_namespace_to_resource
	//           from: attributes.stream
	//           to: attributes["log.iostream"]
	//         - type: move
	//           id: move_namespace_to_resource
	//           output: move_pod_name_to_resource
	//           from: attributes.namespace
	//           to: resource["k8s.namespace.name"]
	//         - type: move
	//           id: move_pod_name_to_resource
	//           output: move_pod_uid_to_resource
	//           from: attributes.pod_name
	//           to: resource["k8s.pod.name"]
	//         - type: move
	//           id: move_pod_uid_to_resource
	//           output: move_container_name_to_resource
	//           from: attributes.uid
	//           to: resource["k8s.pod.uid"]
	//         - type: move
	//           id: move_container_name_to_resource
	//           output: move_restart_count_to_resource
	//           from: attributes.container_name
	//           to: resource["k8s.container.name"]
	//         - type: move
	//           id: move_restart_count_to_resource
	//           output: remove_logtag
	//           from: attributes.restart_count
	//           to: resource["k8s.container.restart_count"]
	//         - type: remove
	//           id: remove_logtag
	//           output: remove_time
	//           field: attributes.logtag
	//         - type: remove
	//           id: remove_time
	//           field: attributes.time
}

func ExampleNewCRIOOperatorsWithOpenShiftLabels() {
	// Create a FileLog receiver for OpenShift container logs
	fileLog := receivers.NewFileLog("", "/var/log/pods/*/*/*.log")
	fileLog.StartAt = "end"

	// Add CRI-O operators with OpenShift-specific attributes
	fileLog.Operators = input.NewCRIOOperatorsWithOpenShiftLabels(
		"cluster-abc123",
		"container",
		"application",
	)

	// Show just the OpenShift-specific attributes
	fmt.Println("OpenShift attributes added:")
	fmt.Println("- openshift.cluster.uid: cluster-abc123")
	fmt.Println("- openshift.log.source: container")
	fmt.Println("- openshift.log.type: application")
	// Output:
	// OpenShift attributes added:
	// - openshift.cluster.uid: cluster-abc123
	// - openshift.log.source: container
	// - openshift.log.type: application
}
