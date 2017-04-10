package main

import (
	"errors"
	"fmt"
	"log"
	"time"
)

func main() {

	for {

		log.Println("Starting reading data...")

		err := Start()
		if err != nil {
			log.Println(err)
			time.Sleep(4 * time.Second)
			continue
		}

	}

}

func Start() error {

	port, err := OpenPort()
	if err != nil {
		log.Fatalf("error opening port %s", err)
	}

	meterChan := make(chan *MeterData)

	go ReadLoop(port, meterChan)

	for {
		select {
		case data := <-meterChan:
			if err := processData(data); err != nil {
				return fmt.Errorf("Error in processData: %s", err)
			}
		case <-time.After(time.Minute):
			return errors.New("No data received for 1 minute")
		}
	}
	return nil
}

func processData(data *MeterData) error {

	// Parse data first
	if err := data.Parse(); err != nil {
		return err
	}

	//TODO store somewhere

	jsonData, err := data.Json()
	if err != nil {
		return err
	}
	fmt.Println(string(jsonData))

	return nil
}
