// +build !arm,!arm64

package gst

import "log"

func getPipeline() string {
	log.Printf("Using x264enc encoder")

	return `v4l2src 
	! video/x-raw, width=640, height=480 
	! videoconvert 
	! queue 
	! video/x-raw,format=I420 
	! x264enc speed-preset=ultrafast tune=zerolatency key-int-max=20 threads=4 sliced-threads=true 
	! video/x-h264,stream-format=byte-stream 
	! appsink name=appsink`
}
