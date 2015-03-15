package main

import (
	"net/http"

	"github.com/simon-whitehead/relayR"
)

type ShapeRelay struct {
}

func (sr ShapeRelay) UpdateShape(relay *relayr.Relay, s map[string]interface{}) {
	relay.Clients.Others("shapeUpdated", s) // Only send to other clients
}

func main() {
	exchange := relayr.NewExchange()
	exchange.RegisterRelay(ShapeRelay{})

	http.Handle("/relayr/", exchange)
	http.Handle("/", http.FileServer(http.Dir(".")))

	http.ListenAndServe(":8080", nil)
}
