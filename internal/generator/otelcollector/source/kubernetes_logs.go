package source

import (
	"github.com/openshift/cluster-logging-operator/internal/generator/otelcollector/api/operators"
	"github.com/openshift/cluster-logging-operator/internal/generator/otelcollector/api/receivers"
)

type KubernetesLogs struct {
	receivers.FileLog
}

func NewKubernetesLogs(id string, includes, excludes []string, maxMergedLineBytes int64) KubernetesLogs {
	logs := KubernetesLogs{
		FileLog: *receivers.NewFileLog(id, func(f *receivers.FileLog) {
			f.Include = includes
			f.Exclude = excludes
			f.StartAt = receivers.StartAtEnd
			f.IncludeFilePath = true
			f.IncludeFileName = false
			f.Operators = []operators.Operator{
				operators.NewRegexParser("", `^(?P<time>[^ Z]+) (?P<stream>stdout|stderr) (?P<logtag>[^ ]*) ?(?P<log>.*)$`,
					func(p *operators.RegexParser) {
						p.Timestamp = &operators.Timestamp{
							ParseFrom:  "attributes.time",
							LayoutType: operators.TimestampLayoutTypeGoTime,
							Layout:     "2006-01-02T15:04:05.999999999Z07:00",
						}
					}),
				operators.NewRegexParser("", `^.*\/(?P<namespace>[^_]+)_(?P<pod_name>[^_]+)_(?P<uid>[a-f0-9\-]{36})\/(?P<container_name>[^\._]+)\/(?P<restart_count>\d+)\.log$`,
					func(p *operators.RegexParser) {
						p.ParseFrom = `attributes["log.file.path"]`
						p.Cache = &operators.Cache{
							Size: 128,
						}
					}),
				operators.NewMove("attributes.log", `body`),
				operators.NewMove("attributes.stream", `attributes["log.iostream"]`),
				operators.NewMove("attributes.container_name", `attributes["k8s.container.name"]`),
				operators.NewMove("attributes.namespace", `attributes["k8s.namespace.name"]`),
				operators.NewMove("attributes.pod_name", `attributes["k8s.pod.name"]`),
				operators.NewMove("attributes.uid", `attributes["k8s.pod.uid"]`),
			}
		}),
	}
	return logs
}
