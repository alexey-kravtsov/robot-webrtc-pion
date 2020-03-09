package main

import "github.com/alexey-kravtsov/robot-webrtc-pion/internal/service"

func main() {
	sigchan := make(chan service.Message, 10)
	wchan := make(chan service.Message, 10)
	serialchan := make(chan string, 10)

	go service.StartSignaling(sigchan, wchan)
	go service.StartWebrtc(wchan, serialchan, sigchan)
	go service.StartSerial(serialchan)

	// Block forever
	select {}
}
