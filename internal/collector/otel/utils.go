package otel

import (
	"path"
)

const (
	ConfigFile      = "config.yaml"
	DefaultDataPath = "/var/lib/otelcol"
	configPath      = "/etc/otelcol"
)

func GetDataPath(namespace, forwarderName string) string {
	return path.Join(DefaultDataPath, namespace, forwarderName)
}
