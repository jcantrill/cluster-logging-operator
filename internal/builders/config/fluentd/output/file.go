package output

import (
	"github.com/openshift/cluster-logging-operator/internal/builders/config/fluentd"
	"strings"
)

type MatchType string

const (
	MatchTypeFile = "file"
)

type OutFile struct {
	fluentd.Configuration
	Pattern string
}

func NewOutFileBuilder(pattern string, configType MatchType) *OutFile {
	return &OutFile{
		Pattern: pattern,
		Configuration: fluentd.Configuration{
			Type: "file",
			AllowedKeys: fluentd.NewSet(
				"path",
				"compress",
				"buffer",
			),
			Config: map[string]interface{}{},
		},
	}
}

func (b *OutFile) Set(key string, value interface{}) {
	b.Config[key] = value
}

func (b *OutFile) AsList() []string {
	return fluentd.Match(b.Pattern, b.Configuration)
}

func (b *OutFile) String() string {
	return strings.Join(b.AsList(), "\n")
}

func (b *OutFile) WithBuffer() *fluentd.Buffer {
	buffer := fluentd.NewBuffer("")
	b.Set("buffer", buffer)
	return buffer
}

func (b *OutFile) WithPath(path string) *OutFile {
	b.Set("path",path)
	return b
}
func (b *OutFile) WithCompress(compress string) *OutFile {
	b.Set("compress",compress)
	return b
}