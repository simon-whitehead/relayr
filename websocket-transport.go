package relayr

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
)

type connection struct {
	ws  *websocket.Conn
	out chan []byte
	c   *webSocketTransport
	id  string
	e   *Exchange
}

type webSocketTransport struct {
	connections  map[string]*connection
	connected    chan *connection
	disconnected chan *connection
	e            *Exchange
}

type webSocketClientMessage struct {
	Server       bool          `json:"S"`
	Relay        string        `json:"R"`
	Method       string        `json:"M"`
	Arguments    []interface{} `json:"A"`
	ConnectionID string        `json:"C"`
}

func newWebSocketTransport(e *Exchange) *webSocketTransport {
	c := &webSocketTransport{
		connected:    make(chan *connection),
		disconnected: make(chan *connection),
		connections:  make(map[string]*connection),
		e:            e,
	}

	go c.listen()

	return c
}

func (c *webSocketTransport) listen() {
	for {
		select {
		case conn := <-c.connected:
			c.connections[conn.id] = conn
		case conn := <-c.disconnected:
			if _, ok := c.connections[conn.id]; ok {
				delete(c.connections, conn.id)
				close(conn.out)
			}
		}
	}
}

func (c *webSocketTransport) CallClientFunction(relay *Relay, fn string, args ...interface{}) {
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

	o := c.connections[relay.ConnectionID]

	if o != nil {
		o.out <- buff.Bytes()
	}
}

func (c *connection) read() {
	for {
		_, message, err := c.ws.ReadMessage()
		if err != nil {
			break
		}
		var m webSocketClientMessage
		err = json.Unmarshal(message, &m)
		if err != nil {
			fmt.Println("ERR:", err)
			continue
		}

		relay := c.e.getRelayByName(m.Relay, m.ConnectionID)

		if m.Server {
			err := c.e.callRelayMethod(relay, m.Method, m.Arguments...)
			if err != nil {
				fmt.Println("ERR:", err)
			}
		} else {
			c.c.CallClientFunction(relay, m.Method, m.Arguments)
		}
	}
	c.ws.Close()
}

func (c *connection) write() {
	for message := range c.out {
		err := c.ws.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			break
		}
	}
	c.ws.Close()
}

type socketHandler struct {
	c *webSocketTransport
	e *Exchange
}

func (wsh socketHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	c := &connection{e: wsh.e, out: make(chan []byte, 256), ws: ws, c: wsh.c, id: r.URL.Query()["connectionId"][0]}
	c.c.connected <- c
	defer func() { c.c.disconnected <- c }()
	go c.write()
	c.read()
}
