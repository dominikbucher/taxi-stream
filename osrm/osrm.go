package osrm

import (
	"time"
	"strconv"
	"fmt"
	"net/http"
	"encoding/json"
)

type OSRMResponse struct {
	Code      string
	Routes    []OSRMRoute
	Waypoints []OSRMWaypoint
}

type OSRMRoute struct {
	Distance    float32
	Duration    float32
	Geometry    string
	Legs        []OSRMLeg
	Weight      float32
	Weight_name string
}

type OSRMLeg struct {
	Distance float32
	Duration float32
	Steps    []OSRMStep
	Summary  string
	Weight   float32
}

type OSRMStep struct {
	Distance float32
	Duration float32
	Geometry string
	Mode     string
	Name     string
	Weight   float32
}

type OSRMWaypoint struct {
	Hint     string
	Location []float32
	Name     string
}

func QueryOSRM(puTime time.Time, puLon, puLat, doLon, doLat float64) string {
	url := "http://ikgoeco.ethz.ch/osrm/route/v1/nyccar/" +
		strconv.FormatFloat(puLon, 'f', 10, 64) + "," +
		strconv.FormatFloat(puLat, 'f', 10, 64) + ";" +
		strconv.FormatFloat(doLon, 'f', 10, 64) + "," +
		strconv.FormatFloat(doLat, 'f', 10, 64) + "?steps=true&overview=full"
	fmt.Println(url)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Error:", err)
		return ""
	}
	// defer resp.Body.Close()
	// body, err := ioutil.ReadAll(resp.Body)
	// bodyString := string(body)
	if err != nil {
		fmt.Println("Error:", err)
		return ""
	}
	j := new(OSRMResponse)
	json.NewDecoder(resp.Body).Decode(&j)
	fmt.Println(j)
	return ""
}
