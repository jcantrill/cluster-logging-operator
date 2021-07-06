package fluentd

import (
	"fmt"
	"strings"
)


var replacer = strings.NewReplacer(" ", "_", "-", "_", ".", "_")

func FormatLabelName(name string) string {
	return strings.ToUpper(fmt.Sprintf("@%s", replacer.Replace(name)))
}

