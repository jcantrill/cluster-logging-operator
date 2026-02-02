package exporters

import (
	"github.com/openshift/cluster-logging-operator/internal/generator/otelcollector/api/helpers"
	"github.com/openshift/cluster-logging-operator/internal/generator/otelcollector/api/transport"
)

type CompressionType string

const (
	CompressionTypeGzip CompressionType = "gzip"
	CompressionTypeNone CompressionType = "none"
)

type EncodingType string

const (
	EncodingTypeJson  EncodingType = "json"
	EncodingTypeProto EncodingType = "proto"
)

type OtlpHttp struct {
	typeName        string
	Endpoint        string          `json:"endpoint"`
	EncodingType    EncodingType    `json:"encoding,omitempty"`
	CompressionType CompressionType `json:"compression,omitempty"`
	TLS             *transport.TLS  `json:"tls,omitempty"`
	RetryOnFailure  *Retry          `json:"retry_on_failure,omitempty"`
}

func NewOtlpHttp(name, endpoint string) *OtlpHttp {
	return &OtlpHttp{
		typeName: helpers.MakeTypeName("otlp_http", name),
		Endpoint: endpoint,
	}

}

func (o *OtlpHttp) GetTypeName() string {
	return o.typeName
}
