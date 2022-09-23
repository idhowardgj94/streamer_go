package app

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// Note: This is the Websocket handler function.
func ServeWs(hub *StreamerHub, w http.ResponseWriter, r *http.Request) {
	// upgrade the http request to socket connect
	w.Header().Set("Access-Control-Allow-Origin", "*")
	log.Println("inside serverWs")

	// TODO: for dev, allow cross site connection.
	upgrader.CheckOrigin = func(r *http.Request) bool {
		return true
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	client := &Client{hub: hub, conn: conn, send: make(chan []byte, 256)}
	client.SendGreetingMessage()
	{
		//	client.SendMessage(newConnectedMsg(string(GetRandomSession())))
	}
	// log.Println(client)

	// client.SendGreetingMessage()

	go client.ReceivedMessageHandler(hub)
	go client.BackgroundPingHandler()
	go client.BackgroundPingHandler()
	// TODO: not implemented
	client.HubEventHandler()

	client.hub.register <- client
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

// Note: path use mux.Vars(request) to get the route var mapping
// url parameter: r.URL.Qery().Get(key)

type MeteorWebsocketResponse struct {
	CookieNeeded bool     `json:"cookie_needed"`
	Origins      []string `json:"origins"`
	Entropy      int      `json:"entropy"`
	Websocket    bool     `json:"websocket"`
}

func defaultMeteorWebsocketResponse() MeteorWebsocketResponse {
	return MeteorWebsocketResponse{
		CookieNeeded: true,
		Origins:      []string{"*:*"},
		// TODO: need to be a random value
		Entropy:   19911118,
		Websocket: true,
	}
}

func MeteorWebsocketRequestHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Inside MeteorWebsocketRequsetHandler")
	w.Header().Set("content-type", "application/json")
	//	Access-Control-Allow-Origin
	w.Header().Set("Access-Control-Allow-Origin", "*")
	j, err := json.Marshal(defaultMeteorWebsocketResponse())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"msg": "Something went wrong."})
	}

	w.Write(j)
}

// Note: Both websocket server and app server
// No need to separate this two server now.
type AppServer struct {
	r      *mux.Router
	hub    *StreamerHub
	client *Client
}

func ExecuteWebServer() {
	log.Println("[Info] starting Hub...")
	hub := NewHub()
	go hub.run()

	log.Println("[Info] starting exectue webserver...")
	r := mux.NewRouter()
	r.Use(mux.CORSMethodMiddleware(r))

	// This is for frontend connected client.
	r.HandleFunc("/sockjs/info", MeteorWebsocketRequestHandler)
	r.HandleFunc("/sockjs/{random_number}/{random}/websocket", func(w http.ResponseWriter, r *http.Request) {
		ServeWs(hub, w, r)
	})

	// This is for bot connected client
	r.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
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
