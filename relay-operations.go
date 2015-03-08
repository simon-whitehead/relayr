package relayR

import "reflect"

type relayOperations struct {
	connectionID string
	t            reflect.Type
	e            *Exchange
}

func (r relayOperations) Call(m string, a ...interface{}) {
	r.e.callMethod(r.t, m, r.connectionID, a...)
}
