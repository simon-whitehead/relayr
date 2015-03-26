package relayr

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"time"
)

type timeout struct {
	reset bool
}

type longPollConnection struct {
	e           *Exchange
	result      chan []byte
	timeoutChan chan timeout
}

type longPollTransport struct {
	e           *Exchange
	connections map[string]longPollConnection
}

func newLongPollConnection(e *Exchange) longPollConnection {
	lp := longPollConnection{
		e:           e,
		result:      make(chan []byte, 100),
		timeoutChan: make(chan timeout),
	}

	go func() {
	out:
		for {
			select {
			case <-time.After(time.Second * 30):
				lp.timeoutChan <- timeout{reset: false} //reset the timeout
				break out
			}
		}
		close(lp.result)
		close(lp.timeoutChan)
	}()

	return lp
}

func newLongPollTransport(e *Exchange) *longPollTransport {
	return &longPollTransport{
		e:           e,
		connections: make(map[string]longPollConnection),
	}
}

func (t *longPollTransport) CallClientFunction(relay *Relay, fn string, args ...interface{}) {
	if c, ok := t.connections[relay.ConnectionID]; ok {
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

		c.result <- buff.Bytes()
	}
}

func (t *longPollTransport) wait(w http.ResponseWriter, cid string) {
	t.connections[cid] = newLongPollConnection(t.e)

	select {
	case m := <-t.connections[cid].result:
		io.Copy(w, bytes.NewBuffer(m))
	case <-t.connections[cid].timeoutChan:
		buff := &bytes.Buffer{}
		encoder := json.NewEncoder(buff)
		encoder.Encode(struct {
			Z string
		}{
			"RECONNECT",
		})
		io.WriteString(w, buff.String())
		delete(t.connections, cid)
		t.e.removeFromAllGroups(cid)
	}
}
