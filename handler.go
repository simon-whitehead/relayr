package relayR

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type callInstruction struct {
	R string        // The Relay to use
	F string        // The function to call
	A []interface{} // The arguments to use
	C string        // ConnectionId
}

type negotiation struct {
	T string
}

type negotiationResponse struct {
	ConnectionID string
}

// Handle specifies the route to server the RelayR client script from.
func Handle(route string, e *Exchange) {
	setupMainRoute(route, e)
	setupNegotiateRoute(route, e)
	setupWebSocketRoute(route, e)
}

func setupMainRoute(route string, e *Exchange) {
	http.HandleFunc(route, func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			jsonResponse(w)
			writeRelayScript(e, w, route)
		} else {
			// Decode the post request
			decoder := json.NewDecoder(r.Body)
			var call callInstruction

			decoder.Decode(&call)

			relay := e.getRelayByName(call.R)
			err := e.callMethod(relay.t, call.F, call.C, call.A)

			// if there was a problem .. report it back
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
	})
}

func setupNegotiateRoute(route string, e *Exchange) {
	http.HandleFunc(route+"/negotiate", func(w http.ResponseWriter, r *http.Request) {
		jsonResponse(w)
		decoder := json.NewDecoder(r.Body)

		var neg negotiation

		decoder.Decode(&neg)

		encoder := json.NewEncoder(w)
		encoder.Encode(negotiationResponse{ConnectionID: e.addClient(neg.T)})
	})
}

func setupWebSocketRoute(route string, e *Exchange) {
	http.Handle(route+"/ws", socketHandler{e: e, c: e.circuits["WebSocket"].(*webSocketCircuit)})
}

// Write the client side script to the response
func writeRelayScript(e *Exchange, w http.ResponseWriter, route string) {
	buff := bytes.Buffer{}

	buff.WriteString(fmt.Sprintf(connectionClassScript, route))

	buff.WriteString(relayClassBegin)

	for _, relay := range e.Relays {
		buff.WriteString(fmt.Sprintf(relayBegin, relay.Name))

		for _, method := range relay.methods {
			buff.WriteString(fmt.Sprintf(relayMethod, lowerFirst(method), relay.Name, method))

		}
		buff.WriteString(relayEnd)
	}

	buff.WriteString(relayClassEnd)
	io.Copy(w, &buff)
}
