package api

import (
	"fmt"
	"strings"

	"github.com/openshift/cluster-logging-operator/internal/generator/otelc/api/receivers"
	"github.com/openshift/cluster-logging-operator/internal/generator/otelc/api/types"
	"gopkg.in/yaml.v3"
)

// Receivers is a set of receivers by their id.
// In OpenTelemetry Collector configuration, receiver IDs follow the pattern "type" or "type/name"
// where the type determines the receiver implementation.
type Receivers map[string]types.Receiver

func (receiverMap *Receivers) Add(id string, receiver types.Receiver) {
	(*receiverMap)[id] = receiver
}

func (receiverMap *Receivers) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind != yaml.MappingNode {
		return fmt.Errorf("receivers data must be a mapping, got: %v", value.Kind)
	}

	if *receiverMap == nil {
		*receiverMap = make(Receivers)
	}

	// Parse the mapping as a map first to extract entries
	entries := make(map[string]*yaml.Node)
	for i := 0; i < len(value.Content); i += 2 {
		key := value.Content[i].Value
		val := value.Content[i+1]
		entries[key] = val
	}

	for id, entry := range entries {
		// Extract receiver type from the ID (e.g., "file_log" or "file_log/my-instance")
		receiverType := id
		if slashIdx := strings.Index(id, "/"); slashIdx != -1 {
			receiverType = id[:slashIdx]
		}

		var receiver types.Receiver
		switch types.ReceiverType(receiverType) {
		case types.ReceiverTypeFileLog:
			var r receivers.FileLog
			if err := entry.Decode(&r); err != nil {
				return fmt.Errorf("failed to unmarshal file_log receiver %s: %w", id, err)
			}
			receiver = &r
		default:
			return fmt.Errorf("unknown receiver type %s for receiver %s", receiverType, id)
		}

		(*receiverMap)[id] = receiver
	}

	return nil
}
