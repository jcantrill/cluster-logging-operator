package source

import (
	"github.com/openshift/cluster-logging-operator/internal/builders/config/fluentd"
	"strings"
)

type Tail struct {
	outPrefix string
	fluentd.Configuration
}

func NewTailBuilder(path string) *Tail {
	return &Tail{
		Configuration: fluentd.Configuration{
			Type: "tail",
			AllowedKeys: fluentd.NewSet(
				"@id",
				"@label",
				"path",
				"pos_file",
				"tag",
				"parse",
				"exclude_path",
				"refresh_interval",
				"rotate_wait",
				"read_from_head",
			),
			Config: map[string]interface{}{
				"path": path,
			},
		},
		outPrefix: "\t",
	}
}

func (b *Tail) Set(key, value string) {
	b.Config[key] = value
}

func (b *Tail) AsList() []string {
	buf := []string{"<source>"}
	buf = append(buf, fluentd.BuildBlock(b.Configuration)...)
	return append(buf, "</source>")
}

func (b *Tail) String() string {
	return strings.Join(b.AsList(), "\n")
}

func (b *Tail) WithPath(value string) *Tail {
	b.Config["path"] = value
	return b
}
