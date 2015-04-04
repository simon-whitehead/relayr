package relayr

// Transport represents a communication mechanism between
// a Relay and a client.
type Transport interface {
	CallClientFunction(relay *Relay, cid, fn string, args ...interface{})
}
