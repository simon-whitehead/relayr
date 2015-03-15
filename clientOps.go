package relayr

// ClientOperations provides helper methods for
// interacting with Clients connected to a Relay.
type ClientOperations struct {
	e     *Exchange
	relay *Relay
}

// All invokes a client side method on all clients for the
// given relay.
func (c *ClientOperations) All(fn string, args ...interface{}) {
	c.e.callGroupMethod(c.relay, "Global", fn, args...)
}

// Others invokes a client side method on all clients except the
// one who calls it.
func (c *ClientOperations) Others(fn string, args ...interface{}) {
	c.e.callGroupMethodExcept(c.relay, "Global", fn, args...)
}
