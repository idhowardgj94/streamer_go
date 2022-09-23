package main

import (
	"bytes"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"time"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	// pingPeriod = (pongWait * 9) / 10

	pingPeriod = 3 * time.Second

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// Client is a 'middleman' between the websocket connection and the hub.
type Client struct {
	hub *Hub

	// websocket communication
	conn *websocket.Conn

	// the send channel to handle / send message.
	send chan []byte
}

// Implement a pump method
// A goroutine running writePump is strated for each connection.
// The application ensures that there is at most one writer to a conncetion by
// executing all writes from the goroutine.
func (it *Client) writePump() {
	log.Println("Start writePump...")
	ticker := time.NewTicker(pingPeriod)

	defer func() {
		log.Println("writePump close")
		ticker.Stop()
		it.conn.Close()
	}()

	for {
		log.Println("for")
		select {
		// A message come from hub
		// which is the message need te be send to client.
		case message, ok := <-it.send:
			// set timeout time
			it.conn.SetWriteDeadline(time.Now().Add(writeWait))
			// if close the send channel, just close the connection to client.
			if !ok {
				it.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			// send message to client
			w, err := it.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued chat message to the current websocket message
			// conusmed queued chat message to empty
			// TODO: Abstract
			n := len(it.send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-it.send)
			}

			if err := w.Close(); err != nil {
				return
			}

		// recived next ticker, should write a message to client side.
		case <-ticker.C:
			log.Println("write ping...")
			it.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := it.conn.WriteMessage(websocket.PingMessage, []byte("ping")); err != nil {
				return
			}
		}
	}
}

// Implement read pump
func (it *Client) readPump() {
	defer func() {
		it.hub.unregister <- it
		it.conn.Close()
	}()
	it.conn.SetReadLimit(maxMessageSize)
	it.conn.SetReadDeadline(time.Now().Add(pongWait))
	it.conn.SetPongHandler(func(string) error { it.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := it.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
		it.hub.broadcast <- message
	}
}

// Hub
// Hub store the client connection.
// And channel to handler client connection event.
type Hub struct {
	// Mapping to store register client.
	clients map[*Client]bool

	broadcast chan []byte

	// Register client chan.
	register chan *Client

	// Unregister client chan
	unregister chan *Client
}

// Initializer
func NewHub() *Hub {
	return &Hub{
		clients:    map[*Client]bool{},
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

// run the Hub service
func (it *Hub) run() {
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
					close(client.send)
					delete(it.clients, client)

				}
			}
		}
	}
}

func ServeWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	// upgrade the http request to socket connect
	w.Header().Set("Access-Control-Allow-Origin", "*")
	log.Println("inside serverWs")
	conn, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		log.Println(err)
		return
	}

	client := &Client{hub: hub, conn: conn, send: make(chan []byte, 256)}
	client.hub.register <- client

	go client.writePump()
	go client.readPump()

}

// TODO: demo, delete me
func ServeHome(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL)
	if r.URL.Path != "/" {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	http.ServeFile(w, r, "home.html")
}

func main() {
	hub := NewHub()
	go hub.run()
	r := mux.NewRouter()
	r.Use(mux.CORSMethodMiddleware(r))
	r.PathPrefix("/ws/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("inside ws handler")
		ServeWs(hub, w, r)
	})
	r.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		log.Println("inside ws handler")
		ServeWs(hub, w, r)
	})

	// TODO: delete me
	r.HandleFunc("/", ServeHome)
	http.Handle("/", r)
	err := http.ListenAndServe(":3999", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}

}
