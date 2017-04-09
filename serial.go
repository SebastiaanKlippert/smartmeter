package main

import (
	"fmt"

	"github.com/tarm/serial"
)

func OpenPort() (*serial.Port, error) {

	cfg := &serial.Config{
		Name:     "/dev/ttyUSB0",
		Baud:     115200,
		Size:     8,
		StopBits: 1,
	}

	port, err := serial.OpenPort(cfg)
	if err != nil {
		return nil, err
	}

	return port, nil
}

func ReadLoop(port *serial.Port, meterChan chan *MeterData) {

	buf := make([]byte, 1024)
	var data *MeterData

	for {

		bits, err := port.Read(buf)
		if err != nil {
			fmt.Println(err)
			continue
		}

		if bits == 0 {
			//should not happen in blocking mode
			continue
		}

		//a new data block always begins with /
		if buf[0] == '/' {
			data = new(MeterData)
		}
		if data == nil {
			//discard any data if we are not "within" a data block
			continue
		}

		complete, err := data.Append(buf[:bits])
		if err != nil {
			fmt.Println(err)
			continue
		}

		if complete {
			meterChan <- data
			data = nil
		}

	}

}
