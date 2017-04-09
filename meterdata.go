package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"
)

type MeterData struct {
	Time             time.Time
	buf              []byte
	parsed           bool
	parseErr         error
	PlusEnergyTar1   float64 // opgenomen energie tarief 1 in kWh
	MinEnergyTar1    float64 // teruggeleverde energie tarief 1 in kWh
	PlusEnergyTar2   float64 // opgenomen energie tarief 2 in kWh
	MinEnergyTar2    float64 // teruggeleverde tarief 2 in kWh
	CurrentTarNumber float64 // huidige tarief nr
	CurrentPlusPower float64 // huidige opgenomen vermogen in kW
	CurrentMinPower  float64 // huidige teruggeleverde vermogen in kW
	GasUsed          float64 // gas verbruik in m3

	sync.Mutex
}

// Append adds data to md.buf and returns true if the block is complete.
// The last line begins with ! followed by 4 characters and CRLF, since we do not use these we
// can stop at ! although usually we get the full line it could be we end with !.
func (md *MeterData) Append(b []byte) (bool, error) {
	md.Lock()
	defer md.Unlock()

	if md.Time.IsZero() {
		md.Time = time.Now().UTC()
	}

	md.buf = append(md.buf, b...)
	if bytes.ContainsRune(b, '!') {
		return true, nil
	}
	if len(md.buf) > 4096 {
		return true, errors.New("buffer too big, receiving invalid data")
	}
	return false, nil
}

// Parse fills our variables from buf
func (md *MeterData) Parse() error {
	md.Lock()
	defer md.Unlock()

	if md.parsed || len(md.buf) == 0 {
		return nil
	}

	scanner := bufio.NewScanner(bytes.NewReader(md.buf))
	for scanner.Scan() {
		b := scanner.Bytes()
		switch string(b[:9]) {
		case "1-0:1.8.1":
			md.setVal(&md.PlusEnergyTar1, b, kWhVal)
		case "1-0:1.8.2":
			md.setVal(&md.MinEnergyTar1, b, kWhVal)
		case "1-0:2.8.1":
			md.setVal(&md.PlusEnergyTar2, b, kWhVal)
		case "1-0:2.8.2":
			md.setVal(&md.MinEnergyTar2, b, kWhVal)
		case "0-0:96.14":
			md.setVal(&md.CurrentTarNumber, b, tarVal)
		case "1-0:1.7.0":
			md.setVal(&md.CurrentPlusPower, b, kWVal)
		case "1-0:2.7.0":
			md.setVal(&md.CurrentMinPower, b, kWVal)
		case "0-1:24.2.":
			md.setVal(&md.GasUsed, b, m3Val)
		}
		if md.parseErr != nil {
			return md.parseErr
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("Error scanning data: %s", err)
	}

	md.parsed = true
	return nil
}

func (md *MeterData) Json() ([]byte, error) {
	md.Lock()
	defer md.Unlock()

	return json.Marshal(md)
}

func (md *MeterData) setVal(f *float64, line []byte, pf parseFunc) {
	if md.parseErr != nil {
		return
	}
	val, err := pf(line)
	if err != nil {
		md.parseErr = err
		return
	}
	*f = val
}

type parseFunc func(line []byte) (float64, error)

func kWhVal(line []byte) (float64, error) {
	if len(line) < 20 {
		return 0, errors.New("unexpected line length")
	}
	//fmt.Println(string(line[10:20]))
	return strconv.ParseFloat(string(line[10:20]), 64)
}

func kWVal(line []byte) (float64, error) {
	if len(line) < 16 {
		return 0, errors.New("unexpected line length")
	}
	//fmt.Println(string(line[10:16]))
	return strconv.ParseFloat(string(line[10:16]), 64)
}

func m3Val(line []byte) (float64, error) {
	if len(line) < 35 {
		return 0, errors.New("unexpected line length")
	}
	//fmt.Println(string(line[26:35]))
	return strconv.ParseFloat(string(line[26:35]), 64)
}

func tarVal(line []byte) (float64, error) {
	if len(line) < 16 {
		return 0, errors.New("unexpected line length")
	}
	//fmt.Println(string(line[12:16]))
	return strconv.ParseFloat(string(line[12:16]), 64)
}
