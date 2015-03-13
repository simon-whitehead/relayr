package relayr

import "reflect"

// Relay encapsulates a connection with a client
// during an interaction with the server. It provides methods
// for interacting with Clients and Groups.
type Relay struct {
	Name         string // The name of the relay it is associated with
	ConnectionID string // The connectionID of the client that this Relay interacts with
	Clients      *ClientOperations

	methods  []string
	t        reflect.Type
	exchange *Exchange
}

// Call will execute a function on another server-side Relay,
// passing along the details of the currently connected Client.
func (r *Relay) Call(fn string, args ...interface{}) {
	r.exchange.callRelayMethod(r, fn, args...)
}
