package helpers

import "strings"

// InputComponent is a vector sink, transformation, source that is
// provided as input to other components
type InputComponent interface {
	// InputIDs are the ids of config elements to use as input to other components
	InputIDs() []string
}

// ComponentReceiver is a vector component that receives input from another component (e.g. transform, sink)
type ComponentReceiver interface {
	AddInputFrom(n InputComponent)
}

// MakeRouteInputID appends sourceType to rerouteId for input ids
func MakeRouteInputID(rerouteId, sourceType string) string {
	return strings.ToLower(strings.Join([]string{rerouteId, sourceType}, "."))
}
