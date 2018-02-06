package osrm

import (
	"strconv"
	"fmt"
	"net/http"
	"encoding/json"
	"errors"
)

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

// Queries ikgoeco (running OSRM) for a route from (puLon, puLat) to (doLon, doLat).
func QueryOSRM(puLon, puLat, doLon, doLat float64) (*OSRMResponse, error) {
	url := "http://ikgoeco.ethz.ch/osrm/route/v1/nyccar/" +
		strconv.FormatFloat(puLon, 'f', 10, 64) + "," +
		strconv.FormatFloat(puLat, 'f', 10, 64) + ";" +
		strconv.FormatFloat(doLon, 'f', 10, 64) + "," +
		strconv.FormatFloat(doLat, 'f', 10, 64) + "?steps=true&overview=full"
	fmt.Println(url)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	osrmResp := new(OSRMResponse)
	json.NewDecoder(resp.Body).Decode(&osrmResp)
	if len(osrmResp.Routes) == 0 {
		return nil, errors.New("no routes found")
	}
	return osrmResp, nil
}
