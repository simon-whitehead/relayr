package relayr

// GroupOperations provides helper methods for communicating
// with clients in groups. Clients must be added to a group
// to be considered a member of a group.
type GroupOperations struct {
	relay *Relay
	group string
	e     *Exchange
}

// Add adds a client to a group via its ConnectionID. It
// is a member of the group for the remainder of its
// connection, until disconnection. At that point, the
// client must re-negotiate its place within the group
// to be considered a member of it.
func (g *GroupOperations) Add(connectionID string) {
	g.e.addToGroup(g.group, connectionID)
}

// Remove removes a client from a group via its ConnectionID.
func (g *GroupOperations) Remove(connectionID string) {
	g.e.removeFromGroupByID(g.group, connectionID)
}

func (g *GroupOperations) Call(fn string, args ...interface{}) {
	g.e.callGroupMethod(g.relay, g.group, fn, args...)
}
