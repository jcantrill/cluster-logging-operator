package yaml

import "sigs.k8s.io/yaml"

// Unmarshal JSON or YAML string into a value according to k8s rules.
// Uses sigs.k8s.io/yaml.
func Unmarshal(s string, v interface{}) error { return yaml.Unmarshal([]byte(s), v) }
