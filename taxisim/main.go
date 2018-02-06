package taxisim

import (
	"fmt"
	"encoding/csv"
	"os"
	"io"
	"time"
	"strconv"
	"database/sql"

	_ "github.com/lib/pq"
	"taxistream/base"
)

// Wraps the CSV processing functionality.
func processTaxiDataCSV(filename string, maxRoutes int32, simulator Simulator,
	processRow func([]string, Simulator) Simulator) Simulator {

	file, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.Comma = ','
	lineCount := int32(0)
	reader.Read()

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			fmt.Println("Error (reading CSV record):", err)
			return simulator
		}
		simulator = processRow(record, simulator)

		lineCount += 1
		if maxRoutes != -1 {
			if lineCount > maxRoutes {
				return simulator
			}
		}
	}
	return simulator
}

// Processes a single taxi record.
func processTaxiRecord(record []string, simulator Simulator) Simulator {
	// fmt.Println("Record", lineCount, "is", record, "and has", len(record), "fields")
	puTime, _ := time.Parse("2006-01-02 15:04:05", record[1])
	puLon, _ := strconv.ParseFloat(record[5], 32)
	puLat, _ := strconv.ParseFloat(record[6], 32)
	doTime, _ := time.Parse("2006-01-02 15:04:05", record[2])
	doLon, _ := strconv.ParseFloat(record[7], 32)
	doLat, _ := strconv.ParseFloat(record[8], 32)

	passengerCount, _ := strconv.ParseInt(record[9], 10, 32)
	fareAmount, _ := strconv.ParseFloat(record[11], 32)
	extra, _ := strconv.ParseFloat(record[12], 32)
	mtaTax, _ := strconv.ParseFloat(record[13], 32)
	tipAmount, _ := strconv.ParseFloat(record[14], 32)
	tollsAmount, _ := strconv.ParseFloat(record[15], 32)
	ehailFee, _ := strconv.ParseFloat(record[16], 32)
	improvementSurcharge, _ := strconv.ParseFloat(record[17], 32)
	totalAmount, _ := strconv.ParseFloat(record[18], 32)
	paymentType, _ := strconv.ParseInt(record[19], 10, 32)
	tripType, _ := strconv.ParseInt(record[20], 10, 32)

	simulator = processRoute(puTime, puLon, puLat, doTime, doLon, doLat, int32(passengerCount), fareAmount, extra,
		mtaTax, tipAmount, tollsAmount, ehailFee, improvementSurcharge, totalAmount, int32(paymentType),
		int32(tripType), simulator)
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
func writeSimulatorOutputToDatabase(simulator Simulator) {
	db := connectToDatabase()
	defer db.Close()

	for idx, taxiMovement := range simulator.TaxiMovements {
		_, err := db.Exec("INSERT INTO taxi_routes VALUES ($1, $2, $3, $4, $5, $6, $7, $8, "+
			"$9, $10, $11, $12, $13, $14, $15, $16, $17, ST_LineFromEncodedPolyline($18))",
			idx, taxiMovement.TaxiId, taxiMovement.PuTime, taxiMovement.DoTime, taxiMovement.PassengerCount,
			taxiMovement.TripDistance, taxiMovement.TripDuration, taxiMovement.FareAmount, taxiMovement.Extra,
			taxiMovement.MTATax, taxiMovement.TipAmount, taxiMovement.TollsAmount, taxiMovement.EhailFee,
			taxiMovement.ImprovementSurcharge, taxiMovement.TotalAmount, taxiMovement.PaymentType,
			taxiMovement.TripType, taxiMovement.Geometry)
		if err != nil {
			panic(err)
		}
	}
}

// Runs the simulation, based on a configuration file.
func RunSim(conf base.Configuration) {
	simulator := setUpSimulation(conf.NumTaxis)
	simulator = processTaxiDataCSV(conf.TaxiData[0], conf.MaxRoutes, simulator, processTaxiRecord)

	fmt.Println(simulator)
	fmt.Println("Unresolved routes:", simulator.UnresolvedRoutes)

	writeSimulatorOutputToDatabase(simulator)
}
