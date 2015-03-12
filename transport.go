package relayr

// Transport represents a communication mechanism between
// the Exchange and the client.
type Transport interface {
	CallClientFunction(relay *Relay, fn string, args ...interface{})
}
