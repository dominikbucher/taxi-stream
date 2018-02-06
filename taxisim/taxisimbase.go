package taxisim

import (
	"time"
	"github.com/twpayne/go-polyline"
	"taxistream/osrm"
)

// Here, we assume an average taxi speed of 2.222 m/s.
var TaxiSpeed = 2.222

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

// The different statuses a taxi can be in
type TaxiStatus int

const (
	inits    TaxiStatus = iota
	free     TaxiStatus = iota
	occupied TaxiStatus = iota
)

// Defines the location of a taxi at a given time.
type Taxi struct {
	Id     int32
	Time   time.Time
	Status TaxiStatus
	Lon    float64
	Lat    float64
}

// The movement of a taxi - this might be a route where the taxi carries a person, or not.
type TaxiMovement struct {
	TaxiId               int32
	PuTime               time.Time
	DoTime               time.Time
	Status               TaxiStatus
	PassengerCount       int32
	TripDistance         float64
	TripDuration         float64
	FareAmount           float64
	Extra                float64
	MTATax               float64
	TipAmount            float64
	TollsAmount          float64
	EhailFee             float64
	ImprovementSurcharge float64
	TotalAmount          float64
	PaymentType          int32
	TripType             int32
	Geometry             string
}

func ResolveRoute(puTime time.Time, puLon, puLat, doLon, doLat float64) (*Route, error) {
	route, err := osrm.QueryOSRM(puLon, puLat, doLon, doLat)
	if err != nil {
		return nil, err
	}
	decodedCoords, _, err := polyline.DecodeCoords([]byte(route.Routes[0].Geometry))
	if err != nil {
		return nil, err
	}
	return &Route{decodedCoords[0][1], decodedCoords[0][0], puTime,
		decodedCoords[len(decodedCoords)-1][1], decodedCoords[len(decodedCoords)-1][0],
		puTime.Add(time.Second * time.Duration(float64(route.Routes[0].Distance)/TaxiSpeed)),
		float64(route.Routes[0].Distance), route.Routes[0].Geometry}, nil
}
