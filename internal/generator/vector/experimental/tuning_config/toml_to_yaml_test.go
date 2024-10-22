//go:build experimental

package tuning_config

import (
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	yaml "sigs.k8s.io/yaml"
)

var _ = Describe("#ParseToml", func() {

	const (
		source = `expire_metrics_secs = 60
data_dir = "/var/lib/vector/openshift-logging/my-forwarder"

[api]
enabled = true

[sources.internal_metrics]
type = "internal_metrics"

[sources.input_audit_host]
type = "file"
include = ["/var/log/audit/audit.log"]
host_key = "hostname"
glob_minimum_cooldown_ms = 15000
ignore_older_secs = 3600
max_read_bytes = 3145728
rotate_wait_secs = 5

[transforms.input_audit_host_meta]
type = "remap"
inputs = ["input_audit_host"]
source = '''
  .log_source = "auditd"
  .log_type = "audit"
'''

[sinks.output_kafka_receiver]
type = "kafka"
inputs = ["output_kafka_receiver_topic"]
bootstrap_servers = "broker1-kafka.svc.messaging.cluster.local:9092"
topic = "{{ _internal.output_kafka_receiver_topic }}"
healthcheck.enabled = false

[sinks.output_kafka_receiver.encoding]
codec = "json"
timestamp_format = "rfc3339"
except_fields = ["_internal"]

[sinks.output_kafka_receiver.tls]
enabled = true
min_tls_version = "VersionTLS12"
ciphersuites = "TLS_AES_128_GCM_SHA256,TLS_AES_256_GCM_SHA384,TLS_CHACHA20_POLY1305_SHA256,ECDHE-ECDSA-AES128-GCM-SHA256,ECDHE-RSA-AES128-GCM-SHA256,ECDHE-ECDSA-AES256-GCM-SHA384,ECDHE-RSA-AES256-GCM-SHA384,ECDHE-ECDSA-CHACHA20-POLY1305,ECDHE-RSA-CHACHA20-POLY1305,DHE-RSA-AES128-GCM-SHA256,DHE-RSA-AES256-GCM-SHA384"
key_file = "/var/run/ocp-collector/secrets/kafka-receiver-1/tls.key"
crt_file = "/var/run/ocp-collector/secrets/kafka-receiver-1/tls.crt"
ca_file = "/var/run/ocp-collector/secrets/kafka-receiver-1/ca-bundle.crt"

# Merge audit api and node logs and group by log_source
[transforms.output_otel_collector_groupby_source]
type = "reduce"
inputs = ["output_otel_collector_kubeapi","output_otel_collector_node","output_otel_collector_openshiftapi","output_otel_collector_ovn"]
expire_after_ms = 15000
max_events = 250
group_by = [".openshift.cluster_id",".openshift.log_type",".openshift.log_source"]
merge_strategies.resource = "retain"
merge_strategies.logRecords = "array"

[sinks.prometheus_output.tls]
enabled = true
key_file = "/etc/collector/metrics/tls.key"
crt_file = "/etc/collector/metrics/tls.crt"
min_tls_version = "VersionTLS12"
ciphersuites = "TLS_AES_128_GCM_SHA256"
`
		tuningYAML = `
inputs:
- name: audit
  params:
    max_line_bytes: 1024
    encoding.charset: utf-8 
outputs:
- name: otel-collector
  transformsParams:
    max_events/groupby: 10
- name: kafka-receiver
  params:
    tls.crt_file: abc/123.crt
    encoding.codec:  csv
    encoding.csv.delimiter: ";"
`
	)

	It("should replace generator values with overrides", func() {
		toml := ParseToml(source)

		tunings := &ExperimentalCLFTuning{}

		Expect(yaml.Unmarshal([]byte(tuningYAML), tunings)).To(Succeed())

		toml.Modify(*tunings)
		Expect(toml.String()).To(BeComparableTo(source), fmt.Sprintf("Actual: \n%s", toml.String()))
	})
})
