package fluentd

type FluentdConfig struct {
	Directives []SerializableToStringList
}

func NewConfigBuilder() *FluentdConfig {
	return &FluentdConfig{
		Directives: []SerializableToStringList{},
	}
}

func (c *FluentdConfig) AsList() []string {
	config := []string{}
	for _,d := range c.Directives{
		config = append(config, d.AsList()...)
	}
	return config
}

func (c *FluentdConfig) AddComment(comment string) *FluentdConfig {
	c.Directives = append(c.Directives, Comment(comment))
	return c
}
