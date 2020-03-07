package main

import (
	//"github.com/alexey-kravtsov/robot-webrtc-pion/internal/webrtcservice"
	"log"

	"github.com/gorilla/websocket"
)

func main() {
	//webrtcservice.Start()
	c, _, err := websocket.DefaultDialer.Dial("ws://localhost:8080/signaling/robot", nil)
	if err != nil {
		log.Fatal(err)
		return
	}
	defer c.Close()

	go handleSignalingMessages(c)

	// Block forever
	select {}
}

func handleSignalingMessages(c *websocket.Conn) {
	for {
		var message sdpMessage
		c.ReadJSON(&message)
		log.Println(message)
	}
}

type sdpMessage struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}
