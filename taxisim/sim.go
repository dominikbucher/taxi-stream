package taxisim

import (
	"errors"
	"math/rand"
	"fmt"
	"time"
)

// Defines the current simulator state.
type Simulator struct {
	Taxis            []Taxi
	TaxiMovements    []TaxiMovement
	TotalRoutes      int64
	UnresolvedRoutes int64
}

// Given a new route, select a taxi that could serve it.
func findTaxi(taxis []Taxi, puTime time.Time, puLon float64, puLat float64) (*Taxi, error) {
	// Of all the taxis, find one that is free at the moment.
	candidates := make([]*Taxi, 0)
	for idx, taxi := range taxis {
		if taxi.Status == inits || (taxi.Status == free && canReach(taxi, puTime, puLon, puLat)) {
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
func canReach(taxi Taxi, puTime time.Time, puLon float64, puLat float64) bool {
	return puTime.After(taxi.Time) &&
		(Distance(taxi.Lon, taxi.Lat, puLon, puLat) < puTime.Sub(taxi.Time).Seconds()*TaxiSpeed)
}

// Creates a random taxi movement and updates the taxi to the newest location.
func createRandomTaxiMovement(taxi Taxi) TaxiMovement {
	randLon := rand.Float64() * 0.2
	randLat := rand.Float64() * 0.2
	drivingRoute, err := resolveRoute(taxi.Time, taxi.Lon, taxi.Lat, randLon, randLat)
	if err != nil {
		panic(err)
	}
	taxi.Lon = drivingRoute.DoLon
	taxi.Lat = drivingRoute.DoLat
	taxi.Time = drivingRoute.DoTime
	return TaxiMovement{taxi.Id, taxi.Time, drivingRoute.DoTime, free, 0,
		drivingRoute.Distance, drivingRoute.DoTime.Sub(drivingRoute.PuTime).Seconds(),
		0, 0, 0, 0, 0, 0, 0, 0,
		-1, -1, drivingRoute.Geometry}
}

// Sets up the simulation.
func setUpSimulation(numTaxis int32) Simulator {
	taxis := make([]Taxi, numTaxis)
	for i := range taxis {
		taxis[i].Id = int32(i)
		taxis[i].Status = inits
	}
	taxiMovements := make([]TaxiMovement, 0)
	return Simulator{taxis, taxiMovements, 0, 0}
}

// Processes a single route and integrates it into the simulator.
func processRoute(puTime time.Time, puLon float64, puLat float64, doTime time.Time, doLon float64, doLat float64,
	passengerCount int32, fareAmount float64, extra float64, mtaTax float64, tipAmount float64,
	tollsAmount float64, ehailFee float64, improvementSurcharge float64, totalAmount float64, paymentType int32,
	tripType int32, simulator Simulator) Simulator {

	simulator.TotalRoutes += 1

	taxi, err := findTaxi(simulator.Taxis, puTime, puLon, puLat)
	if err != nil {
		fmt.Println("Error (no taxi found to process route):", err)
		simulator.UnresolvedRoutes += 1
		return simulator
	}

	route, err := resolveRoute(puTime, puLon, puLat, doLon, doLat)
	if err != nil {
		fmt.Println("Error (unable to resolve route):", err)
		simulator.UnresolvedRoutes += 1
		return simulator
	}

	if taxi.Status != inits {
		drivingRoute, err := resolveRoute(taxi.Time, taxi.Lon, taxi.Lat, route.PuLon, route.PuLat)
		if err != nil {
			panic(err)
		}
		timeBudget := taxi.Time.Sub(route.PuTime).Seconds()
		drivingDurationHigh := drivingRoute.Distance / TaxiSpeed * 1.1

		// If the time to drive from the current position to the pickup location of the taxi is very large,
		// we simply let this taxi drive around a bit.
		for timeBudget > drivingDurationHigh {
			simulator.TaxiMovements = append(simulator.TaxiMovements, createRandomTaxiMovement(*taxi))

			drivingRoute, err := resolveRoute(taxi.Time, taxi.Lon, taxi.Lat, route.PuLon, route.PuLat)
			if err != nil {
				panic(err)
			}
			timeBudget = taxi.Time.Sub(route.PuTime).Seconds()
			drivingDurationHigh = drivingRoute.Distance / TaxiSpeed * 1.1
		}

		// Once it is close enough, route to the route pickup location.
		simulator.TaxiMovements = append(simulator.TaxiMovements,
			TaxiMovement{taxi.Id, taxi.Time, drivingRoute.DoTime, free, 0,
				drivingRoute.Distance, drivingRoute.DoTime.Sub(drivingRoute.PuTime).Seconds(),
				0, 0, 0, 0, 0, 0, 0, 0,
				-1, -1, drivingRoute.Geometry})
	}

	// Finally, write the real route back to the simulator, and update all taxi variables.
	simulator.TaxiMovements = append(simulator.TaxiMovements,
		TaxiMovement{taxi.Id, route.PuTime, route.DoTime, occupied,
			passengerCount, route.Distance, doTime.Sub(puTime).Seconds(),
			fareAmount, extra, mtaTax, tipAmount, tollsAmount,
			ehailFee, improvementSurcharge, totalAmount, paymentType,
			tripType, route.Geometry})
	taxi.Status = free
	taxi.Time = route.DoTime
	taxi.Lon = route.DoLon
	taxi.Lat = route.DoLat

	return simulator
}
