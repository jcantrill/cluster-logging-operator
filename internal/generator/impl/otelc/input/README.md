# OpenTelemetry Collector Input Package

This package contains input-specific configurations for OpenTelemetry Collector receivers, using the operator types defined in `internal/generator/otelc/api/receivers/operators`.

## Purpose

The `input` package provides high-level, opinionated operator pipelines for common log input sources. These functions create complete operator chains that:
1. Parse input-specific log formats
2. Extract metadata
3. Map to standardized, non-deprecated attributes
4. Transform data for compatibility with observability backends

## Structure

### Container Logs (`container.go`)

Functions for parsing CRI-O container logs from Kubernetes pods, mapping them to the attributes defined in the [Red Hat Observability Data Model](https://github.com/rhobs/observability-data-model/blob/main/cluster-logging.md).

**Available Functions:**

- **`NewCRIOOperators()`** - Base CRI-O parser pipeline (11 operators)
  - Parses CRI-O log format: `<timestamp> <stream> <tags> <log>`
  - Extracts Kubernetes metadata from file path
  - Maps to non-deprecated attributes (k8s.*, log.iostream)

- **`NewCRIOOperatorsWithNodeName(nodeName string)`** - Adds `k8s.node.name` attribute
  - Use when node name is available via environment variable or downward API

- **`NewCRIOOperatorsWithOpenShiftLabels(clusterUID, logSource, logType string)`** - Adds OpenShift attributes
  - `openshift.cluster.uid` (required)
  - `openshift.log.source` (required)
  - `openshift.log.type` (required)

## Attribute Mapping

### Non-Deprecated Attributes

The operator pipelines map CRI-O container logs to these **non-deprecated** attributes:

#### Resource Attributes (same for all records in a stream)
- `k8s.namespace.name` ← extracted from file path
- `k8s.pod.name` ← extracted from file path
- `k8s.pod.uid` ← extracted from file path
- `k8s.container.name` ← extracted from file path
- `k8s.container.restart_count` ← extracted from file path
- `k8s.node.name` ← provided via parameter (optional)
- `openshift.cluster.uid` ← provided via parameter (optional)
- `openshift.log.source` ← provided via parameter (optional)
- `openshift.log.type` ← provided via parameter (optional)

#### Log Attributes (can vary per record)
- `log.iostream` ← parsed from CRI-O stream field (stdout/stderr)

### Deprecated Attributes (NOT used)

These attributes from the legacy ViaQ data model are **not** used:
- ❌ `kubernetes.namespace_name` → use `k8s.namespace.name`
- ❌ `kubernetes.pod_name` → use `k8s.pod.name`
- ❌ `kubernetes.container_name` → use `k8s.container.name`
- ❌ `kubernetes.host` → use `k8s.node.name`
- ❌ `openshift.cluster_id` → use `openshift.cluster.uid`
- ❌ `log_type` → use `openshift.log.type`
- ❌ `log_source` → use `openshift.log.source`

## Usage Example

```go
package main

import (
    "github.com/openshift/cluster-logging-operator/internal/generator/otelc/api/receivers"
    "github.com/openshift/cluster-logging-operator/internal/generator/otelc/input"
    "gopkg.in/yaml.v3"
)

func main() {
    // Create FileLog receiver
    fileLog := receivers.NewFileLog("/var/log/pods/*/*/*.log")
    fileLog.StartAt = "end"
    fileLog.IncludeFilePath = true
    
    // Add CRI-O operators with OpenShift attributes
    fileLog.Operators = input.NewCRIOOperatorsWithOpenShiftLabels(
        "my-cluster-uuid",
        "container",
        "application",
    )
    
    // Marshal to YAML
    data, _ := yaml.Marshal(fileLog)
    println(string(data))
}
```

### Output

```yaml
include:
  - /var/log/pods/*/*/*.log
start_at: end
include_file_path: true
operators:
  - type: regex_parser
    id: parser-crio
    output: extract_metadata_from_filepath
    regex: ^(?P<time>[^ Z]+) (?P<stream>stdout|stderr) (?P<logtag>[^ ]*) ?(?P<log>.*)$
    timestamp:
      parse_from: attributes.time
      layout_type: gotime
      layout: "2006-01-02T15:04:05.999999999Z07:00"
  # ... (11 total operators for base pipeline)
  # ... (3 additional operators for OpenShift labels)
```

## Testing

Run the test suite:

```bash
go test -v github.com/openshift/cluster-logging-operator/internal/generator/otelc/input
```

## References

- [Red Hat Observability Data Model - Cluster Logging](https://github.com/rhobs/observability-data-model/blob/main/cluster-logging.md)
- [OpenTelemetry Collector FileLog Receiver](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/filelogreceiver)
- [CRI-O Log Format](https://github.com/cri-o/cri-o/blob/main/docs/crio.8.md#logging)
