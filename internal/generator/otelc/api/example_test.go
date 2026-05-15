package api_test

import (
	"fmt"

	"github.com/openshift/cluster-logging-operator/internal/generator/otelc/api"
	"github.com/openshift/cluster-logging-operator/internal/generator/otelc/api/exporters"
	"github.com/openshift/cluster-logging-operator/internal/generator/otelc/api/receivers"
	"github.com/openshift/cluster-logging-operator/internal/generator/otelc/api/receivers/operators"
	"github.com/openshift/cluster-logging-operator/internal/generator/otelc/api/types"
	"gopkg.in/yaml.v3"
)

func ExampleFileLog() {
	// Create a new FileLog receiver
	fileLog := receivers.NewFileLog("", "/var/log/app/*.log", "/var/log/service/*.log")
	fileLog.StartAt = "beginning"
	fileLog.Encoding = "utf-8"
	fileLog.MaxConcurrentFiles = 256
	fileLog.Attributes = map[string]interface{}{
		"log.type":    "application",
		"environment": "production",
	}

	// Add multiline support
	fileLog.Multiline = &receivers.Multiline{
		LineStartPattern: "^\\d{4}-\\d{2}-\\d{2}",
	}

	// Add operators for log processing
	fileLog.Operators = []operators.Operator{
		{
			Type: operators.OperatorTypeJSONParser,
			ID:   "json_parse",
		},
	}

	// Create receivers collection
	receiversMap := api.Receivers{
		"file_log/app": fileLog,
	}

	// Marshal to YAML
	data, err := yaml.Marshal(receiversMap)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(data))
	// Output:
	// file_log/app:
	//     include:
	//         - /var/log/app/*.log
	//         - /var/log/service/*.log
	//     start_at: beginning
	//     encoding: utf-8
	//     max_concurrent_files: 256
	//     multiline:
	//         line_start_pattern: ^\d{4}-\d{2}-\d{2}
	//     attributes:
	//         environment: production
	//         log.type: application
	//     operators:
	//         - type: json_parser
	//           id: json_parse
}

func ExampleFileLog_withRetry() {
	// Create a FileLog receiver with retry configuration
	fileLog := receivers.NewFileLog("", "/var/log/*.log")
	fileLog.RetryOnFailure = &receivers.RetryOnFailure{
		Enabled:         true,
		InitialInterval: "1s",
		MaxInterval:     "30s",
		MaxElapsedTime:  "5m",
	}

	data, err := yaml.Marshal(fileLog)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(data))
	// Output:
	// include:
	//     - /var/log/*.log
	// retry_on_failure:
	//     enabled: true
	//     initial_interval: 1s
	//     max_interval: 30s
	//     max_elapsed_time: 5m
}

func ExampleOtlpHttp() {
	// Create OtlpHttp exporter for Loki
	otlphttp := exporters.NewOtlpHttp("", "http://loki:3100/otlp")
	otlphttp.Encoding = "proto"
	otlphttp.Compression = "gzip"
	otlphttp.Headers = map[string]string{
		"X-Scope-OrgID": "tenant1",
	}
	otlphttp.TLS = &types.TLSClientConfig{
		Insecure: true,
	}

	// Create exporters collection
	exportersMap := api.Exporters{
		"otlphttp/loki": otlphttp,
	}

	// Marshal to YAML
	data, err := yaml.Marshal(exportersMap)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(data))
	// Output:
	// otlphttp/loki:
	//     endpoint: http://loki:3100/otlp
	//     tls:
	//         insecure: true
	//     headers:
	//         X-Scope-OrgID: tenant1
	//     compression: gzip
	//     encoding: proto
}

func ExampleOtlpHttp_withQueueAndRetry() {
	// Create OtlpHttp exporter with queue and retry settings
	otlphttp := exporters.NewOtlpHttp("", "http://loki:3100/otlp")
	otlphttp.SendingQueue = &types.QueueSettings{
		Enabled:      true,
		NumConsumers: 5,
		QueueSize:    1000,
	}
	otlphttp.RetryOnFailure = &types.RetrySettings{
		Enabled:         true,
		InitialInterval: "5s",
		MaxInterval:     "30s",
		MaxElapsedTime:  "5m",
	}

	data, err := yaml.Marshal(otlphttp)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(data))
	// Output:
	// endpoint: http://loki:3100/otlp
	// sending_queue:
	//     enabled: true
	//     num_consumers: 5
	//     queue_size: 1000
	// retry_on_failure:
	//     enabled: true
	//     initial_interval: 5s
	//     max_interval: 30s
	//     max_elapsed_time: 5m
}
