package main

import (
	"net/http"
	"time"

	"github.com/simon-whitehead/relayR"
)

type TimeRelay struct {
}

func (tr TimeRelay) PushTime(relay *relayr.Relay) {
	relay.Clients.All("timeUpdated", time.Now().Local().Format("Mon Jan 2 2006 03:04:05 PM"))
}

func main() {
	exchange := relayr.NewExchange()
	exchange.RegisterRelay(TimeRelay{})

	go func() {
		for {
			select {
			case <-time.After(time.Second * 1):
				exchange.Relay(TimeRelay{}).Call("PushTime")
			}
		}
	}()

	http.Handle("/relayr/", exchange)
	http.Handle("/", http.FileServer(http.Dir(".")))
	http.ListenAndServe(":8080", nil)
}
