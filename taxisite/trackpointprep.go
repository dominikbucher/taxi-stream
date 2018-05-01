package taxisite

import (
	"time"
	"taxistream/base"
	"database/sql"
	"fmt"
	"github.com/lib/pq"
	"taxistream/taxisim"
	"github.com/twpayne/go-polyline"
	"encoding/json"
	"math/rand"
)

// The trackpoint preparation component constantly retrieves routes from a database,
// and generates taxi updates from it.
//
// As for now, this only supports location and occupancy updates. Later, things like
// prices, ratings, destinations, fuel and motor status, and also non-taxi events
// such as transport requests, congestion updates, etc. might be added.
type TrackpointPrepper struct {
	WindowStart   time.Time
	WindowEnd     time.Time
	Routes        []Route
	ReservedTaxis map[int32]bool
}

// A route as it is stored in the database.
type Route struct {
	Id     int64
	TaxiId int32
	PuTime time.Time
	DoTime time.Time

	PassengerCount       int32
	Distance             float64
	Duration             float64
	FareAmount           float64
	Extra                float64
	MTATax               float64
	TipAmount            float64
	TollsAmount          float64
	EHailFee             float64
	ImprovementSurcharge float64
	TotalAmount          float64
	PaymentType          int32
	TripType             int32

	Geometry string

	StartLon float64
	StartLat float64
	EndLon   float64
	EndLat   float64
}

// A taxi location update that is serialized as JSON and sent to interested parties.
type TaxiUpdate struct {
	TaxiId int32   `json:"taxiId"`
	Lon    float64 `json:"lon"`
	Lat    float64 `json:"lat"`

	NumOccupants   int32    `json:"numOccupants"`
	DestLon        *float64 `json:"destLon"`
	DestLat        *float64 `json:"destLat"`
	ReservationLon *float64 `json:"reservationLon"`
	ReservationLat *float64 `json:"reservationLat"`
}

// A taxi occupancy update.
type TaxiOccupancyUpdate struct {
	TaxiId       int32   `json:"taxiId"`
	NumOccupants int32   `json:"numOccupants"`
	DestLon      float64 `json:"destLon"`
	DestLat      float64 `json:"destLat"`
}

// Updates where a taxi will travel to (when it gets booked).
type TaxiDestinationUpdate struct {
	TaxiId       int32   `json:"taxiId"`
	NumOccupants int32   `json:"numOccupants"`
	DestLon      float64 `json:"destLon"`
	DestLat      float64 `json:"destLat"`
}

// Update when a taxi receives a reservation.
type TaxiReservationUpdate struct {
	TaxiId         int32   `json:"taxiId"`
	ReservationLon float64 `json:"reservationLon"`
	ReservationLat float64 `json:"reservationLat"`
}

// When a taxi finishes serving a route, its price is sent.
type TaxiRouteCompletedUpdate struct {
	TaxiId               int32   `json:"taxiId"`
	PassengerCount       int32   `json:"passengerCount"`
	Distance             float64 `json:"tripDistance"`
	Duration             float64 `json:"tripDuration"`
	FareAmount           float64 `json:"fareAmount"`
	Extra                float64 `json:"extra"`
	MTATax               float64 `json:"mtaTax"`
	TipAmount            float64 `json:"tipAmount"`
	TollsAmount          float64 `json:"tollsAmount"`
	EHailFee             float64 `json:"ehailFee"`
	ImprovementSurcharge float64 `json:"improvementSurcharge"`
	TotalAmount          float64 `json:"totalAmount"`
	PaymentType          int32   `json:"paymentType"`
	TripType             int32   `json:"tripType"`
}

// Sets up the connection to the database.
func connectToDatabase(conf base.Configuration) *sql.DB {
	dataSourceName := fmt.Sprintf("host=%s port=%s dbname=%s sslmode=%s user=%s password=%s",
		conf.DbHost, conf.DbPort, conf.DbName, conf.DbSSLMode, conf.DbUser, conf.DbPassword)
	db, err := sql.Open("postgres", dataSourceName)
	if err != nil {
		panic(err)
	}
	err = db.Ping()
	if err != nil {
		panic(err)
	}
	return db
}

