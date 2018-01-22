package main

import (
	"fmt"
	"encoding/csv"
	"os"
	"io"
	"time"
	"strconv"
	"taxistream/osrm"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {
	file, err := os.Open("data/green_tripdata_2016-01.csv")
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
			return
		}

		// fmt.Println("Record", lineCount, "is", record, "and has", len(record), "fields")
		puTime, _ := time.Parse("2006-01-02 15:04:05", record[1])
		puLon, _ := strconv.ParseFloat(record[5], 32)
		puLat, _ := strconv.ParseFloat(record[6], 32)
		doLon, _ := strconv.ParseFloat(record[7], 32)
		doLat, _ := strconv.ParseFloat(record[8], 32)
		resp := osrm.QueryOSRM(puTime, puLon, puLat, doLon, doLat)
		fmt.Println(resp)

		lineCount += 1

		if lineCount > 0 {
			return
		}
	}
}
