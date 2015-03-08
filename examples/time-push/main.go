package main

import (
	"net/http"
	"time"

	"github.com/simon-whitehead/relayR"
)

type TimeRelay struct {
	relayR.Relay
}

func (r TimeRelay) PushTimeToClient() {
	t := time.Now().Local()
	r.Clients.All("timeReceive", t.Format("Mon Jan 2 2006 03:04:05 PM"))
}

func main() {
	exchange := relayR.NewExchange()
	exchange.RegisterRelay(TimeRelay{})

	go func() {
		for {
			select {
			case <-time.After(time.Second * 1):
				exchange.Relay(TimeRelay{}).Call("PushTimeToClient")
			}
		}
	}()

	relayR.Handle("/~relayr", exchange)

	http.Handle("/", http.FileServer(http.Dir(".")))

	http.ListenAndServe(":8080", nil)
}
