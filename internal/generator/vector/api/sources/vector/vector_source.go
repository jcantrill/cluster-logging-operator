package vector

import (
	vectorapi "github.com/openshift/cluster-logging-operator/internal/generator/vector/api"
)

const (
	DefaultListenPort = 6000
)

// Source is a source for receiving logs from another vector instance
type Source struct {
	// id is temporary for building using the old generators
	id string

	// Type is required to be 'vector'
	Type string `json:"type" yaml:"type" toml:"type"`

	// Address on which to receive logs
	Address string `json:"address" yaml:"address" toml:"address"`
}

func New(id, address string) *Source {
	return &Source{
		id:      id,
		Type:    vectorapi.ComponentTypeVector,
		Address: address,
	}
}

func (s *Source) Kind() string {
	return vectorapi.ComponentKindSource
}

func (s *Source) String() string {
	c := vectorapi.Config{
		Sources: map[string]vectorapi.SourceKind{
			s.id: s,
		},
	}
	return c.String()
}

// Name is deprecated function to adapt to existing generator framework
func (s *Source) Name() string {
	return s.Kind()
}

// Template is a deprecated method to adapt to the existing generator framework
func (s *Source) Template() string {
	return `{{define "` + s.Name() + `" -}}
{{ if ne "" .String }}
{{.}}
{{end}}
{{end}}`
}
