package main

import (
	"fmt"
	"os"
	"encoding/json"
	"taxistream/base"
	"taxistream/taxisim"
	"taxistream/taxisite"

	_ "github.com/lib/pq"
)

// Reads a configuration file.
// The file needs to adhere to the Configuration struct.
func readConfig(configFile string) base.Configuration {
	file, err1 := os.Open(configFile)
	if err1 != nil {
		panic(err1)
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	configuration := base.Configuration{}
	err2 := decoder.Decode(&configuration)
	if err2 != nil {
		panic(err2)
	}
	return configuration
}

// Main function of the processing program.
func main() {
	conf := readConfig("config.json")
	fmt.Println(conf)

	switch conf.Mode {
	case "process":
		taxisim.RunSim(conf)
	case "stream":
		taxisite.ExposeEndpoints(conf)
	default:
		fmt.Println("Please specify run 'mode' in config file: {'process', 'stream'}.")
	}
}