// Gets the simulated routes from PostGIS.
func getRoutes(windowStart time.Time, windowEnd time.Time, ids []int64, db *sql.DB) []Route {
	rows, err := db.Query("SELECT id, taxi_id, pickup_time, dropoff_time, passenger_count, "+
		"trip_distance, trip_duration, fare_amount, extra, mta_tax, tip_amount, tolls_amount, ehail_fee, "+
		"improvement_surcharge, total_amount, payment_type, trip_type, ST_AsEncodedPolyline(geometry), "+
		"ST_X(ST_StartPoint(geometry)), ST_Y(ST_StartPoint(geometry)), "+
		"ST_X(ST_EndPoint(geometry)), ST_Y(ST_EndPoint(geometry)) "+
		"FROM taxi_routes WHERE dropoff_time > $1 AND pickup_time < $2 AND id <> ALL ($3)",
		windowStart, windowEnd, pq.Int64Array(ids))
	defer rows.Close()
	if err != nil {
		fmt.Println("Error (with query):", err)
		panic(err)
	}

	routes := make([]Route, 0)
	for rows.Next() {
		route := Route{}
		err := rows.Scan(&route.Id, &route.TaxiId, &route.PuTime, &route.DoTime, &route.PassengerCount,
			&route.Distance, &route.Duration, &route.FareAmount, &route.Extra, &route.MTATax, &route.TipAmount,
			&route.TollsAmount, &route.EHailFee, &route.ImprovementSurcharge, &route.TotalAmount,
			&route.PaymentType, &route.TripType, &route.Geometry, &route.StartLon, &route.StartLat,
			&route.EndLon, &route.EndLat)
		if err != nil {
			fmt.Println("Error (parsing route data):", err)
		}
		routes = append(routes, route)
	}
	err = rows.Err()
	if err != nil {
		panic(err)
	}
	return routes
}

