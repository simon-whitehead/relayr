package relayr

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"sync"
	"time"
)

type longPollConnection struct {
	e            *Exchange
	result       chan []byte
	timeoutChan  chan struct{}
	t            *time.Timer
	ConnectionID string
}

type longPollTransport struct {
	e           *Exchange
	connections map[string]longPollConnection
	clock       *sync.RWMutex
}

func (t *longPollTransport) clientInConnections(cid string) bool {
	_, exists := t.connections[cid]

	return exists
}

func (t *longPollTransport) addConnection(cid string) longPollConnection {
	lp := longPollConnection{
		e:            t.e,
		result:       make(chan []byte, 100),
		timeoutChan:  make(chan struct{}, 10),
		ConnectionID: cid,
	}

	t.connections[cid] = lp

	return lp
}

func (t *longPollTransport) withClient(cid string, fn func(c longPollConnection)) {
	if t.clientInConnections(cid) {
		fn(t.connections[cid])
	}
}

func newLongPollTransport(e *Exchange) *longPollTransport {
	lp := &longPollTransport{
		e:           e,
		connections: make(map[string]longPollConnection),
		clock:       &sync.RWMutex{},
	}

	return lp
}

func (t *longPollTransport) CallClientFunction(relay *Relay, cid, fn string, args ...interface{}) {
	buff := &bytes.Buffer{}
	encoder := json.NewEncoder(buff)

	encoder.Encode(struct {
		R string
		M string
		A []interface{}
	}{
		relay.Name,
		fn,
		args,
	})

	go t.withClient(cid, func(c longPollConnection) {
		// force a timeout if we block on sending too long..
		go func() {
			<-time.After(time.Second * 30)
			c.timeoutChan <- struct{}{}
		}()
		c.result <- buff.Bytes()
	})
}

func (t *longPollTransport) removeConnection(cid string) {
	delete(t.connections, cid)
}

func (t *longPollTransport) wait(w http.ResponseWriter, cid string) {
	var conn longPollConnection
	if !t.clientInConnections(cid) {
		conn = t.addConnection(cid)
	} else {
		conn = t.connections[cid]
	}

	select {
	case m := <-conn.result:
		io.Copy(w, bytes.NewBuffer(m))
	case <-conn.timeoutChan:
		buff := &bytes.Buffer{}
		encoder := json.NewEncoder(buff)
		encoder.Encode(struct {
			Z string
		}{
			"RECONNECT",
		})
		io.WriteString(w, buff.String())
		t.removeConnection(cid)
		t.e.removeFromAllGroups(cid)
	}
}
