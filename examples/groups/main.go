package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/simon-whitehead/relayR"
)

var group int = 1

type NotificationRelay struct {
}

func (nr NotificationRelay) SendNotification(relay *relayr.Relay, g string) {
	relay.Groups(g).Call("notificationReceive", "Notification for group "+g)
}

func (nr NotificationRelay) Subscribe(relay *relayr.Relay, g string) {
	relay.Groups(g).Add(relay.ConnectionID)
}

func (nr NotificationRelay) Unsubscribe(relay *relayr.Relay, g string) {
	relay.Groups(g).Remove(relay.ConnectionID)
}

func main() {
	exchange := relayr.NewExchange()
	exchange.RegisterRelay(NotificationRelay{})

	go func() {
		for {
			select {
			case <-time.After(time.Second * 2):
				exchange.Relay(NotificationRelay{}).Call("SendNotification", fmt.Sprintf("Group %d", group))
				if group == 1 {
					group = 2
				} else {
					group = 1
				}
			}
		}
	}()

	http.Handle("/relayr/", exchange)
	http.Handle("/", http.FileServer(http.Dir(".")))

	http.ListenAndServe(":8080", nil)
}