// Gets routes from a database, and transforms them into the appropriate number of taxi update
// messages. These are then sent to the streamer component of the application.
func prepTrackpoints(trackpointPrepper *TrackpointPrepper, streamer *Streamer, db *sql.DB, conf base.Configuration) {
	fmt.Println("TrackpointPrepper:", trackpointPrepper.WindowStart, "-", trackpointPrepper.WindowEnd)
	windowSize := conf.TrackpointPrepWindowSize
	timeWarp := conf.TimeWarp
	targetSpeed := conf.TargetSpeedPerSecond

	// Get all currently active routes.
	ids := make([]int64, 0)
	for _, r := range trackpointPrepper.Routes {
		ids = append(ids, r.Id)
	}
	routes := getRoutes(trackpointPrepper.WindowStart, trackpointPrepper.WindowEnd, ids, db)

	// Get new set of active routes.
	trackpointPrepper.Routes = append(trackpointPrepper.Routes, routes...)
	newRoutes := make([]Route, 0)
	for _, r := range trackpointPrepper.Routes {
		if !r.DoTime.Before(trackpointPrepper.WindowStart) {
			newRoutes = append(newRoutes, r)
		}
	}

	// Update everything to contain the final set of routes and make ready for next iteration.
	trackpointPrepper.Routes = newRoutes
	fmt.Println("TrackpointPrepper.Routes.len:", len(trackpointPrepper.Routes))

	if len(trackpointPrepper.Routes) > int(conf.NumTaxis / 10) {
		// Create updates for all taxis. First, compute how many updates we need to reach the target speed.
		numUpdates := windowSize * targetSpeed
		numTimeSlices := numUpdates / float64(len(trackpointPrepper.Routes))
		timeInc := time.Duration(1000000000.0*windowSize*timeWarp/numTimeSlices) * time.Nanosecond

		timeSlice := trackpointPrepper.WindowStart
		updates := make([][]byte, 0)
		for timeSlice.Before(trackpointPrepper.WindowEnd) {
			sliceEnd := timeSlice.Add(timeInc)

			for _, r := range trackpointPrepper.Routes {
				// Check if this route just started now. If so, we have to create an occupancy message.
				// If it's a route with passengers, a destination message has to be added too.
				if r.PuTime.After(timeSlice) && r.PuTime.Before(sliceEnd) {
					// This is a new route, we have to generate an occupancy message.
					// Since we include all messages in both streams, here we kinda redundantly send both messages.
					o, _ := json.Marshal(TaxiOccupancyUpdate{r.TaxiId, r.PassengerCount,
						r.EndLon, r.EndLat})
					updates = append(updates, o)

					b, _ := json.Marshal(TaxiDestinationUpdate{r.TaxiId, r.PassengerCount,
						r.EndLon, r.EndLat})
					updates = append(updates, b)
				}

				// Check if this route is just stopping now. If so, we have to send the journey (esp. price) information.
				if r.DoTime.After(timeSlice) && r.DoTime.Before(sliceEnd) {
					b, _ := json.Marshal(TaxiRouteCompletedUpdate{r.TaxiId, r.PassengerCount,
						r.Distance, r.Duration, r.FareAmount, r.Extra,
						r.MTATax, r.TipAmount, r.TollsAmount, r.EHailFee,
						r.ImprovementSurcharge, r.TotalAmount, r.PaymentType,
						r.TripType})
					updates = append(updates, b)
					delete(trackpointPrepper.ReservedTaxis, r.TaxiId)
				}

				// In some rare cases, the taxi gets ordered to the pickup location (let's say by a reservation call).
				// Optimally, the simulator would already generate these events...
				// For now, we do this approx. for one taxi every 10 seconds.
				if r.PassengerCount == 0 && rand.Float64() < 1.0 /
					(10000000000.0/float64(timeInc.Nanoseconds())*float64(len(trackpointPrepper.Routes))) {
					trackpointPrepper.ReservedTaxis[r.TaxiId] = true
					b, _ := json.Marshal(TaxiReservationUpdate{r.TaxiId, r.EndLon, r.EndLat})
					updates = append(updates, b)
				}

				// In any case, we want to generate some location updates.
				// TODO Auf UNIX / Mac scheint es anders kodiert zu sein, d.h. das strings Replace ist nicht nötig.
				// TODO Auf Ubuntu geht es so (gleich wie Windows).
				coords, _, err := polyline.DecodeCoords([]byte(r.Geometry))
				// coords, _, err := polyline.DecodeCoords([]byte(strings.Replace(r.Geometry, "\\\\", "\\", -1)))
				if err != nil {
					panic(err)
				}
				perc := timeSlice.Sub(r.PuTime).Seconds() / r.DoTime.Sub(r.PuTime).Seconds()
				if perc > 0 && perc < 1 {
					lon, lat := taxisim.AlongPolyline(taxisim.PolylineLength(coords)*perc, coords)
					if streamer.TaxiupdateChannel != nil {
						var resLon *float64
						var resLat *float64
						if trackpointPrepper.ReservedTaxis[r.TaxiId] {
							resLon = &r.EndLon
							resLat = &r.EndLat
						}
						if r.PassengerCount > 0 {
							b, _ := json.Marshal(TaxiUpdate{r.TaxiId, lon, lat,
								r.PassengerCount, &r.EndLon, &r.EndLat, resLon, resLat})
							updates = append(updates, b)
						} else {
							b, _ := json.Marshal(TaxiUpdate{r.TaxiId, lon, lat,
								r.PassengerCount, nil, nil, resLon, resLat})
							updates = append(updates, b)
						}
					}
				}
			}
			timeSlice = timeSlice.Add(timeInc)
		}
		// Because some routes are not within the time slices, there are not enough updates. We fill in the missing ones
		// by repeating some.
		missingUpdates := int(numUpdates) - len(updates)
		updateCount := float64(len(updates)) / float64(missingUpdates)
		cnt := 0.0
		totCnt := 0
		for _, r := range updates {
			*streamer.TaxiupdateChannel <- r
			totCnt += 1
			if updateCount > 0 && cnt > updateCount {
				*streamer.TaxiupdateChannel <- r
				totCnt += 1
				cnt -= updateCount
			}

			cnt += 1
		}
		fmt.Println("Added messages", totCnt)

		trackpointPrepper.WindowStart = trackpointPrepper.WindowStart.Add(time.Second * time.Duration(windowSize*timeWarp))
		trackpointPrepper.WindowEnd = trackpointPrepper.WindowEnd.Add(time.Second * time.Duration(windowSize*timeWarp))
	} else {
		trackpointPrepper.WindowStart = time.Date(2016, time.January, 1, 0, 29, 20, 0, time.UTC)
		trackpointPrepper.WindowEnd = time.Date(2016, time.January, 1, 0, 29, int(20+windowSize*conf.TimeWarp), 0, time.UTC)
	}
}

// Sets up the trackpoint preparation component.
func setUpTrackpointPrep(conf base.Configuration, streamer Streamer) {
	db := connectToDatabase(conf)
	windowSize := conf.TrackpointPrepWindowSize
	trackpointPrepper := TrackpointPrepper{
		time.Date(2016, time.January, 1, 0, 29, 20, 0, time.UTC),
		time.Date(2016, time.January, 1, 0, 29, int(20+windowSize*conf.TimeWarp), 0, time.UTC),
		make([]Route, 0), make(map[int32]bool)}

	ticker := time.NewTicker(time.Duration(windowSize) * time.Second)
	quit := make(chan struct{})
	prepTrackpoints(&trackpointPrepper, &streamer, db, conf)

	go func() {
		for {
			select {
			case <-ticker.C:
				prepTrackpoints(&trackpointPrepper, &streamer, db, conf)
			case <-quit:
				ticker.Stop()
				db.Close()
				return
			}
		}
	}()
}
