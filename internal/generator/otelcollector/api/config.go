package api

import (
	"github.com/openshift/cluster-logging-operator/internal/generator/otelcollector/api/helpers"
)

type Component interface {
	GetTypeName() string
}

type Config struct {
	Extensions ComponentList `json:"extensions,omitempty"`
	Exporters  ComponentList `json:"exporters,omitempty"`
	Receivers  ComponentList `json:"receivers,omitempty"`
	Service    *Service      `json:"service,omitempty"`
}

func NewConfig() *Config {
	return &Config{
		Extensions: make(ComponentList),
		Exporters:  make(ComponentList),
		Receivers:  make(ComponentList),
		Service: &Service{
			Pipelines: make(ComponentList),
		},
	}
}

func (c *Config) SetExtension(comp Component) *Config {
	c.Extensions.Set(comp)
	c.Service.Extensions = append(c.Service.Extensions, comp.GetTypeName())

	return c
}
func (c *Config) SetReceiver(receiver Component) *Config {
	c.Receivers.Set(receiver)
	return c
}
func (c *Config) SetExporter(exporter Component) *Config {
	c.Exporters.Set(exporter)
	return c
}

func (c *Config) SetPipeline(pipeline Component) *Config {
	c.Service.Pipelines.Set(pipeline)
	return c
}

type Service struct {
	Extensions []string      `json:"extensions,omitempty"`
	Pipelines  ComponentList `json:"pipelines"`
}

type ComponentList map[string]Component

func (cl ComponentList) Set(component Component) {
	cl[component.GetTypeName()] = component
}

type LogsPipeline struct {
	Pipeline
}

// NewLogsPipeline returns a LogsPipeline with the typename = 'logs'
// unless the name arg is not empty then typename = 'logs/name
func NewLogsPipeline(name string) *LogsPipeline {
	return &LogsPipeline{
		Pipeline: Pipeline{
			typeName: helpers.MakeTypeName("logs", name),
		},
	}
}

func (p *LogsPipeline) AddReciever(comp Component) *LogsPipeline {
	p.Receivers = append(p.Receivers, comp.GetTypeName())
	return p
}

func (p *LogsPipeline) AddExporter(comp Component) *LogsPipeline {
	p.Exporters = append(p.Exporters, comp.GetTypeName())
	return p
}

func (p *LogsPipeline) AddProcesor(comp Component) *LogsPipeline {
	p.Processors = append(p.Processors, comp.GetTypeName())
	return p
}

func (p *LogsPipeline) GetTypeName() string {
	return p.typeName
}

type Pipeline struct {
	typeName   string
	Receivers  []string `json:"receivers"`
	Exporters  []string `json:"exporters"`
	Processors []string `json:"processors,omitempty"`
}
