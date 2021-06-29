package fluentd

import (
	"fmt"
	"strings"

)

type Filter struct {
	Configuration
	Tag string
}

func (b *Filter) Set(key string, value interface{}) {
	b.Config[key] = value
}

func (b *Filter) AsList() []string {
	buf := []string{fmt.Sprintf("<filter %s>", b.Tag)}
	buf = append(buf, BuildBlock(b.Configuration)...)
	return append(buf, "</filter>")
}

func (b *Filter) String() string {
	return strings.Join(b.AsList(), "\n")
}