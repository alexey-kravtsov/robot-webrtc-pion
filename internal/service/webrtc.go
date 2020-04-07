package service

import (
	"encoding/json"
	"log"
	"math/rand"

	pion "github.com/pion/webrtc/v2"

	gst "github.com/alexey-kravtsov/robot-webrtc-pion/internal/gstreamer-src"

	"github.com/pion/sdp/v2"
)

var pc *pion.PeerConnection
var dc *pion.DataChannel
var iceCandidates []pion.ICECandidateInit

func StartWebrtc(wchan <-chan Message, serialchan chan<- string, sigchan chan<- Message) {
	for {
		message := <-wchan
		switch message.Type {
		case "sdp":
			{
				offer := pion.SessionDescription{}
				if err := json.Unmarshal([]byte(message.Data), &offer); err != nil {
					log.Printf("Unable to deserialize offer %s \n", err)
					continue
				}

				sessionDescr := sdp.SessionDescription{}
				if err := sessionDescr.Unmarshal([]byte(offer.SDP)); err != nil {
					log.Printf("Unable to parse offer SDP %s \n", err)
					continue
				}

				codecID, err := sessionDescr.GetPayloadTypeForCodec(sdp.Codec{
					Name:      "H264",
					ClockRate: 90000,
				})
				if err != nil {
					log.Printf("Client does not support codec %s \n", err)
					continue
				}

				mediaEngine := pion.MediaEngine{}
				if err := mediaEngine.PopulateFromSDP(offer); err != nil {
					log.Printf("Unable to populate codecs from offer %s \n", err)
					continue
				}

				// Create a new RTCPeerConnection
				api := pion.NewAPI(pion.WithMediaEngine(mediaEngine))
				pc, err := api.NewPeerConnection(pion.Configuration{
					ICEServers: []pion.ICEServer{
						{
							URLs: []string{"stun:stun.l.google.com:19302"},
						}, {
							URLs: []string{"stun:stun.ekiga.net"},
						}, {
							URLs: []string{"stun:stun.ideasip.com"},
						}, {
							URLs: []string{"stun:stun.stunprotocol.org:3478"},
						}, {
							URLs: []string{"stun:stun.voiparound.com"},
						}, {
							URLs:           []string{"turn:numb.viagenie.ca"},
							Username:       "webrtc@live.com",
							Credential:     "muazkh",
							CredentialType: pion.ICECredentialTypePassword,
						},
					},
				})
				if err != nil {
					panic(err)
				}

				defer pc.Close()

				pc.OnICECandidate(func(candidate *pion.ICECandidate) {
					sendIceCandidate(sigchan, candidate)
				})

				pc.OnDataChannel(func(d *pion.DataChannel) {
					dc = d
					dc.OnMessage(func(m pion.DataChannelMessage) {
						serialchan <- string(m.Data)
						//log.Printf("Datachannel message: %s\n", string(m.Data))
					})
				})

				// Create a video track
				trackID := rand.Uint32()
				videoTrack, err := pc.NewTrack(codecID, trackID, "video", "pion2")
				if err != nil {
					log.Printf("Unable to create video track %s \n", err)
					continue
				}

				transOptions := pion.RtpTransceiverInit{
					Direction: pion.RTPTransceiverDirectionSendonly,
					SendEncodings: []pion.RTPEncodingParameters{
						pion.RTPEncodingParameters{
							RTPCodingParameters: pion.RTPCodingParameters{SSRC: trackID, PayloadType: codecID},
						},
					},
				}
				pc.AddTransceiverFromTrack(videoTrack, transOptions)

				pc.AddTransceiver(pion.RTPCodecTypeAudio, pion.RtpTransceiverInit{
					Direction: pion.RTPTransceiverDirectionInactive,
				})

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
				gst.CreatePipeline([]*pion.Track{videoTrack}).Start()
			}
		case "ice":
			{
				ice := pion.ICECandidateInit{}
				err := json.Unmarshal([]byte(message.Data), &ice)
				if err != nil {
					log.Printf("Unable to deserialize ICE candidate %s \n", err)
					continue
				}

				if pc == nil {
					iceCandidates = append(iceCandidates, ice)
					continue
				}

				if len(iceCandidates) != 0 {
					for _, queuedIce := range iceCandidates {
						pc.AddICECandidate(queuedIce)
					}
					iceCandidates = iceCandidates[:0]
				}

				pc.AddICECandidate(ice)
			}
		}
	}
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
