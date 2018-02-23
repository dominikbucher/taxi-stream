# New York City Taxistream Provider

This program uses taxi data from New York to create (half-simulated) data streams of taxis moving throughout New York. The taxi data is available from http://www.nyc.gov/html/tlc/html/about/trip_record_data.shtml. The idea is to evaluate methods that are required for stream processing mobility data with the intent of providing people with transport. For example, taxis (or also autonomous cars, buses, etc.) constantly send location updates and if they are free or not. This data has to be cleaned, processed, and probably / partially stored in order to make it queryable for people looking for transport.


## Program Structure

The program in this repository consists of three parts:
* Generation of taxi data (which is later used for streaming) based on the above dataset.
* Streaming of generated taxi data.
* A template for a stream consumer written in Java. This is later going to be removed from this repository.


## Setup

Both the taxi data generator as well as the streaming component are written in Go. After installing and setting up Go as described on https://golang.org, install `dep` (https://github.com/golang/dep) to pull in all the dependencies. Using `dep ensure` (I guess...), you can install all required dependencies. 

Secondly, you need a PostGIS (https://postgis.net) database installed on your system (this of course reuqires Postgres as well). 

In the `config.json` file, you can specify all important parameters, such as the database users, database name, number of taxis to be simulated, etc. Importantly, the `mode` parameter (either `process` or `stream`) determines if the application builds a simulated dataset, or serves this as a data stream on port 8080.

Using `go build main.go` you can finally compile and run the program. Note that for building the dataset, you need to be in the ETH network, as access to the OSRM instance running on ikgoeco.ethz.ch is restricted to the ETH network. 


## Base Data and Taxi Route Generation

The taxi data is available as CSV files, containing for example:

| VendorID | lpep_pickup_datetime | Lpep_dropoff_datetime | Store_and_fwd_flag | RateCodeID |
| ----- | ----- | ----- | ----- | ----- |
| 2 | 2016-01-01 00:29:24 | 2016-01-01 00:39:36 | N | 1 |

| Pickup_longitude | Pickup_latitude | Dropoff_longitude | Dropoff_latitude |
| ----- | ----- | ----- | ----- |
| -73.928642272949219 | 40.680610656738281 | -73.924278259277344 | 40.698043823242188 |

| Passenger_count | Trip_distance | Fare_amount | Extra | MTA_tax | Tip_amount | Tolls_amount | Ehail_fee |
| ----- | ----- | ----- | ----- | ----- | ----- | ----- | ----- |
| 1 | 1.46 | 8 | 0.5 | 0.5 | 1.86 | 0 | 0 |

| improvement_surcharge | Total_amount | Payment_type | Trip_type |
| ----- | ----- | ----- | ----- |
| 0.3 | 11.16 | 1 | 1 |

From the pickup and dropoff locations, a route is computed using the Open Source Routing Machine (www.project-osrm.org). It is simply assumed that taxis have a uniform speed on any routes (for now - we might add speed depending on the road type later).

The dataset has several drawbacks:
* No taxi IDs are given, i.e., we don't know which taxi serves which route. 
* Only routes with passengers are recorded in the dataset. It is not known what taxis do between these routes, nor if they randomly pick up a passenger or drive somewhere on purpose.
* The taxi datasets are available for yellow cabs (NYC), green cabs (surrounding suburbian areas), and FHW vehicles (FHVs). It is not known how many of these taxis are on the streets at any given time.

To circumvent these problems, we have to use a model to generate the missing data. To improve the model, we additionally use data from the 2014 NYC taxicab factbook (http://www.nyc.gov/html/tlc/downloads/pdf/2014_taxicab_fact_book.pdf). For example, we have a typical pattern of "taxis on the road". Note that this pattern is most likely extracted from the same or a similar dataset as we are using, but we are still going to use it to determine how many taxis we simulate with our model. In particular, we assume the following number of taxis on the road at each hour of the day (yellow taxis, for green taxis multiply by around 1.4, as there are approx. 13'200 yellow taxis, and 18'000 green taxis), starting at midnight:

| 00:00 | 01:00 | 02:00 | 03:00 | 04:00 | 05:00 | 06:00 | 07:00 | 08:00 | 09:00 | 10:00 | 11:00 |
| ----- | ----- | ----- | ----- | ----- | ----- | ----- | ----- | ----- | ----- | ----- | ----- |
| 6500 | 5325 | 4150 | 2975 | 1800 | 3340 | 4880 | 6420 | 7960 | 9500 | 9167 | 9833 |

| 12:00 | 13:00 | 14:00 | 15:00 | 16:00 | 17:00 | 18:00 | 19:00 | 20:00 | 21:00 | 22:00 | 23:00 |
| ----- | ----- | ----- | ----- | ----- | ----- | ----- | ----- | ----- | ----- | ----- | ----- |
| 10000 | 9500 | 8767 | 8033 | 7300 | 9150 | 11000 | 10500 | 10000 | 9500 | 9000 | 7750 |

To generate the taxi routes, we apply the following method:

1. For each route, we check if there is any taxi that is either in init state (i.e., has never served a customer before) or is free and can reach the pickup location before the pickup time.
2. We randomly choose one of these taxis (if no taxi is available, this route is simply skipped), and route to the pickup location. In case the taxi would arrive way too early, we let it cruise around randomly for a while. 
3. All the routes are entered into a PostGIS database, both the one with customers from the CSV file, as well as the one driving to the pickup location (or potential "idle cruising" routes).

### Notes About Data

* It seems the taxi data **does not contain** lat/lon after June 2016. So probably better to use data from before, as this is a hypothetical example anyways.
* Other interesting repositories and blogs:
  * https://github.com/toddwschneider/nyc-taxi-data
  * http://minimaxir.com/2015/11/nyc-ggplot2-howto/


## Stream Generation

The stream is now created by periodically taking a set of routes from the PostGIS database, and transforming them into individual trackpoints / stream packets. This means that for example every 15 seconds all routes somehow intersecting with the *next 15 second window* are pulled from the database, and transformed into a number of smaller location packets. These location packets are stored in a queue.

A second part of the program simply constantly pipes out location and other packets from the queue. 
 
## Known Simulator Problems

* Taxi movements do not line up, i.e., sometimes a taxi arrives later than it starts from a certain point. This can be resolved by actually routing to the pickup location, and checking if it's feasible. Otherwise, one can choose another candidate. This will increase running time though.
* Taxis do not necessarily stay in vicinity. I saw a taxi that happily drove back to Manhattan from the airport, even though realistically, it would probably wait for a pickup at the airport. Maybe we could introduce a random waiting period?
* Sometimes taxis will get ordered to go somewhere (I imagine quite frequently). They are not able to pick up someone else during this time. This can be modeled by simply randomly make them drive to a pickup location on order, i.e., by not being free during this time. 

## Analysis and Visualization

To visualize the taxis with e.g. https://github.com/anitagraser/TimeManager, use the following script:
```sql
CREATE TABLE IF NOT EXISTS interpolated_taxi_routes (
	taxi_id integer,
    ts timestamp without time zone,
    geom Geometry
);
INSERT INTO interpolated_taxi_routes
(
    WITH start AS 
      (SELECT id, taxi_id, geometry, pickup_time, dropoff_time FROM taxi_routes), 
    intervals AS 
      (SELECT generate_series (0, 20) as steps)
    SELECT  
         taxi_id,
         pickup_time + INTERVAL '1 MINUTES' * steps AS ts,
         ST_Line_Interpolate_Point(geometry, steps/(SELECT count(steps)::float-1 FROM intervals)) AS geom
    FROM start, intervals
);
```
