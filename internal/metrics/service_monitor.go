package metrics

import (
	"fmt"
	"strings"

	"github.com/openshift/cluster-logging-operator/internal/constants"
	"github.com/openshift/cluster-logging-operator/internal/reconcile"
	"github.com/openshift/cluster-logging-operator/internal/runtime"
	"github.com/openshift/cluster-logging-operator/internal/utils"
	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	prometheusCAFile = "/etc/prometheus/configmaps/serving-certs-ca-bundle/service-ca.crt"
)

// MetricDropConfig represents configuration for dropping metrics based on label values
type MetricDropConfig struct {
	// LabelName is the label to match (e.g., "component_kind")
	LabelName string
	// LabelValue is the value to match for dropping (e.g., "transform")
	LabelValue string
	// ExcludeMetrics is a list of metric names to exclude from dropping
	ExcludeMetrics []string
}

// MetricAllowlistConfig represents an allowlist of metrics to keep
type MetricAllowlistConfig struct {
	// AllowedMetrics is a list of metric name patterns to keep (all others are dropped)
	AllowedMetrics []string
}

func newServiceMonitor(namespace, name string, owner metav1.OwnerReference, selector map[string]string, portName string, dropConfigs []MetricDropConfig, allowlistConfig *MetricAllowlistConfig) *monitoringv1.ServiceMonitor {
	// Start with the base relabel config that replaces `-` with `_` in metric names
	relabelConfigs := []*monitoringv1.RelabelConfig{
		{
			SourceLabels: []monitoringv1.LabelName{
				"__name__",
			},
			TargetLabel: "__name__",
			Regex:       "(.*)-(.*)",
			Replacement: "${1}_${2}",
		},
	}

	// Add allowlist config first (if specified)
	if allowlistConfig != nil {
		relabelConfigs = append(relabelConfigs, buildAllowlistRelabelConfig(*allowlistConfig))
	}

	// Add drop configs (can be combined with allowlist)
	// Drop configs are applied after allowlist, so they filter from the kept metrics
	for _, dropConfig := range dropConfigs {
		relabelConfigs = append(relabelConfigs, buildDropRelabelConfig(dropConfig))
	}

	var endpoint = []monitoringv1.Endpoint{
		{
			Port:   portName,
			Path:   "/metrics",
			Scheme: "https",
			TLSConfig: &monitoringv1.TLSConfig{
				CAFile: prometheusCAFile,
				SafeTLSConfig: monitoringv1.SafeTLSConfig{
					ServerName: fmt.Sprintf("%s.%s.svc", name, namespace),
				},
			},
			MetricRelabelConfigs: relabelConfigs,
		},
	}

	desired := runtime.NewServiceMonitor(namespace, name)
	desired.Spec = monitoringv1.ServiceMonitorSpec{
		JobLabel:  fmt.Sprintf("monitor-%s", name),
		Endpoints: endpoint,
		Selector: metav1.LabelSelector{
			MatchLabels: selector,
		},
		NamespaceSelector: monitoringv1.NamespaceSelector{
			MatchNames: []string{namespace},
		},
		PodTargetLabels: []string{
			constants.LabelK8sName,
			constants.LabelK8sComponent,
			constants.LabelK8sPartOf,
			constants.LabelK8sInstance,
		},
	}

	utils.AddOwnerRefToObject(desired, owner)

	return desired
}

// buildDropRelabelConfig creates a Prometheus relabel config that drops metrics
// with a specific label value, optionally excluding certain metrics from the drop
func buildDropRelabelConfig(config MetricDropConfig) *monitoringv1.RelabelConfig {
	sourceLabels := []monitoringv1.LabelName{
		monitoringv1.LabelName(config.LabelName),
	}

	var regex string
	if len(config.ExcludeMetrics) == 0 {
		// Simple case: drop all metrics with this label value
		regex = config.LabelValue
	} else {
		// Complex case: drop metrics with this label value except specific metric names
		// Use negative lookahead to exclude specific metrics
		sourceLabels = append(sourceLabels, "__name__")
		excludePattern := strings.Join(config.ExcludeMetrics, "|")
		regex = fmt.Sprintf("%s;(?!%s).*", config.LabelValue, excludePattern)
	}

	return &monitoringv1.RelabelConfig{
		SourceLabels: sourceLabels,
		Regex:        regex,
		Action:       "drop",
	}
}

// buildAllowlistRelabelConfig creates a Prometheus relabel config that only keeps
// metrics matching the allowed patterns (drops all others)
func buildAllowlistRelabelConfig(config MetricAllowlistConfig) *monitoringv1.RelabelConfig {
	// Create a regex pattern that matches any of the allowed metrics
	pattern := strings.Join(config.AllowedMetrics, "|")

	return &monitoringv1.RelabelConfig{
		SourceLabels: []monitoringv1.LabelName{
			"__name__",
		},
		Regex:  pattern,
		Action: "keep",
	}
}

func BuildSelector(component, instance string) map[string]string {
	return map[string]string{
		constants.LabelLoggingServiceType: constants.ServiceTypeMetrics,
		constants.LabelK8sComponent:       component,
		constants.LabelK8sInstance:        instance,
	}
}

// newCollectorServiceMonitor creates a ServiceMonitor for the log collector with metric filtering.
// It only keeps metrics that are used in alerts, dashboards, and telemetry rules,
// and further filters out transform component metrics except for errors and discards.
func newCollectorServiceMonitor(namespace, name string, owner metav1.OwnerReference, selector map[string]string, portName string) *monitoringv1.ServiceMonitor {
	allowlistConfig := &MetricAllowlistConfig{
		AllowedMetrics: collectorAllowMetrics,
	}
	return newServiceMonitor(namespace, name, owner, selector, portName, collectorDropConfigs, allowlistConfig)
}

// newLogFileMetricExporterServiceMonitor creates a ServiceMonitor for the log file metrics exporter
// without any metric filtering.
func newLogFileMetricExporterServiceMonitor(namespace, name string, owner metav1.OwnerReference, selector map[string]string, portName string) *monitoringv1.ServiceMonitor {
	return newServiceMonitor(namespace, name, owner, selector, portName, nil, nil)
}

// ReconcileCollectorServiceMonitor reconciles a ServiceMonitor for the collector with appropriate metric filtering
func ReconcileCollectorServiceMonitor(k8sClient client.Client, namespace, name string, owner metav1.OwnerReference, selector map[string]string, portName string) error {
	desired := newCollectorServiceMonitor(namespace, name, owner, selector, portName)
	return reconcile.ServiceMonitor(k8sClient, desired)
}

// ReconcileLogFileMetricExporterServiceMonitor reconciles a ServiceMonitor for the log file metric exporter
func ReconcileLogFileMetricExporterServiceMonitor(k8sClient client.Client, namespace, name string, owner metav1.OwnerReference, selector map[string]string, portName string) error {
	desired := newLogFileMetricExporterServiceMonitor(namespace, name, owner, selector, portName)
	return reconcile.ServiceMonitor(k8sClient, desired)
}

// ReconcileServiceMonitor reconciles a ServiceMonitor with no metric filtering.
// Deprecated: Use ReconcileCollectorServiceMonitor or ReconcileLogFileMetricExporterServiceMonitor instead.
func ReconcileServiceMonitor(k8sClient client.Client, namespace, name string, owner metav1.OwnerReference, selector map[string]string, portName string) error {
	desired := newServiceMonitor(namespace, name, owner, selector, portName, nil, nil)
	return reconcile.ServiceMonitor(k8sClient, desired)
}
