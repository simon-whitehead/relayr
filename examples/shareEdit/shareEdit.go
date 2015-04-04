package main

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/simon-whitehead/relayr"
)

type shareEdit struct {
	val int64
}

func (this *shareEdit) Get(relay *relayr.Relay, cid string) {
	relay.CallClient(cid, "updateValue", fmt.Sprint(this.val))
}

func (this *shareEdit) Add(relay *relayr.Relay, valstr string) {
	if val, err := strconv.ParseInt(valstr, 10, 64); err == nil {
		this.val += val
		relay.Clients.All("updateValue", fmt.Sprint(this.val))
	}
}

func (this *shareEdit) Del(relay *relayr.Relay, valstr string) {
	if val, err := strconv.ParseInt(valstr, 10, 64); err == nil {
		this.val -= val
		relay.Clients.All("updateValue", fmt.Sprint(this.val))
	}
}

func main() {
	exchange := relayr.NewExchange()
	exchange.RegisterRelay(&shareEdit{})

	http.Handle("/relayr/", exchange)
	http.Handle("/", http.FileServer(http.Dir(".")))
	http.ListenAndServe(":8080", nil)
}
