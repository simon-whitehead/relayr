package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/simon-whitehead/relayR"
)

var group int = 1

type NotificationRelay struct {
	relayR.Relay
}

func (r NotificationRelay) SendNotification(g string) {
	r.Clients.Group(g, "notificationReceive", "Notification for group "+g)
}

func (r NotificationRelay) Subscribe(g string) {
	r.Clients.AddToGroup(g)
}

func (r NotificationRelay) Unsubscribe(g string) {
	r.Clients.RemoveFromGroup(g)
}

func main() {
	exchange := relayR.NewExchange()
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

	relayR.Handle("/~relayr", exchange)
	http.Handle("/", http.FileServer(http.Dir(".")))

	http.ListenAndServe(":8080", nil)
}
