package helpers

import (
	"fmt"
	"strings"
)

// MakeTypeName formats the typeName of the component to the OTLP spec based upon the
// componentType and a name (e.g. logs/foo_var).  Panics if componentType is empty.
func MakeTypeName(componentType string, names ...string) string {
	if strings.TrimSpace(componentType) == "" {
		panic("componentType must not be empty")
	}
	if names != nil {
		name := strings.Join(names, "_")
		if name != "" {
			componentType = fmt.Sprintf("%s/%s", componentType, name)
		}
	}
	return componentType
}
