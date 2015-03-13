package relayr

type client struct {
	ConnectionID string
	exchange     *Exchange
	transport    Transport
}

type clientMessage struct {
	Relay     string `json:"R"`
	Function  string `json:"F"`
	Arguments string `json:"A"`
}
