package taxisim

import (
	"time"
	"errors"
	"math/rand"
	"taxistream/osrm"
	"fmt"
)

// Defines the current simulator state.
type Simulator struct {
	Taxis         []Taxi
	TaxiMovements []TaxiMovement
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
	PuTime   time.Time
	DoTime   time.Time
	Status   TaxiStatus
	Geometry string
}

// Given a new route, select a taxi that could serve it.
func findTaxi(taxis []Taxi, route osrm.Route) (*Taxi, error) {
	// Of all the taxis, find one that is free at the moment.
	candidates := make([]*Taxi, 0)
	for idx, taxi := range taxis {
		if taxi.Status == inits || (taxi.Status == free && canReach(taxi, route)) {
			candidates = append(candidates, &taxis[idx])
		}
	}

	if len(candidates) == 0 {
		return nil, errors.New("no taxi candidates left")
	}

	choice := rand.Intn(len(candidates))
	return candidates[choice], nil
}

// Determines if a taxi could reach a given route (pickup location).
func canReach(taxi Taxi, route osrm.Route) bool {
	return route.PuTime.After(taxi.Time) &&
		(Distance(taxi.Lon, taxi.Lat, route.PuLon, route.PuLat) < taxi.Time.Sub(route.PuTime).Seconds()*osrm.TaxiSpeed)
}

// Creates a random taxi movement and updates the taxi to the newest location.
func createRandomTaxiMovement(taxi Taxi) TaxiMovement {
	randLon := rand.Float64() * 0.2
	randLat := rand.Float64() * 0.2
	drivingRoute, err := osrm.QueryOSRM(taxi.Time, taxi.Lon, taxi.Lat, randLon, randLat)
	if err != nil {
		panic(err)
	}
	taxi.Lon = drivingRoute.DoLon
	taxi.Lat = drivingRoute.DoLat
	taxi.Time = drivingRoute.DoTime
	return TaxiMovement{taxi.Time, drivingRoute.DoTime, free, drivingRoute.Geometry}
}

// Sets up the simulation.
func SetUpSimulation(numTaxis int32) Simulator {
	taxis := make([]Taxi, numTaxis)
	for i, taxi := range taxis {
		taxi.Id = int32(i)
		taxi.Status = inits
	}
	taxiMovements := make([]TaxiMovement, 0)
	return Simulator{taxis, taxiMovements}
}

// Processes a single route and integrates it into the simulator.
func ProcessRoute(route osrm.Route, simulator Simulator) Simulator {
	taxi, err := findTaxi(simulator.Taxis, route)
	if err != nil {
		fmt.Println(err)
		return simulator
	}

	if taxi.Status != inits {
		drivingRoute, err := osrm.QueryOSRM(taxi.Time, taxi.Lon, taxi.Lat, route.PuLon, route.PuLat)
		if err != nil {
			panic(err)
		}
		timeBudget := taxi.Time.Sub(route.PuTime).Seconds()
		drivingDurationHigh := drivingRoute.Distance / osrm.TaxiSpeed * 1.1

		// If the time to drive from the current position to the pickup location of the taxi is very large,
		// we simply let this taxi drive around a bit.
		for timeBudget > drivingDurationHigh {
			simulator.TaxiMovements = append(simulator.TaxiMovements, createRandomTaxiMovement(*taxi))

			drivingRoute, err := osrm.QueryOSRM(taxi.Time, taxi.Lon, taxi.Lat, route.PuLon, route.PuLat)
			if err != nil {
				panic(err)
			}
			timeBudget = taxi.Time.Sub(route.PuTime).Seconds()
			drivingDurationHigh = drivingRoute.Distance / osrm.TaxiSpeed * 1.1
		}

		// Once it is close enough, route to the route pickup location.
		simulator.TaxiMovements = append(simulator.TaxiMovements,
			TaxiMovement{taxi.Time, drivingRoute.DoTime, free, drivingRoute.Geometry})
	}

	// Finally, write the real route back to the simulator, and update all taxi variables.
	simulator.TaxiMovements = append(simulator.TaxiMovements,
		TaxiMovement{route.PuTime, route.DoTime, occupied, route.Geometry})
	taxi.Status = free
	taxi.Time = route.DoTime
	taxi.Lon = route.DoLon
	taxi.Lat = route.DoLat

	return simulator
}
