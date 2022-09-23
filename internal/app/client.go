package app

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/gorilla/websocket"
	"github.com/segmentio/ksuid"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	// pingPeriod = (pongWait * 9) / 10

	pingPeriod = 10 * time.Second

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

// Client is a 'middleman' between the websocket connection and the hub.
type Client struct {
	// Deprecated, remove in the futrue
	hub *StreamerHub

	// websocket communication
	conn *websocket.Conn

	// the send channel to handle / send message.
	send chan []byte
}

type ClientAction interface {
	SendMessage(msg interface{})
	// This is for Meteor client only.
	SendGreetingMessage()
	// Handle ping handler.
	BackgroundPingHandler()
	// Received client sended message.
	ReceivedMessageHandler(hub *StreamerHub)
	// hub event handler
	HubEventHandler()
}

func (it *Client) HubEventHandler() {
	// TODO: not yet impletetment.
}

func (it *Client) SendMessage(msg interface{}) {
	it.conn.SetWriteDeadline(time.Now().Add(writeWait))
	w, e := it.conn.NextWriter(websocket.TextMessage)
	if e != nil {
		log.Println("[Error] SendMessage. e: ", e)
		return
	}
	// Note: type switches
	switch msg.(type) {
	case string:
		w.Write([]byte(msg.(string)))
	default:
		msgString, err := json.Marshal(msg)
		if err != nil {
			log.Println("[Error] SendMessage. e: ", e)
		}
		// Add a is meteor client
		// TODO: need to add a switch
		w.Write([]byte(fmt.Sprintf("a%s", msgString)))
	}
	w.Close()
}

func (it *Client) SendGreetingMessage() {
	it.conn.SetWriteDeadline(time.Now().Add(writeWait))
	w, e := it.conn.NextWriter(websocket.TextMessage)
	if e != nil {
		log.Println("[Error] SendGreetingMessage. e: ", e)
		return
	}
	w.Write([]byte("o"))
	w.Close()
}

func (it *Client) BackgroundPingHandler() {
	log.Println("[Info] Start background ping handler")
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		log.Println("[Info] Background ping handler close")
		ticker.Stop()
		it.conn.Close()
	}()

	for {
		select {
		case <-ticker.C:
			it.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := it.conn.WriteMessage(websocket.PingMessage, []byte("ping")); err != nil {
				return
			}
		}
	}
}

type ConnectedMsg struct {
	Msg     string `json:"msg"`
	Session string `json:"session"`
}

func newConnectedMsg(session string) ConnectedMsg {
	return ConnectedMsg{
		Msg:     "connected",
		Session: session,
	}
}

func GetRandomSession() string {
	id := ksuid.New()
	return id.String()
}

func (it *Client) ReceivedMessageHandler(hub *StreamerHub) {
	defer func() {
		// TODO: handler unregister
		it.conn.Close()
	}()
	it.conn.SetReadLimit(maxMessageSize)
	// TODO: pong handler?

	// Note: input is an array object, for example:
	// [{"msg":"connect","version":"1","support":["1","pre2","pre1"]}]
	// var input map[string]interface{}

	for {
		op, msg, err := it.conn.ReadMessage()
		if op == websocket.CloseMessage {
			log.Println("[Warn] client close the websocket. return...")
			return
		}
		if err != nil {
			log.Println("[Error] ", err)
			return
		}
		// log.Println(input)
		log.Println("[Info] Received message: ", string(msg))
		var input map[string]interface{}

		it.UnMarshalMsg(msg, &input)

		// TODO: dispatch here
		msgType := input["msg"]
		switch msgType {
		case "connect":
			msg := newConnectedMsg(string(GetRandomSession()))
			it.SendMessage(msg)
		}
	}
}
func (it *Client) UnMarshalMsg(msg []byte, msgObj *map[string]interface{}) {
	var unwrapArr []string
	var err error
	err = json.Unmarshal(msg, &unwrapArr)
	if err != nil {
		log.Println("[Error] ", err)
	}
	log.Println("unwrapArr", unwrapArr[0])

	err = json.Unmarshal([]byte(unwrapArr[0]), &msgObj)
	if err != nil {
		log.Println("[Error] ", err)
	}
	// TODO: error handling
}
