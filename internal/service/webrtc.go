package service

import (
	"encoding/json"
	"log"
	"math/rand"

	pion "github.com/pion/webrtc/v2"

	gst "github.com/alexey-kravtsov/robot-webrtc-pion/internal/gstreamer-src"
)

func StartWebrtc(wchan <-chan Message, serialchan chan<- string, sigchan chan<- Message) {
	// Everything below is the pion-WebRTC API! Thanks for using it ❤️.

	// Prepare the configuration
	config := pion.Configuration{
		ICEServers: []pion.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	}

	// Create a new RTCPeerConnection
	pc, err := pion.NewPeerConnection(config)
	if err != nil {
		panic(err)
	}

	defer pc.Close()

	pc.OnICECandidate(func(candidate *pion.ICECandidate) {
		sendIceCandidate(sigchan, candidate)
	})

	pc.OnDataChannel(func(d *pion.DataChannel) {

		d.OnMessage(func(m pion.DataChannelMessage) {
			serialchan <- string(m.Data)
		})
	})

	handleSignalingMessages(wchan, sigchan, pc)
}

func sendIceCandidate(sigchan chan<- Message, candidate *pion.ICECandidate) {
	if candidate == nil {
		return
	}

	data, err := json.Marshal(candidate.ToJSON())
	if err != nil {
		log.Printf("Unable to serialize ICE candidate %s \n", err)
		return
	}

	sigchan <- Message{"ice", string(data)}
}

func handleSignalingMessages(wchan <-chan Message, sigchan chan<- Message, pc *pion.PeerConnection) {
	for {
		message := <-wchan
		switch message.Type {
		case "sdp":
			{
				// Create a video track
				firstVideoTrack, err := pc.NewTrack(pion.DefaultPayloadTypeH264, rand.Uint32(), "video", "pion2")
				if err != nil {
					log.Printf("Unable to create video track %s \n", err)
					continue
				}
				_, err = pc.AddTrack(firstVideoTrack)
				if err != nil {
					log.Printf("Unable add track %s \n", err)
					continue
				}

				// Wait for the offer to be pasted
				offer := pion.SessionDescription{}
				err = json.Unmarshal([]byte(message.Data), &offer)
				if err != nil {
					log.Printf("Unable to deserialize offer %s \n", err)
					continue
				}

				// Set the remote SessionDescription
				err = pc.SetRemoteDescription(offer)
				if err != nil {
					log.Printf("Unable to set remote description %s \n", err)
					continue
				}

				// Create an answer
				answer, err := pc.CreateAnswer(nil)
				if err != nil {
					log.Printf("Unable to create answer %s \n", err)
					continue
				}

				// Sets the LocalDescription, and starts our UDP listeners
				err = pc.SetLocalDescription(answer)
				if err != nil {
					log.Printf("Unable to set local description %s \n", err)
					continue
				}

				// Output the answer in base64 so we can paste it in browser
				jsonAnswer, err := json.Marshal(answer)
				if err != nil {
					log.Printf("Unable to serialize answer %s \n", err)
					continue
				}

				sigchan <- Message{"sdp", string(jsonAnswer)}

				// Start pushing buffers on these tracks
				gst.CreatePipeline([]*pion.Track{firstVideoTrack}).Start()
			}
		case "ice":
			{
				ice := pion.ICECandidateInit{}
				err := json.Unmarshal([]byte(message.Data), &ice)
				if err != nil {
					log.Printf("Unable to deserialize ICE candidate %s \n", err)
					continue
				}

				pc.AddICECandidate(ice)
			}
		}
	}
}
