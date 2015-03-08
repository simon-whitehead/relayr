package relayR

type clientOperations struct {
	e *Exchange
	r *Relay
}

func (c *clientOperations) All(m string, a ...interface{}) {
	c.e.relayAll(*c.r, m, a)
}

func (c *clientOperations) Others(m string, a ...interface{}) {
	c.e.relayOthers(*c.r, m, a)
}

func (c *clientOperations) Group(g, m string, a ...interface{}) {
	c.e.relayGroup(*c.r, g, m, a)
}

func (c *clientOperations) AddToGroup(g string) {
	c.e.addToGroup(g, *c.r)
}

func (c *clientOperations) RemoveFromGroup(g string) {
	c.e.removeFromGroup(g, *c.r)
}

type client struct {
	connectionID string
	t            string
	c            circuit
}

// clientMessage represents a message from a Circuit
// to a client
type clientMessage struct {
	C string        // ConnectionId
	R string        // Relay to call
	M string        // Method to call
	A []interface{} // arguments to send along
	S bool          // Server or client?
}
