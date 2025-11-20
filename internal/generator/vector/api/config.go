package api

import (
	"strings"

	"github.com/openshift/cluster-logging-operator/internal/utils/toml"
)

// Config represents a configuration for vector
type Config struct {

	// Sources is the set of transform ids to transform configurations
	Sources map[string]SourceKind `json:"sources" yaml:"sources" toml:"sources"`

	// Transforms is the set of transform ids to transform configurations
	Transforms map[string]interface{} `json:"transforms" yaml:"transforms" toml:"transforms"`
}

// Name is a deprecated method to adapt to the existing generator framework
func (c Config) Name() string {
	return "config"
}

// Template is a deprecated method to adapt to the existing generator framework
func (c Config) Template() string {
	return `{{define "` + c.Name() + `" -}}
{{ if ne "" .String }}
{{.}}
{{end}}
{{end}}`
}

func (c Config) String() string {
	config := strings.ReplaceAll(toml.MustMarshal(c), "[transforms]", "")
	config = strings.ReplaceAll(config, "[sinks]", "")
	return strings.ReplaceAll(config, "[sources]", "")
}

type ComponentKind interface {
	Kind() string
}

type SourceKind interface {
	ComponentKind
}
