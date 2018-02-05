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
	"database/sql"

	"github.com/cridenour/go-postgis"
	_ "github.com/lib/pq"
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
		if lineCount > 100 {
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

// Sets up the connection to the database.
func connectToDatabase() *sql.DB {
	db, err := sql.Open("postgres",
		"host=127.0.0.1 port=5432 dbname=taxi-streaming sslmode=disable user=dobucher password=dominik")
	if err != nil {
		panic(err)
	}
	err = db.Ping()
	if err != nil {
		panic(err)
	}
	// Ensure we have PostGIS on the table.
	db.Exec("CREATE EXTENSION IF NOT EXISTS postgis;")

	// And clean database as well.
	db.Exec("TRUNCATE TABLE taxi_routes;")
	return db
}

// Takes the output of a simulation run and writes it to PostGIS.
func writeSimulatorOutputToDatabase(simulator taxisim.Simulator) {
	db := connectToDatabase()
	defer db.Close()

	point := postgis.Point{-84.5014, 39.1064}
	var newPoint postgis.Point

	for idx, taxiMovement := range simulator.TaxiMovements {
		_, err := db.Exec("INSERT INTO taxi_routes VALUES ($1, $2, $3, $4, $5, $6, $7, $8, "+
			"$9, $10, $11, $12, $13, $14, $15, $16, ST_LineFromEncodedPolyline($17))",
			idx, taxiMovement.PuTime, taxiMovement.DoTime, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1,
			taxiMovement.Geometry)
		if err != nil {
			panic(err)
		}
	}

	// Demonstrate both driver.Valuer and sql.Scanner support
	db.QueryRow("SELECT ST_GeomFromText('POINT(-71.064544 42.28787)');").Scan(&newPoint)

	fmt.Println(point)
	fmt.Println(newPoint)
	if point == newPoint {
		fmt.Println("Point returned equal from PostGIS!")
	}
}

// Main function of the processing program.
func main() {
	conf := readConfig("config.json")
	fmt.Println(conf)

	simulator := taxisim.SetUpSimulation(conf.NumTaxis)
	simulator = processTaxiDataCSV(conf.TaxiData[0], simulator, processTaxiRecord)

	fmt.Println(simulator)

	writeSimulatorOutputToDatabase(simulator)
}
