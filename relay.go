package relayr

import "reflect"

type Relay struct {
	Name         string
	ConnectionID string
	methods      []string
	t            reflect.Type
	exchange     *Exchange
}

func (r *Relay) Call(fn string, args ...interface{}) {
	r.exchange.callRelayMethod(r, fn, args...)
}
