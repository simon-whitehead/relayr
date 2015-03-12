package relayr

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strings"

	"github.com/gorilla/websocket"
)

var ClientScriptFunc func([]byte) []byte

var upgrader = &websocket.Upgrader{ReadBufferSize: 1024, WriteBufferSize: 1024}

// Exchange represents a hub that lets clients exchange information
// via Relays. Relays registered with the Exchange provide methods
// that can be invoked by clients.
type Exchange struct {
	relays     []Relay
	groups     map[string][]*Client
	transports map[string]Transport
}

type negotiation struct {
	t string // the transport that the client is comfortable using (e.g, websockets)
}

type negotiationResponse struct {
	ConnectionID string
}

// NewExchange initializes and returns a new *Exchange
func NewExchange() *Exchange {
	e := &Exchange{}
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
	encoder.Encode(negotiationResponse{ConnectionID: e.addClient(neg.t)})
}

func (e *Exchange) addClient(t string) string {
	cID := generateConnectionID()
	e.groups["Global"] = append(e.groups["Global"], &Client{ConnectionID: cID, exchange: e, transport: e.transports[t]})
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
			return &r
		}
	}

	return nil
}

func (e *Exchange) callRelayMethod(relay *Relay, fn string, args ...interface{}) error {
	newInstance := reflect.New(relay.t)
	method := newInstance.MethodByName(fn)
	empty := reflect.Value{}
	if method == empty {
		return errors.New(fmt.Sprintf("Method '%v' does not exist on relay '%v'", fn, relay.Name))
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

func (e *Exchange) Relay(x interface{}) *Relay {
	return e.getRelayByName(reflect.TypeOf(x).Name(), generateConnectionID())
}
