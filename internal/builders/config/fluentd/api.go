package fluentd

import "fmt"

type Set []string

func NewSet(entries ...string) *Set {
	set := Set{}
	set.Add(entries)
	return &set
}
func (s *Set) Entries() []string {
	return []string(*s)
}

func(c *Set) Add(keys []string) {
	for _, k := range keys {
		for _, e := range *c {
			if e == k {
				break
			}
		}
		*c = append(*c, k)
	}
}

type SerializableToStringList interface {
	AsList() []string
}

type Directive interface {
	SerializableToStringList
	Set(key string, value interface{})
	SetAll(configs map[string]interface{})
}

type Configuration struct {
	AllowedKeys *Set
	Config      map[string]interface{}
	Type        string
}

func Match(pattern string, config Configuration) []string {
	buf := []string{fmt.Sprintf("<match %s>",pattern)}
	buf = append(buf, BuildBlock(config)...)
	return append(buf, "</match>")
}

func Label(name string, config []string) []string {
	buf := []string{fmt.Sprintf("<label %s>", name)}
	buf = append(buf, config...)
	return append(buf, "</label>")
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
