package input_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/openshift/cluster-logging-operator/internal/generator/otelc/api/receivers"
	"github.com/openshift/cluster-logging-operator/internal/generator/otelc/api/receivers/operators"
	"github.com/openshift/cluster-logging-operator/internal/generator/otelc/input"
	"gopkg.in/yaml.v3"
)

var _ = Describe("CRI-O Container Input Operators", func() {
	Context("NewCRIOOperators", func() {
		It("should create a complete operator pipeline", func() {
			ops := input.NewCRIOOperators()

			// Verify we have all expected operators
			Expect(ops).To(HaveLen(11))

			// Verify operator IDs in order
			expectedIDs := []string{
				"parser-crio",
				"extract_metadata_from_filepath",
				"move_log_to_body",
				"set_iostream",
				"move_namespace_to_resource",
				"move_pod_name_to_resource",
				"move_pod_uid_to_resource",
				"move_container_name_to_resource",
				"move_restart_count_to_resource",
				"remove_logtag",
				"remove_time",
			}

			for i, expectedID := range expectedIDs {
				Expect(ops[i].ID).To(Equal(expectedID), "Operator %d should have ID %s", i, expectedID)
			}
		})

		It("should marshal to YAML correctly", func() {
			ops := input.NewCRIOOperators()

			data, err := yaml.Marshal(ops)
			Expect(err).To(BeNil())

			yamlStr := string(data)

			// Verify key components are present
			Expect(yamlStr).To(ContainSubstring("type: regex_parser"))
			Expect(yamlStr).To(ContainSubstring("id: parser-crio"))
			Expect(yamlStr).To(ContainSubstring("type: move"))
			Expect(yamlStr).To(ContainSubstring("type: remove"))
			Expect(yamlStr).To(ContainSubstring("k8s.namespace.name"))
			Expect(yamlStr).To(ContainSubstring("k8s.pod.name"))
			Expect(yamlStr).To(ContainSubstring("k8s.container.name"))
			Expect(yamlStr).To(ContainSubstring("log.iostream"))
		})

		It("should create proper CRI-O regex pattern", func() {
			ops := input.NewCRIOOperators()

			// First operator should be the CRI-O parser
			Expect(ops[0].Type).To(Equal(operators.OperatorTypeRegexParser))
			Expect(ops[0].ID).To(Equal("parser-crio"))

			config := ops[0].Config
			Expect(config).To(HaveKey("regex"))
			regex := config["regex"].(string)

			// Verify regex captures time, stream, logtag, and log
			Expect(regex).To(ContainSubstring("(?P<time>"))
			Expect(regex).To(ContainSubstring("(?P<stream>stdout|stderr)"))
			Expect(regex).To(ContainSubstring("(?P<logtag>"))
			Expect(regex).To(ContainSubstring("(?P<log>"))
		})
	})

	Context("NewCRIOOperatorsWithNodeName", func() {
		It("should add node name operator", func() {
			ops := input.NewCRIOOperatorsWithNodeName("worker-1")

			// Should have one more operator than base
			Expect(ops).To(HaveLen(12))

			// Last operator should add node name
			lastOp := ops[len(ops)-1]
			Expect(lastOp.Type).To(Equal(operators.OperatorTypeAdd))
			Expect(lastOp.ID).To(Equal("add_node_name"))

			config := lastOp.Config
			Expect(config).To(HaveKeyWithValue("field", "resource[\"k8s.node.name\"]"))
			Expect(config).To(HaveKeyWithValue("value", "worker-1"))
		})
	})

	Context("NewCRIOOperatorsWithOpenShiftLabels", func() {
		It("should add OpenShift resource attributes", func() {
			ops := input.NewCRIOOperatorsWithOpenShiftLabels(
				"cluster-123",
				"container",
				"application",
			)

			// Should have three more operators than base
			Expect(ops).To(HaveLen(14))

			// Verify cluster UID operator
			clusterOp := ops[11]
			Expect(clusterOp.Type).To(Equal(operators.OperatorTypeAdd))
			Expect(clusterOp.ID).To(Equal("add_cluster_uid"))
			Expect(clusterOp.Config).To(HaveKeyWithValue("field", "resource[\"openshift.cluster.uid\"]"))
			Expect(clusterOp.Config).To(HaveKeyWithValue("value", "cluster-123"))

			// Verify log source operator
			sourceOp := ops[12]
			Expect(sourceOp.Type).To(Equal(operators.OperatorTypeAdd))
			Expect(sourceOp.ID).To(Equal("add_log_source"))
			Expect(sourceOp.Config).To(HaveKeyWithValue("field", "resource[\"openshift.log.source\"]"))
			Expect(sourceOp.Config).To(HaveKeyWithValue("value", "container"))

			// Verify log type operator
			typeOp := ops[13]
			Expect(typeOp.Type).To(Equal(operators.OperatorTypeAdd))
			Expect(typeOp.ID).To(Equal("add_log_type"))
			Expect(typeOp.Config).To(HaveKeyWithValue("field", "resource[\"openshift.log.type\"]"))
			Expect(typeOp.Config).To(HaveKeyWithValue("value", "application"))
		})

		It("should marshal to YAML with OpenShift attributes", func() {
			ops := input.NewCRIOOperatorsWithOpenShiftLabels(
				"cluster-456",
				"container",
				"infrastructure",
			)

			data, err := yaml.Marshal(ops)
			Expect(err).To(BeNil())

			yamlStr := string(data)
			Expect(yamlStr).To(ContainSubstring("openshift.cluster.uid"))
			Expect(yamlStr).To(ContainSubstring("openshift.log.source"))
			Expect(yamlStr).To(ContainSubstring("openshift.log.type"))
			Expect(yamlStr).To(ContainSubstring("cluster-456"))
			Expect(yamlStr).To(ContainSubstring("infrastructure"))
		})
	})

	Context("Complete FileLog Receiver with CRI-O Operators", func() {
		It("should create a valid configuration", func() {
			// Create a FileLog receiver with CRI-O operators
			fileLog := receivers.NewFileLog("", "/var/log/pods/*/*/*.log")
			fileLog.Exclude = []string{"/var/log/pods/*/otel-collector/*.log"}
			fileLog.StartAt = "end"
			fileLog.IncludeFilePath = true

			// Add CRI-O operators with OpenShift labels
			fileLog.Operators = input.NewCRIOOperatorsWithOpenShiftLabels(
				"my-cluster",
				"container",
				"application",
			)

			data, err := yaml.Marshal(fileLog)
			Expect(err).To(BeNil())

			yamlStr := string(data)

			// Verify receiver configuration
			Expect(yamlStr).To(ContainSubstring("include:"))
			Expect(yamlStr).To(ContainSubstring("/var/log/pods/*/*/*.log"))
			Expect(yamlStr).To(ContainSubstring("exclude:"))
			Expect(yamlStr).To(ContainSubstring("start_at: end"))

			// Verify operators are included
			Expect(yamlStr).To(ContainSubstring("operators:"))
			Expect(yamlStr).To(ContainSubstring("parser-crio"))
			Expect(yamlStr).To(ContainSubstring("k8s.namespace.name"))
			Expect(yamlStr).To(ContainSubstring("openshift.cluster.uid"))
		})
	})
})
