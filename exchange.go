package relayr

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strings"

	"github.com/gorilla/websocket"
)

// ClientScriptFunc is a callback for altering the client side
// generated Javascript. This can be used to minify/alter the
// generated RelayR library.
var ClientScriptFunc func([]byte) []byte

var upgrader = &websocket.Upgrader{ReadBufferSize: 1024, WriteBufferSize: 1024}

// Exchange represents a hub where clients exchange information
// via Relays. Relays registered with the Exchange expose methods
// that can be invoked by clients.
type Exchange struct {
	relays     []Relay
	groups     map[string][]*client
	transports map[string]Transport
}

type negotiation struct {
	T string // the transport that the client is comfortable using (e.g, websockets)
}

type negotiationResponse struct {
	ConnectionID string
}

// NewExchange initializes and returns a new Exchange
func NewExchange() *Exchange {
	e := &Exchange{}
	e.groups = make(map[string][]*client)
	e.transports = map[string]Transport{"websocket": newWebSocketTransport(e)}

	return e
}

func (e *Exchange) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	route := extractRouteFromURL(r)
	op := extractOperationFromURL(r)

	switch op {
	case OpWebSocket:
		e.upgradeWebSocket(w, r)
	case OpNegotiate:
		e.negotiateConnection(w, r)
	default:
		e.writeClientScript(w, route)
	}
}

func extractRouteFromURL(r *http.Request) string {
	lastSlash := strings.LastIndex(r.URL.Path, "/")
	return r.URL.Path[:lastSlash]
}

func extractOperationFromURL(r *http.Request) string {
	lastSlash := strings.LastIndex(r.URL.Path, "/")
	return r.URL.Path[lastSlash+1:]
}

func (e *Exchange) upgradeWebSocket(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	c := &connection{e: e, out: make(chan []byte, 256), ws: ws, c: e.transports["websocket"].(*webSocketTransport), id: r.URL.Query()["connectionId"][0]}
	c.c.connected <- c
	defer func() { c.c.disconnected <- c }()
	go c.write()
	c.read()
}

func (e *Exchange) negotiateConnection(w http.ResponseWriter, r *http.Request) {
	jsonResponse(w)
	decoder := json.NewDecoder(r.Body)

	var neg negotiation

	decoder.Decode(&neg)

	encoder := json.NewEncoder(w)
	encoder.Encode(negotiationResponse{ConnectionID: e.addClient(neg.T)})
}

func (e *Exchange) addClient(t string) string {
	cID := generateConnectionID()
	e.groups["Global"] = append(e.groups["Global"], &client{ConnectionID: cID, exchange: e, transport: e.transports[t]})
	return cID
}

func (e *Exchange) writeClientScript(w http.ResponseWriter, route string) {
	resultBuff := bytes.Buffer{}
	buff := bytes.Buffer{}

	buff.WriteString(fmt.Sprintf(connectionClassScript, route))

	buff.WriteString(relayClassBegin)

	for _, relay := range e.relays {
		buff.WriteString(fmt.Sprintf(relayBegin, relay.Name))

		for _, method := range relay.methods {
			buff.WriteString(fmt.Sprintf(relayMethod, lowerFirst(method), relay.Name, method))

		}
		buff.WriteString(relayEnd)
	}

	buff.WriteString(relayClassEnd)

	if ClientScriptFunc != nil {
		resultBuff.Write(ClientScriptFunc(buff.Bytes()))
	} else {
		resultBuff.Write(buff.Bytes())
	}

	io.Copy(w, &resultBuff)
}

// RegisterRelay registers a struct as a Relay with the Exchange. This allows clients
// to invoke server methods on a Relay and allows the Exchange to invoke
// methods on a Relay on the server side.
func (e *Exchange) RegisterRelay(x interface{}) {
	t := reflect.TypeOf(x)

	methods := e.getMethodsForType(t)

	e.relays = append(e.relays, Relay{Name: t.Name(), t: t, methods: methods, exchange: e})
}

func (e *Exchange) getMethodsForType(t reflect.Type) []string {
	r := []string{}
	for i := 0; i < t.NumMethod(); i++ {
		r = append(r, t.Method(i).Name)
	}

	return r
}

func (e *Exchange) getRelayByName(name string, cID string) *Relay {
	// Create an instance of Relay
	for _, r := range e.relays {
		if r.Name == name {
			relay := &Relay{
				Name:         name,
				ConnectionID: cID,
				t:            r.t,
				exchange:     e,
			}

			relay.Clients = &ClientOperations{
				e:     e,
				relay: relay,
			}

			return relay
		}
	}

	return nil
}

func (e *Exchange) callRelayMethod(relay *Relay, fn string, args ...interface{}) error {
	newInstance := reflect.New(relay.t)
	method := newInstance.MethodByName(fn)
	empty := reflect.Value{}
	if method == empty {
		return fmt.Errorf("Method '%v' does not exist on relay '%v'", fn, relay.Name)
	}
	method.Call(buildArgValues(relay, args...))

	return nil
}

func buildArgValues(relay *Relay, args ...interface{}) []reflect.Value {
	r := []reflect.Value{reflect.ValueOf(relay)}
	for _, a := range args {
		r = append(r, reflect.ValueOf(a))
	}

	return r
}

// Relay generates an instance of a Relay, allowing calls to be made to
// it on the server side. It is generated a random ConnectionID for the duration
// of the call and it does not represent an actual client.
func (e *Exchange) Relay(x interface{}) *Relay {
	return e.getRelayByName(reflect.TypeOf(x).Name(), generateConnectionID())
}

func (e *Exchange) callClientMethod(r *Relay, fn string, args ...interface{}) {
	if r.ConnectionID == "" {
		e.callGroupMethod(r, "Global", fn, args...)
		return
	}

	c := e.getClientByConnectionID(r.ConnectionID)
	if c != nil {
		c.transport.CallClientFunction(r, fn, args...)
	}
}

func (e *Exchange) callGroupMethod(relay *Relay, group, fn string, args ...interface{}) {
	if _, ok := e.groups[group]; ok {
		for _, c := range e.groups[group] {
			r := e.getRelayByName(relay.Name, c.ConnectionID)
			c.transport.CallClientFunction(r, fn, args...)
		}
	}
}

func (e *Exchange) callGroupMethodExcept(relay *Relay, group, fn string, args ...interface{}) {
	for _, c := range e.groups[group] {
		if c.ConnectionID == relay.ConnectionID {
			continue
		}
		r := e.getRelayByName(relay.Name, c.ConnectionID)
		c.transport.CallClientFunction(r, fn, args...)
	}
}

func (e *Exchange) getClientByConnectionID(cID string) *client {
	for _, c := range e.groups["Global"] {
		if c.ConnectionID == cID {
			return c
		}
	}
	return nil
}

func (e *Exchange) removeFromAllGroups(id string) {
	for group := range e.groups {
		e.removeFromGroupByID(group, id)
	}
}

func (e *Exchange) removeFromGroupByID(g, id string) {
	if i := e.getClientIndexInGroup(g, id); i > -1 {
		group := e.groups[g]
		group[i] = nil
		e.groups[g] = append(group[:i], group[i+1:]...)
	}
}

func (e *Exchange) getClientIndexInGroup(g, id string) int {
	for i, c := range e.groups[g] {
		if c.ConnectionID == id {
			return i
		}
	}

	return -1
}

func (e *Exchange) addToGroup(group, connectionID string) {
	// only add them if they aren't currently in the group
	if e.getClientIndexInGroup(group, connectionID) == -1 {
		e.groups[group] = append(e.groups[group], e.getClientByConnectionID(connectionID))
	}
}
