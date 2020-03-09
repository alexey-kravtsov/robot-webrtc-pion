package service

import (
	"fmt"

	"github.com/jacobsa/go-serial/serial"

	"log"
)

func StartSerial(serialchan <-chan string) {
	options := serial.OpenOptions{
		PortName:        "/dev/ttyACM0",
		BaudRate:        115200,
		DataBits:        8,
		StopBits:        1,
		MinimumReadSize: 4,
	}

	// Open the port.
	port, err := serial.Open(options)
	if err != nil {
		log.Fatalf("Unable to open serial port: %s", err)
	}

	// Make sure to close it later.
	defer port.Close()

	// write first byte
	if _, err = port.Write([]byte{0x01}); err != nil {
		log.Fatalf("Unable to open serial port: %s", err)
	}

	for {
		m := <-serialchan
		bytes, err := convertPayload(m)
		if err != nil {
			log.Printf("Unable to convert payload: %s \n", m)
			continue
		}

		if _, err = port.Write(bytes); err != nil {
			log.Printf("Unable to write payload: %s \n", m)
		}
	}
}

func convertPayload(data string) ([]byte, error) {
	const headerLength = 3
	const flag = 0
	const version = 1

	length := len(data)

	if length == 2 {
		var bytes [headerLength + 2]byte
		bytes[0] = flag
		bytes[1] = byte(len(bytes))
		bytes[2] = version
		bytes[3] = data[0]
		bytes[4] = data[1]
		return bytes[:], nil
	}

	if length == 5 {
		var bytes [headerLength + 5]byte
		bytes[0] = flag
		bytes[1] = byte(len(bytes))
		bytes[2] = version
		bytes[3] = data[0]
		bytes[4] = data[1]
		bytes[5] = data[2]
		bytes[6] = data[3]

		var speed byte
		switch data[4] {
		case '0':
			speed = 0
		case '1':
			speed = 1
		case '2':
			speed = 2
		case '3':
			speed = 3
		default:
			{
				return nil, fmt.Errorf("Incorrect speed: %s", data[4])
			}
		}

		bytes[7] = speed
		return bytes[:], nil
	}

	return nil, fmt.Errorf("Incorrect data length: %s", length)
}
