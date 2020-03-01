package main

import (
	"github.com/alexey-kravtsov/robot-webrtc-pion/internal/webrtcservice"
)

func main() {
	webrtcservice.Start()

	// Block forever
	select {}
}
