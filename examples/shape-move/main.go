package main

import (
	"net/http"

	"github.com/simon-whitehead/relayR"
)

type ShapeRelay struct {
	relayR.Relay
}

func (r ShapeRelay) UpdateShape(s map[string]interface{}) {
	r.Clients.Others("shapeUpdated", s) // Only send to other clients
}

func main() {
	exchange := relayR.NewExchange()
	exchange.RegisterRelay(ShapeRelay{})

	relayR.Handle("/~relay", exchange)

	http.Handle("/", http.FileServer(http.Dir(".")))

	http.ListenAndServe(":8080", nil)
}
