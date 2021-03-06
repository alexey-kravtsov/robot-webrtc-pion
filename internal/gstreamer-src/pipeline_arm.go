// +build arm arm64

package gst

import "log"

func getPipeline() string {
	log.Printf("Using omxh264enc encoder")

	return `v4l2src 
	! video/x-raw, width=640, height=480 
	! queue 
	! videoconvert 
	! video/x-raw,format=I420 
	! queue
	! omxh264enc control-rate=1 target-bitrate=600000 
	! queue
	! h264parse config-interval=3 
	! video/x-h264,stream-format=byte-stream 
	! appsink name=appsink`
}
