package osrm

import (
	"time"
	"strconv"
	"fmt"
	"net/http"
	"encoding/json"
	"errors"
	"github.com/twpayne/go-polyline"
)

// Here, we assume an average taxi speed of 2.222 m/s.
var TaxiSpeed = 2.222

// The response of an OSRM request.
// This wraps concrete parts of the route.
type OSRMResponse struct {
	Code      string
	Routes    []OSRMRoute
	Waypoints []OSRMWaypoint
}

// A single route as given by OSRM.
// Different routes could be presented to a user as different possible choices.
type OSRMRoute struct {
	Distance    float32
	Duration    float32
	Geometry    string
	Legs        []OSRMLeg
	Weight      float32
	Weight_name string
}

// A leg of a route is simply a part of it.
type OSRMLeg struct {
	Distance float32
	Duration float32
	Steps    []OSRMStep
	Summary  string
	Weight   float32
}

// A step is even a smaller unit within a route.
type OSRMStep struct {
	Distance float32
	Duration float32
	Geometry string
	Mode     string
	Name     string
	Weight   float32
}

// A waypoint is a single coordinate on a route.
type OSRMWaypoint struct {
	Hint     string
	Location []float32
	Name     string
}

// Defines a route as used within this application.
type Route struct {
	PuLon    float64
	PuLat    float64
	PuTime   time.Time
	DoLon    float64
	DoLat    float64
	DoTime   time.Time
	Distance float64
	Geometry string
}

// Queries ikgoeco (running OSRM) for a route from (puLon, puLat) to (doLon, doLat).
func QueryOSRM(puTime time.Time, puLon, puLat, doLon, doLat float64) (*Route, error) {
	url := "http://ikgoeco.ethz.ch/osrm/route/v1/nyccar/" +
		strconv.FormatFloat(puLon, 'f', 10, 64) + "," +
		strconv.FormatFloat(puLat, 'f', 10, 64) + ";" +
		strconv.FormatFloat(doLon, 'f', 10, 64) + "," +
		strconv.FormatFloat(doLat, 'f', 10, 64) + "?steps=true&overview=full"
	fmt.Println(url)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Error:", err)
		return nil, err
	}
	j := new(OSRMResponse)
	json.NewDecoder(resp.Body).Decode(&j)
	if len(j.Routes) == 0 {
		return nil, errors.New("no routes found")
	}
	decodedCoords, _, _ := polyline.DecodeCoords([]byte(j.Routes[0].Geometry))
	return &Route{decodedCoords[0][1], decodedCoords[0][0], puTime,
		decodedCoords[len(decodedCoords)-1][1], decodedCoords[len(decodedCoords)-1][0],
		puTime.Add(time.Second * time.Duration(float64(j.Routes[0].Distance)/TaxiSpeed)),
		float64(j.Routes[0].Distance), j.Routes[0].Geometry}, nil
}
