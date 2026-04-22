package metrics

// collectorDropConfigs are Drop transform component metrics for the collector
var collectorDropConfigs = []MetricDropConfig{
	{
		LabelName:  "component_kind",
		LabelValue: "transform",
		ExcludeMetrics: []string{
			"vector_component_errors_total",
			"vector_component_discarded_events_total",
		},
	},
}
