package fluentd

import "fmt"

type Set map[string]interface{}

func NewSet(entries ...string) Set {
	set := map[string]interface{}{}
	for _, k := range entries {
		set[k] = nil
	}
	return set
}
func (s *Set) Entries() []string {
	entries := []string{}
	for k, _ := range *s {
		entries = append(entries, k)
	}
	return entries
}
func(c Set) Add(keys []string) {
	set := map[string]interface{}(c)
	for _, k := range keys {
		set[k] = nil
	}
}

type SerializableToStringList interface {
	AsList() []string
}

type Directive interface {
	SerializableToStringList
	Set(key string, value interface{})
}

type Configuration struct {
	AllowedKeys Set
	Config      map[string]interface{}
	Type        string
}

func Match(pattern string, config Configuration) []string {
	buf := []string{fmt.Sprintf("<match %s>",pattern)}
	buf = append(buf, BuildBlock(config)...)
	return append(buf, "</match>")
}

func BuildBlock(config Configuration) []string {
	out := []string{}
	if config.Type != "" {
		out = append(out, fmt.Sprintf("@type %s", config.Type))
	}
	for _, key := range config.AllowedKeys.Entries() {
		if value, ok := config.Config[key]; ok {
			block, ok := value.(SerializableToStringList)
			if ok {
				out = append(out, block.AsList()...)
			} else {
				out = append(out, fmt.Sprintf("%v %v", key, value))
			}
		}
	}
	return out
}
