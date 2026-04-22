package metrics

// collectorAllowMetrics are allowed metrics for the collector
var collectorAllowMetrics = []string{
	// Metrics used in alerts (collector_alerts.yaml)
	"logcollector_component_event_unmatched_count",
	"vector_http_client_errors_total",
	"vector_http_client_requests_sent_total",
	"vector_http_client_responses_total",
	"vector_buffer_byte_size",
	"vector_component_errors_total",
	"vector_component_received_events_total",

	// Metrics used in dashboards (openshift-logging-dashboard.json)
	"vector_component_received_bytes_total",
	"vector_component_sent_bytes_total",
	"vector_component_received_event_bytes_total",
	"vector_open_files",
	"vector_component_discarded_events_total",

	// Metrics used in telemetry rules (telemetry_rules.yaml)
	// (vector_component_received_bytes_total already listed above)

	// Note: Kubernetes metrics (up, node_*, kubelet_*, container_*, csv_succeeded)
	// and LFME metrics (log_logged_bytes_total) are scraped by other ServiceMonitors
}
