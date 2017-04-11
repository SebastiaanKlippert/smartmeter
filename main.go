package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

var (
	apiHost = flag.String("api", "", "location of smartmeter-api")
	apiUrl  = ""
)

func main() {

	flag.Parse()

	if *apiHost == "" {
		*apiHost = os.Getenv("API_HOST")
	}
	if *apiHost == "" {
		log.Fatal("api flag or API_HOST not set")
	}
	log.Printf("Using API host %s", *apiHost)

	apiUrl = fmt.Sprintf("http://%s/meterdata", *apiHost)

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

	// Create JSON output
	jsonData, err := data.Json()
	if err != nil {
		return err
	}

	// Send JSON to our API
	resp, err := http.Post(apiUrl, "application/json", bytes.NewReader(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Received statuscode %d from API", resp.StatusCode)
	}

	return nil
}
