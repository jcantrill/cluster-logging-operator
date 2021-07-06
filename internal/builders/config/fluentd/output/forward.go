package output

import (
	"github.com/openshift/cluster-logging-operator/internal/builders/config/fluentd"
)

type OutForward struct {
	fluentd.Configuration
	Pattern string
}

func (o *OutForward) AsList() []string {
	return fluentd.Match(o.Pattern, o.Configuration)
}

func (o *OutForward) Set(key string, value interface{}){
	if value != nil {
		o.Config[key] = value
	}
}

func (o *OutForward) SetAll(configs map[string]interface{}) {
	for k, v := range configs {
		o.Set(k,v)
	}
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
				"security",
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
			Config: map[string]interface{}{
				"server": ServerList{},
			},
		},
	}
}


type ServerList []*Server
func(sl ServerList) AsList() []string{
	config := []string{}
	servers := []*Server(sl)
	for _, s := range servers {
		config = append(config, "<server>")
		config = append(config, fluentd.BuildBlock(s.Configuration)...)
		config = append(config, "</server>")
	}
	return config
}


func (o *OutForward) WithBuffer(bufferType string) *fluentd.Buffer {
	buffer := fluentd.NewBuffer(bufferType)
	o.Config["buffer"] = buffer
	return buffer

}

type Security struct {
	fluentd.Configuration
}
func NewSecurityBuilder() *Security{
	return &Security{
		Configuration: fluentd.Configuration{
			AllowedKeys: fluentd.NewSet(
				"self_hostname",
				"shared_key",
			),
			Config: map[string]interface{}{},
		},
	}
}
func(s *Security) Set(key string, value interface{}){
	if value != nil {
		s.Config[key] = value
	}
}

func(s *Security)AsList() []string{
	config := []string{"<security>"}
	config = append(config, fluentd.BuildBlock(s.Configuration)...)
	return append(config, "</security>")
}
func(s *Security) WithHostname(value string) *Security {
	s.Set("self_hostname", value)
	return s
}

func(s *Security) WithShardKey(value string) *Security {
	s.Set("shared_key", value)
	return s
}

type Server struct {
	fluentd.Configuration
}
func (b *Server) Set(key string, value interface{}) {
	b.Config[key] = value
}

func (b *Server) WithName(name string) *Server {
	b.Set("name", name)
	return b
}
func (b *Server) WithPort(port int) *Server {
	b.Set("port", port)
	return b
}
func (b *Server) WithWeight(weight int) *Server {
	b.Set("weight", weight)
	return b
}

func (o *OutForward) AddServer(host string) *Server {
	server := &Server{
		Configuration: fluentd.Configuration{
			Config: map[string]interface{}{
				"host": host,
			},
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
	servers := []*Server(o.Config["server"].(ServerList))
	servers = append(servers, server)
	o.Set("server", ServerList(servers))
	return server
}

func (o *OutForward) WithSecurity() *Security {
	security := NewSecurityBuilder()
	o.Set("security", security)
	return security
}

func (o *OutForward) WithHeartBeatType(value string) *OutForward {
	o.Set("heartbeat_type", value)
	return o
}

func (o *OutForward) WithKeepAlive(alive bool) *OutForward {
	o.Set("keepalive", alive)
	return o
}

