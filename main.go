package main

import (
	"fmt"
	"encoding/csv"
	"os"
	"io"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {
	file, err := os.Open("data/yellow_tripdata_2017-01.csv")
	check(err)
	defer file.Close()

	reader := csv.NewReader(file)
	reader.Comma = ';'
	lineCount := 0

	for {
		record, error := reader.Read()
		if error == io.EOF {
			break
		} else if error != nil {
			fmt.Println("Error:", error)
			return
		}

		fmt.Println("Record", lineCount, "is", record, "and has", len(record), "fields")
		lineCount += 1
	}
}
