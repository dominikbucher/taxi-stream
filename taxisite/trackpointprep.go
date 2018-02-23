package taxisite

import (
	"time"
	"taxistream/base"
	"database/sql"
	"fmt"
	"github.com/lib/pq"
	"taxistream/taxisim"
	"github.com/twpayne/go-polyline"
	"strings"
	"encoding/json"
)

type TrackpointPrepper struct {
	WindowStart time.Time
	WindowEnd   time.Time
	Routes      []Route
}

type Route struct {
	Id             int64
	TaxiId         int32
	PuTime         time.Time
	DoTime         time.Time
	PassengerCount int32
	Distance       float64
	Geometry       string
}

type TaxiUpdate struct {
	TaxiId int32   `json:"taxiId"`
	Lon    float64 `json:"lon"`
	Lat    float64 `json:"lat"`
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

// Takes the output of a simulation run and writes it to PostGIS.
func getRoutes(windowStart time.Time, windowEnd time.Time, ids []int64, db *sql.DB) []Route {
	rows, err := db.Query("SELECT id, taxi_id, pickup_time, dropoff_time, passenger_count, "+
		"trip_distance, ST_AsEncodedPolyline(geometry) "+
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
			&route.Distance, &route.Geometry)
		if err != nil {
			fmt.Println("Error (parsing route data):", err)
		}
		routes = append(routes, route)
	}
	err = rows.Err()
	if err != nil {
		panic(err)
	}
	fmt.Println("Routes:", routes)
	return routes
}

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

	// Create updates for all taxis. First, compute how many updates we need to reach the target speed.
	numUpdates := windowSize * targetSpeed
	numTimeSlices := numUpdates / float64(len(trackpointPrepper.Routes))
	timeInc := time.Duration(1000000000.0*windowSize*timeWarp/numTimeSlices) * time.Nanosecond
	fmt.Println("#Updates", numUpdates)
	fmt.Println("#Time Slices", numTimeSlices)
	fmt.Println("Time increase", timeInc)

	timeSlice := trackpointPrepper.WindowStart
	cnt := 0
	for timeSlice.Before(trackpointPrepper.WindowEnd) {
		for _, r := range trackpointPrepper.Routes {
			coords, _, err := polyline.DecodeCoords([]byte(strings.Replace(r.Geometry, "\\\\", "\\", -1)))
			if err != nil {
				panic(err)
			}
			perc := timeSlice.Sub(r.PuTime).Seconds() / r.DoTime.Sub(r.PuTime).Seconds()
			if perc > 0 && perc < 1 {
				lon, lat := taxisim.AlongPolyline(taxisim.PolylineLength(coords)*perc, coords)
				if streamer.TaxiupdateChannel != nil {
					// fmt.Print(".")
					//fmt.Println(perc, r.TaxiId, lon, lat)
					b, _ := json.Marshal(TaxiUpdate{r.TaxiId, lon, lat})
					//fmt.Println(string(b))
					*streamer.TaxiupdateChannel <- b
					cnt += 1
				}
			}
		}
		timeSlice = timeSlice.Add(timeInc)
	}
	fmt.Println("Added messages", cnt)
	trackpointPrepper.WindowStart = trackpointPrepper.WindowStart.Add(time.Second * time.Duration(windowSize*timeWarp))
	trackpointPrepper.WindowEnd = trackpointPrepper.WindowEnd.Add(time.Second * time.Duration(windowSize*timeWarp))
}

func setUpTrackpointPrep(conf base.Configuration, streamer Streamer) {
	db := connectToDatabase(conf)
	windowSize := conf.TrackpointPrepWindowSize
	trackpointPrepper := TrackpointPrepper{
		time.Date(2016, time.January, 1, 0, 29, 20, 0, time.UTC),
		time.Date(2016, time.January, 1, 0, 29, int(20+windowSize*conf.TimeWarp), 0, time.UTC),
		make([]Route, 0)}

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
