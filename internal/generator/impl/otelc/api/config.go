package api

// Config represents a configuration for the OpenTelementryContrib collector
type Config struct {
	Receivers Receivers `yaml:"receivers,omitempty"`

	Exporters Exporters `yaml:"exporters,omitempty"`

	Service Service `yaml:"service,omitempty"`
}

type Service struct {
	Pipelines Pipelines `yaml:"pipelines,omitempty"`
}

func (s Service) AddPipeline(pipeline *Pipeline) {
	s.Pipelines.Add(pipeline)
}

func NewConfig() *Config {
	c := &Config{
		Receivers: make(Receivers),
		Exporters: make(Exporters),
		Service: Service{
			Pipelines: make(Pipelines),
		},
	}
	return c
}

func (c *Config) AddReceivers(receivers Receivers) {
	for id, s := range receivers {
		c.Receivers[id] = s
	}
}

func (c *Config) AddExporters(exporters Exporters) {
	for id, s := range exporters {
		c.Exporters[id] = s
	}
}
