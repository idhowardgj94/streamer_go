package app

import "log"

// StreamerHub
// StreamerHub store the client connection.
// And channel to handler client connection event.
type StreamerHub struct {
	// Mapping to store register client.
	clients map[*Client]bool

	broadcast chan []byte

	// Register client chan.
	register chan *Client

	// Unregister client chan
	unregister chan *Client
}

// Initializer
func NewHub() *StreamerHub {
	return &StreamerHub{
		clients:    map[*Client]bool{},
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

// run the Hub service
func (it *StreamerHub) run() {
	for {
		select {
		case client := <-it.register:
			it.clients[client] = true
		case client := <-it.unregister:
			if _, ok := it.clients[client]; ok {
				delete(it.clients, client)
				close(client.send)
			}

		case message := <-it.broadcast:
			for client := range it.clients {
				select {
				case client.send <- message:
					// empty
				default:
					// delete the message which can't send
					log.Println("inside close in line 195")
					close(client.send)
					delete(it.clients, client)

				}
			}
		}
	}
}
