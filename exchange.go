package relayR

import (
	"errors"
	"reflect"
)

// Exchange is the root object required to interact with
// the RelayR package
type Exchange struct {
	Relays   map[reflect.Type]Relay
	circuits map[string]circuit
	groups   map[string][]*client
}

// NewExchange initializes and returns an *Exchange
func NewExchange() *Exchange {
	e := &Exchange{
		Relays: map[reflect.Type]Relay{},
		groups: map[string][]*client{"Global": make([]*client, 0)},
	}
	e.circuits = map[string]circuit{"WebSocket": newWebSocketCircuit(e)}

	return e
}

// RegisterRelay registers a struct with an Exchange. If the struct does
// not embed relayR.Relay, it will panic.
func (e *Exchange) RegisterRelay(relay interface{}) {
	t := reflect.TypeOf(relay)
	if _, ok := t.FieldByName("Relay"); ok {
		name := t.Name()
		methods := e.getMethods(t)
		e.Relays[t] = Relay{Name: name, methods: methods, t: t}
	} else {
		panic(errors.New("ERR: Relay must embed relayR.Relay"))
	}
}

func (e *Exchange) getMethods(t reflect.Type) []string {
	r := []string{}
	for i := 0; i < t.NumMethod(); i++ {
		r = append(r, t.Method(i).Name)
	}

	return r
}

func (e *Exchange) callMethod(r reflect.Type, m, id string, a ...interface{}) error {
	relay, ok := e.Relays[r]
	if !ok {
		return errors.New("Unknown Relay: " + r.Name())
	}
	m = upperFirst(m)
	var args []reflect.Value
	if len(a) > 0 {
		args = buildArgValues(a...)
	}

	instance := e.createRelayInstance(relay, id)
	instanceMethod := instance.Elem().MethodByName(m)

	instanceMethod.Call(args)

	return nil
}

func (e *Exchange) createRelayInstance(relay Relay, id string) reflect.Value {
	instance := reflect.New(relay.t)
	instanceRelay := reflect.ValueOf(instance.Interface()).Elem().FieldByName("Relay").Interface().(Relay)
	instanceRelay.Name = relay.Name
	instanceRelay.Clients = &clientOperations{e: e, r: &instanceRelay}
	instanceRelay.ConnectionID = id
	instance.Elem().FieldByName("Relay").Set(reflect.ValueOf(instanceRelay))

	return instance
}

func buildArgValues(a ...interface{}) []reflect.Value {
	var r []reflect.Value

	for _, arg := range a {
		r = append(r, reflect.ValueOf(arg))
	}

	return r
}

// Add a new client to the required Circuit and return
// their generated Connection Id
func (e *Exchange) addClient(t string) string {
	cID := generateConnectionID()
	e.groups["Global"] = append(e.groups["Global"], &client{connectionID: cID, t: t, c: e.circuits[t]})
	return cID
}

func (e *Exchange) relayAll(r Relay, m string, a ...interface{}) {
	for _, c := range e.groups["Global"] {
		c.c.Send(clientMessage{R: r.Name, M: m, A: a, C: c.connectionID})
	}
}

func (e *Exchange) relayOthers(r Relay, m string, a ...interface{}) {
	for _, c := range e.groups["Global"] {
		if c.connectionID != r.ConnectionID {
			c.c.Send(clientMessage{S: false, R: r.Name, M: m, A: a, C: c.connectionID})
		}
	}
}

func (e *Exchange) relayGroup(r Relay, g, m string, a ...interface{}) {
	for _, c := range e.groups[g] {
		c.c.Send(clientMessage{S: false, R: r.Name, M: m, A: a, C: c.connectionID})
	}
}

func (e *Exchange) addToGroup(g string, r Relay) {
	e.groups[g] = append(e.groups[g], e.getClientByConnectionID(r.ConnectionID))
}

func (e *Exchange) removeFromGroup(g string, r Relay) {
	if i := e.getClientIndexInGroup(g, r.ConnectionID); i > -1 {
		group := e.groups[g]
		group[i] = nil
		e.groups[g] = append(group[:i], group[i+1:]...)
	}
}

func (e *Exchange) removeFromGroupByID(g, id string) {
	if i := e.getClientIndexInGroup(g, id); i > -1 {
		group := e.groups[g]
		group[i] = nil
		e.groups[g] = append(group[:i], group[i+1:]...)
	}
}

func (e *Exchange) removeFromAllGroups(id string) {
	for k := range e.groups {
		e.removeFromGroupByID(k, id)
	}
}

func (e *Exchange) getClientIndexInGroup(g, id string) int {
	for i, c := range e.groups[g] {
		if c.connectionID == id {
			return i
		}
	}

	return -1
}

func (e *Exchange) getClientByConnectionID(id string) *client {
	for _, c := range e.groups["Global"] {
		if c.connectionID == id {
			return c
		}
	}

	return nil
}

func (e *Exchange) getRelayByName(n string) Relay {
	for _, r := range e.Relays {
		if r.Name == n {
			return r
		}
	}

	return Relay{}
}

// Relay allows operations to be performed on a registered Relay
func (e *Exchange) Relay(v interface{}) relayOperations {
	return relayOperations{
		t: reflect.TypeOf(v),
		e: e,
	}
}
