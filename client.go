package relayr

type Client struct {
	ConnectionID string
	exchange     *Exchange
	transport    Transport
}

type clientMessage struct {
	Relay     string `json:"R"`
	Function  string `json:"F"`
	Arguments string `json:"A"`
}

func (c *Client) Exchange() *Exchange {
	return c.exchange
}

func (c *Client) Call(relay *Relay, fn string, args ...interface{}) {
	c.transport.CallClientFunction(relay, fn, args)
}
