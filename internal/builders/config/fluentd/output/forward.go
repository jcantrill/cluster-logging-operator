package output

import "github.com/openshift/cluster-logging-operator/internal/builders/config/fluentd"

type OutForward struct {
	fluentd.Configuration
	Pattern string
	Servers ServerList
}

func (o *OutForward) AsList() []string {
	return fluentd.Match(o.Pattern, o.Configuration)
}

func NewOutForwardBuilder(pattern string) *OutForward {
	return &OutForward{
		Pattern: pattern,
		Configuration: fluentd.Configuration{
			Type: "forward",
			AllowedKeys: fluentd.NewSet(
				"@id",
				"heartbeat_type",
				"keepalive",
				"transport",
				"tls_verify_hostname",
				"tls_version",
				"tls_insecure_mode",
				"tls_client_private_key_path",
				"tls_client_cert_path",
				"tls_cert_path",
				"secret",
				"buffer",
				"server",
			),
			Config: map[string]interface{}{},
		},
		Servers: ServerList{},
	}
}

func (o *OutForward) WithBuffer(bufferType string) *fluentd.Buffer {
	buffer := fluentd.NewBuffer(bufferType)
	o.Config["buffer"] = buffer
	return buffer

}

type ServerBuilder struct {
	fluentd.Configuration
}
func (b *ServerBuilder) Set(key string, value interface{}) {
	b.Config[key] = value
}

func (b *ServerBuilder) WithName(name string) *ServerBuilder {
	b.Set("name", name)
	return b
}
func (b *ServerBuilder) WithPort(port int) *ServerBuilder{
	b.Set("port", port)
	return b
}
func (b *ServerBuilder) WithWeight(weight int) *ServerBuilder{
	b.Set("weight", weight)
	return b
}

type ServerList []*ServerBuilder
func(sl *ServerList) AsList() []string{
	config := []string{}
	servers := []*ServerBuilder(*sl)
	for _, s := range servers {
		config = append(config, fluentd.BuildBlock(s.Configuration)...)
	}
	return config
}
func (o *OutForward) AddServer(host string) *ServerBuilder {
	server := &ServerBuilder{
		Configuration: fluentd.Configuration{
			Config: map[string]interface{}{},
			AllowedKeys: fluentd.NewSet(
				"name",
				"host",
				"port",
				"shared_key",
				"username",
				"password",
				"standby",
				"weight",
			),
		},
	}
	o.Servers = append(o.Servers, server)
	return server
}
