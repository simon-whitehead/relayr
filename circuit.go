package relayR

// Circuit provides a contract for transports to send data
// to connected clients based on their capabilities.
type circuit interface {
	Send(m clientMessage)
	Receive(m clientMessage)
}
