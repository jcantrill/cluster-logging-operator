package v1

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/openshift/cluster-logging-operator/pkg/apis/logging/v1/outputs"
	sets "k8s.io/apimachinery/pkg/util/sets"
)

var ReservedOutputNames = sets.NewString(OutputNameDefault)

func IsOutputTypeName(s string) bool {
	_, ok := goNames[s]
	return ok || ReservedOutputNames.Has(s)
}

// Output defines a destination for log messages.
type OutputSpec struct {
	// Name used to refer to the output from a `pipeline`.
	//
	// +required
	Name string `json:"name"`

	// Type of output plugin, for example 'syslog'
	//
	// +required
	Type string `json:"type"`

	// OutputTypeSpec is a union of pointers to extra configuration
	// for specific output types.
	OutputTypeSpec `json:",inline"`

	// URL to send log messages to.
	//
	// Must be an absolute URL, with a scheme. Valid URL schemes depend on `type`.
	// Special schemes 'tcp', 'udp' and 'tls' are used for output types that don't
	// define their own URL scheme.  Example:
	//
	//     { type: syslog, url: tls://syslog.example.com:1234 }
	//
	// TLS with server authentication is enabled by the URL scheme, for
	// example 'tls' or 'https'.  See `secret` for TLS client authentication.
	//
	// +optional
	URL string `json:"url"`

	// Secret for secure communication.
	// Secrets must be stored in the namespace containing the cluster logging operator.
	//
	// Client-authenticated TLS is enabled if the secret contains keys `tls.crt`,
	// `tls.key` and `ca.crt`. Output types with password authentication will use
	// keys `password` and `username`, not the exposed 'username@password' part of
	// the `url`.
	//
	// +optional
	Secret *OutputSecretSpec `json:"secret,omitempty"`

	// Insecure must be true for intentionally insecure outputs.
	// Has no function other than a marker to help avoid configuration mistakes.
	//
	// +optional
	Insecure bool `json:"insecure,omitempty"`

	// TODO(alanconway) not yet supported.
	//
	// Reconnect configures how the output handles connection failures.
	// Auto-reconnect is enabled by default.
	//
	// +optional
	// Reconnect *Reconnect `json:"reconnect,omitempty"`
}

// SecretName is a secret reference containing name only, no namespace.
type OutputSecretSpec struct {
	// Name of a secret in the namespace configured for log forwarder secrets.
	//
	// +required
	Name string `json:"name"`
}

// +kubebuilder:validation:Enum=Unreliable;Retry
type Reliability string

const (
	// Unreliable may drop data after a reconnect (at-most-once).
	Unreliable Reliability = "Unreliable"

	// Resend "in doubt" data after a reconnect. May cause duplicates (at-least-once).
	// May enable buffering, blocking and/or acknowledgment features of the output type.
	Resend Reliability = "Resend"
)

// Reconnect configures reconnect behavior after a disconnect.
type Reconnect struct {
	// FirstDelayMilliseconds is the time to wait after a disconnect before
	// the first reconnect attempt. If reconnect fails, the delay is doubled
	// on each subsequent attempt. The default is determined by the output type.
	//
	// +optional
	FirstDelayMilliseconds int64 `json:"firstDelayMilliseconds,omitempty"`

	// MaxDelaySeconds is the maximum delay between failed re-connect
	// attempts, and also the maximum time to wait for an unresponsive
	// connection attempt. The default is determined by the output type.
	//
	// +optional
	MaxDelayMilliseconds int64 `json:"maxDelayMilliseconds,omitempty"`

	// Reliability policy for data delivery after a re-connect.  This is
	// simple short-hand for configuring the output to a given level of
	// reliability.  The exact meaning depends on the output `type`.  The
	// default is determined by the output type.
	//
	// +optional
	Reliability Reliability `json:"reliability,omitempty"`
}

// OutputTypeSpec is a union of optional additional configuration for the output type.
type OutputTypeSpec struct {
	// +optional
	Syslog *outputs.Syslog `json:"syslog,omitempty"`
	// +optional
	FluentForward *outputs.FluentForward `json:"fluentForward,omitempty"`
	// +optional
	ElasticSearch *outputs.ElasticSearch `json:"elasticsearch,omitempty"`
}

// OutputTypeHandler has methods for each of the valid output types.
// They receive the output type spec field (possibly nil) and
// return a validation error.
//
type OutputTypeHandler interface {
	Syslog(*outputs.Syslog) error
	FluentForward(*outputs.FluentForward) error
	ElasticSearch(*outputs.ElasticSearch) error
}

// HandleType validates spec.Type and spec.OutputType,
// then calls the relevant handler method with the OutputType
// pointer, which may be nil.
//
func (spec OutputSpec) HandleType(h OutputTypeHandler) error {
	if !IsOutputTypeName(spec.Type) {
		return fmt.Errorf("not a valid output type: '%s'", spec.Type)
	}
	// Call handler method with OutputSpec field value
	goName := goNames[spec.Type]
	args := []reflect.Value{reflect.ValueOf(spec).FieldByName(goName)}
	result := reflect.ValueOf(h).MethodByName(goName).Call(args)[0].Interface()
	err, _ := result.(error)
	return err
}

var goNames = map[string]string{}

func init() {
	otsType := reflect.TypeOf(OutputTypeSpec{})
	for i := 0; i < otsType.NumField(); i++ {
		f := otsType.Field(i)
		jsonName := jsonTag(f)
		goNames[jsonName] = f.Name
	}
}

func jsonTag(field reflect.StructField) string {
	tag := field.Tag.Get("json")
	if j := strings.Index(tag, ","); j != -1 {
		tag = tag[:j]
	}
	return tag
}

// Output type and name constants.
const (
	OutputTypeElasticsearch = "elasticsearch"
	OutputTypeFluentForward = "fluentForward"
	OutputTypeSyslog        = "syslog"

	OutputNameDefault = "default" // Default log store.
)
