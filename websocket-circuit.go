package relayR

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = &websocket.Upgrader{ReadBufferSize: 1024, WriteBufferSize: 1024}

type webSocketCircuit struct {
	connections  map[string]*connection
	connected    chan *connection
	disconnected chan *connection
	e            *Exchange
}

type connection struct {
	ws  *websocket.Conn
	out chan []byte
	c   *webSocketCircuit
	id  string
	e   *Exchange
}

func newWebSocketCircuit(e *Exchange) *webSocketCircuit {
	c := &webSocketCircuit{
		connected:    make(chan *connection),
		disconnected: make(chan *connection),
		connections:  make(map[string]*connection),
		e:            e,
	}

	go c.listen()

	return c
}

func (c *webSocketCircuit) listen() {
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

func (c *webSocketCircuit) Send(m clientMessage) {
	buff := &bytes.Buffer{}
	encoder := json.NewEncoder(buff)

	encoder.Encode(struct {
		R string
		M string
		A []interface{}
	}{
		m.R,
		m.M,
		m.A,
	})

	o := c.connections[m.C]
	if o != nil {
		o.out <- buff.Bytes()
	}
}

func (c *webSocketCircuit) Receive(m clientMessage) {
	if m.S {
		relay := c.e.getRelayByName(m.R)
		err := c.e.callMethod(relay.t, m.M, m.C, m.A...)
		if err != nil {
			fmt.Println(err)
		}
	} else {
		c.Send(m)
	}
}

func (c *connection) read() {
	for {
		_, message, err := c.ws.ReadMessage()
		if err != nil {
			break
		}
		var m clientMessage
		err = json.Unmarshal(message, &m)
		c.c.Receive(m)
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
	c *webSocketCircuit
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
