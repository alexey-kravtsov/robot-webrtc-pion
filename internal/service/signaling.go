package service

import (
	"log"

	"github.com/gorilla/websocket"
)

var websocketConn *websocket.Conn

func StartSignaling(sigchan <-chan Message, wchan chan<- Message) {
	websocketConn, _, err := websocket.DefaultDialer.Dial("wss://robot-signaling-server.herokuapp.com/signaling/robot", nil)
	//websocketConn, _, err := websocket.DefaultDialer.Dial("ws://localhost:8080/signaling/robot", nil)
	if err != nil {
		log.Fatal(err)
	}
	defer websocketConn.Close()

	go sendIncoming(sigchan, websocketConn)

	for {
		var m Message
		err := websocketConn.ReadJSON(&m)
		if err != nil {
			log.Printf("Unable to deserialize message %s \n", err)
			continue
		}

		switch m.Type {
		case "sdp", "ice":
			wchan <- m
		default:
			log.Printf("Unknown message type: %s \n", m.Type)
		}
	}
}

func sendIncoming(schan <-chan Message, c *websocket.Conn) {
	for {
		message := <-schan
		err := c.WriteJSON(message)
		if err != nil {
			log.Printf("Error sending signaling message %s \n", err)
		}
	}
}
