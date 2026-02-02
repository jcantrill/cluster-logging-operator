package extensions

import (
	"github.com/openshift/cluster-logging-operator/internal/generator/otelcollector/api/helpers"
)

type BearerTokenAuth struct {
	typeName string
	FileName string `json:"filename,omitempty"`
}

func NewBearerTokenAuth(name, fileName string) *BearerTokenAuth {
	return &BearerTokenAuth{
		typeName: helpers.MakeTypeName("bearertokenauth", name),
		FileName: fileName,
	}
}

func (b *BearerTokenAuth) GetTypeName() string {
	return b.typeName
}
