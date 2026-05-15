# OpenTelemetry Collector API

This package provides Go structs for generating OpenTelemetry Collector configurations with YAML-only serialization, similar to the vector API in `internal/generator/vector/api`.

## Structure

- `types/` - Core types and interfaces for receivers
- `receivers/` - Receiver implementations (currently: FileLog)
- `receivers.go` - Collection and unmarshaling logic for receivers

## Supported Receivers

### FileLog Receiver

The FileLog receiver reads logs from files on the filesystem. It is based on the [opentelemetry-collector-contrib filelog receiver](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/filelogreceiver).

#### Features

- File selection with include/exclude patterns
- Multiple reading behaviors (start position, polling interval, encoding)
- Multiline log support
- File metadata attributes
- Retry on failure
- Operator-based log processing
- Custom attributes and resource labels

## Usage Example

```go
package main

import (
    "fmt"
    "gopkg.in/yaml.v3"
    
    "github.com/openshift/cluster-logging-operator/internal/generator/otelc/api"
    "github.com/openshift/cluster-logging-operator/internal/generator/otelc/api/receivers"
)

func main() {
    // Create a new FileLog receiver
    fileLog := receivers.NewFileLog("/var/log/app/*.log", "/var/log/service/*.log")
    fileLog.StartAt = "beginning"
    fileLog.Encoding = "utf-8"
    fileLog.MaxConcurrentFiles = 256
    fileLog.Attributes = map[string]interface{}{
        "log.type": "application",
        "environment": "production",
    }
    
    // Add multiline support
    fileLog.Multiline = &receivers.Multiline{
        LineStartPattern: "^\\d{4}-\\d{2}-\\d{2}",
    }
    
    // Add operators for log processing
    fileLog.Operators = []receivers.Operator{
        {
            Type: "json_parser",
            ID:   "json_parse",
        },
    }
    
    // Create receivers collection
    receivers := api.Receivers{
        "file_log/app": fileLog,
    }
    
    // Marshal to YAML
    data, err := yaml.Marshal(receivers)
    if err != nil {
        panic(err)
    }
    
    fmt.Println(string(data))
}
```

### Output

```yaml
file_log/app:
  include:
    - /var/log/app/*.log
    - /var/log/service/*.log
  start_at: beginning
  encoding: utf-8
  max_concurrent_files: 256
  attributes:
    environment: production
    log.type: application
  multiline:
    line_start_pattern: ^\d{4}-\d{2}-\d{2}
  operators:
    - type: json_parser
      id: json_parse
```

## YAML-Only Serialization

Unlike the vector API which supports both TOML and YAML, this API only uses YAML serialization tags. This is because OpenTelemetry Collector uses YAML for its configuration files.

All structs use only `yaml:` tags with appropriate options:
- `yaml:"field_name"` - Required field
- `yaml:"field_name,omitempty"` - Optional field (omitted if zero value)
- Fields without tags are not serialized

## Testing

Run the tests:

```bash
cd internal/generator/otelc/api
go test -v
```

## References

- [OpenTelemetry Collector Contrib - FileLog Receiver](https://github.com/open-telemetry/opentelemetry-collector-contrib/tree/main/receiver/filelogreceiver)
- [OpenTelemetry Collector Configuration](https://opentelemetry.io/docs/collector/configuration/)
