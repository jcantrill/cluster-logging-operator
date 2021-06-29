package fluentd

import "fmt"

type Comment string

//Set - No-op for Comment
func (c Comment) Set(key string, value interface{}) {
}

//AsList returns the comment as a ruby comment in a list
func (c Comment) AsList() []string {
	return []string{fmt.Sprintf("# %v", c)}
}
