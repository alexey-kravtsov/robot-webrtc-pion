package service

import (
	"log"

	"github.com/gorilla/websocket"
)

func StartSignaling(sigchan <-chan Message, wchan chan<- Message) {
	c, _, err := websocket.DefaultDialer.Dial("ws://192.168.0.147:8080/signaling/robot", nil)
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()

	go sendIncoming(sigchan, c)

	for {
		var m Message
		err := c.ReadJSON(&m)
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
