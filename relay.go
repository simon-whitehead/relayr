package relayR

import "reflect"

// Relay represents a type that the runtime is aware of
type Relay struct {
	Name string // The name of the relay

	methods      []string // A list of known methods for this Relay
	ConnectionID string   // Unique identifier given to a connection

	t reflect.Type

	Clients *clientOperations
}
