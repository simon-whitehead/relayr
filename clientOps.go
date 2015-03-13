package relayr

// ClientOperations provides helper methods for
// interacting with Clients connected to a Relay.
type ClientOperations struct {
	e     *Exchange
	relay *Relay
}

func (c *ClientOperations) All(fn string, args ...interface{}) {
	c.e.callGroupMethod(c.relay, "Global", fn, args...)
}

func (c *ClientOperations) Others(fn string, args ...interface{}) {
	c.e.callGroupMethodExcept(c.relay, "Global", fn, args...)
}
