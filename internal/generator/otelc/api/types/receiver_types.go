package types

type ReceiverType string

const (
	ReceiverTypeFileLog ReceiverType = "file_log"
)

// Receiver is an OpenTelemetry Collector receiver for signals coming into the collector
type Receiver interface {
	ReceiverType() ReceiverType
}
