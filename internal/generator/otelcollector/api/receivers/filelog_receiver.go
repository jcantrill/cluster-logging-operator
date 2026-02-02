package receivers

import (
	"fmt"

	"github.com/openshift/cluster-logging-operator/internal/generator/otelcollector/api/operators"
)

type FileLog struct {
	id              string
	Include         []string             `json:"include,omitempty"`
	Exclude         []string             `json:"exclude,omitempty"`
	StartAt         StartAt              `json:"start_at,omitempty"`
	IncludeFilePath bool                 `json:"include_file_path,omitempty"`
	IncludeFileName bool                 `json:"include_file_name,omitempty"`
	Operators       []operators.Operator `json:"operators,omitempty"`
	//TODO Impl me. This is the pos file
	Storage string `json:"storage,omitempty"`
}

func NewFileLog(name string, init ...func(*FileLog)) *FileLog {
	typeName := "filelog"
	if name != "" {
		typeName = fmt.Sprintf("%s/%s", typeName, name)
	}
	f := &FileLog{
		id: typeName,
	}
	if init != nil {
		for _, i := range init {
			i(f)
		}
	}
	return f
}

func (f FileLog) GetTypeName() string {
	return f.id
}

type StartAt string

const (
	StartAtEnd StartAt = "end"
)
