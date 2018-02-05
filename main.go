package main

import (
	"fmt"
	"encoding/csv"
	"os"
	"io"
	"time"
	"strconv"
	"taxistream/osrm"
	"encoding/json"
	"taxistream/taxisim"
)

// General function to panic on an error.
func check(e error) {
	if e != nil {
		panic(e)
	}
}

// Configure this program using the following parameters.
type Configuration struct {
	Mode                 string
	TaxiData             []string
	NumTaxis             int32
	TargetSpeedPerSecond float32
}

// Reads a configuration file.
// The file needs to adhere to the Configuration struct.
func readConfig(configFile string) Configuration {
	file, err1 := os.Open(configFile)
	check(err1)
	defer file.Close()
	decoder := json.NewDecoder(file)
	configuration := Configuration{}
	err2 := decoder.Decode(&configuration)
	check(err2)
	return configuration
}

// Wraps the CSV processing functionality.
func processTaxiDataCSV(filename string, simulator taxisim.Simulator,
	processRow func([]string, taxisim.Simulator) taxisim.Simulator) taxisim.Simulator {

	file, err := os.Open(filename)
	check(err)
	defer file.Close()

	reader := csv.NewReader(file)
	reader.Comma = ','
	lineCount := 0
	reader.Read()

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			fmt.Println("Error:", err)
			return simulator
		}
		simulator = processRow(record, simulator)

		lineCount += 1
		if lineCount > 3 {
			return simulator
		}
	}
	return simulator
}

// Processes a single taxi record.
func processTaxiRecord(record []string, simulator taxisim.Simulator) taxisim.Simulator {
	// fmt.Println("Record", lineCount, "is", record, "and has", len(record), "fields")
	puTime, _ := time.Parse("2006-01-02 15:04:05", record[1])
	puLon, _ := strconv.ParseFloat(record[5], 32)
	puLat, _ := strconv.ParseFloat(record[6], 32)
	doLon, _ := strconv.ParseFloat(record[7], 32)
	doLat, _ := strconv.ParseFloat(record[8], 32)
	route, err := osrm.QueryOSRM(puTime, puLon, puLat, doLon, doLat)
	if err != nil {
		panic(err)
	}
	simulator = taxisim.ProcessRoute(*route, simulator)
	return simulator
}

// Main function of the processing program.
func main() {
	conf := readConfig("config.json")
	fmt.Println(conf)

	simulator := taxisim.SetUpSimulation(conf.NumTaxis)
	simulator = processTaxiDataCSV(conf.TaxiData[0], simulator, processTaxiRecord)

	fmt.Println(simulator)
}
