package relayr

import "reflect"

// Relay encapsulates a connection with a client
// during an interaction with the server. It provides methods
// for interacting with clients and groups.
type Relay struct {
	Name         string            // The name of the relay it is associated with
	ConnectionID string            // The connectionID of the client that this Relay interacts with
	Clients      *ClientOperations // An abstraction over clients currently connected to this Relay

	methods  []string
	t        reflect.Type
	exchange *Exchange
}

// Call will execute a function on another server-side Relay,
// passing along the details of the currently connected client.
func (r *Relay) Call(fn string, args ...interface{}) {
	r.exchange.callRelayMethod(r, fn, args...)
}

// Groups returns a GroupOperations object, which offers helper
// methods for communicating with and grouping clients.
func (r *Relay) Groups(group string) *GroupOperations {
	return &GroupOperations{
		group: group,
		e:     r.exchange,
		relay: r,
	}
}
