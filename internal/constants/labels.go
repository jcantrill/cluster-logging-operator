package constants

const (

	// K8s recommended label names: https://kubernetes.io/docs/concepts/overview/working-with-objects/common-labels/

	// LabelK8sName The name of the application (string)
	LabelK8sName = "app.kubernetes.io/name"
	// LabelK8sInstance A unique name identifying the instance of an application (string)
	LabelK8sInstance = "app.kubernetes.io/instance"
	// LabelK8sVersion The current version of the application (e.g., a semantic version, revision hash, etc.) (string)
	LabelK8sVersion = "app.kubernetes.io/version"
	// LabelK8sComponent The component within the architecture (string)
	LabelK8sComponent = "app.kubernetes.io/component"
	// LabelK8sPartOf The name of a higher level application this one is part of (string)
	LabelK8sPartOf = "app.kubernetes.io/part-of"
	// LabelK8sManagedBy The tool being used to manage the operation of an application (string)
	LabelK8sManagedBy = "app.kubernetes.io/managed-by"

	LabelLoggingServiceType      = "logging.observability.openshift.io/service-type"
	LabelLoggingInputServiceType = "logging.observability.openshift.io/input-service-type"

	ServiceTypeMetrics = "metrics"
	ServiceTypeInput   = "input"
)
