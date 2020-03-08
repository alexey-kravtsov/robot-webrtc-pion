package main

import (
	"github.com/alexey-kravtsov/robot-webrtc-pion/internal/service"
)

func main() {
	signaling := make(chan service.Message, 10)
	webrtc := make(chan service.Message, 10)

	go service.StartSignaling(signaling, webrtc)
	go service.StartWebrtc(webrtc, signaling)

	// Block forever
	select {}
}
