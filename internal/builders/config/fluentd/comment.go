package fluentd

import "fmt"

type Comment string

//AsList returns the comment as a ruby comment in a list
func (c Comment) AsList() []string {
	return []string{fmt.Sprintf("# %v", c)}
}
