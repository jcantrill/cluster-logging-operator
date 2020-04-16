package v1

import (
	"github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1/inputs"
	sets "k8s.io/apimachinery/pkg/util/sets"
)

// Reserved input names.
const (
	InputNameApplication    = "application"    // Non-infrastructure container logs.
	InputNameInfrastructure = "infrastructure" // Infrastructure containers and system logs.
	InputNameAudit          = "audit"          // System audit logs.
)

var ReservedInputNames = sets.NewString(InputNameApplication, InputNameInfrastructure, InputNameAudit)

func IsInputTypeName(s string) bool { return ReservedInputNames.Has(s) }

// InputSpec defines a source of log messages.
type InputSpec struct {
	// Name used to refer to the input of a `pipeline`.
	//
	// +required
	Name string `json:"name"`

	// Type of input source.
	//
	// +kubebuilder:validation:Enum:=application;infrastructure;audit
	// +required
	Type string `json:"type"`

	// InputTypeSpec is inlined with a required `type` and optional extra configuration.
	//
	// +optional
	InputTypeSpec `json:",inline,omitempty"`
}

// InputTypeSpec is a union of optional type-specific extra specs.
//
type InputTypeSpec struct {
	// Filter for application logs.
	// +optional
	Application *inputs.ApplicationType `json:"application,omitempty"`

	// TODO(alanconway) in future we may add other types of filter,
	// for example filtering infra or audit logs by some other critiera.
	// That's why the ApplicationType is a separate struct even though
	// there aren't any other types yet.
}
