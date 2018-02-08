# Taxistream Generator

This program uses taxi data from New York to create (half-simulated) data streams. The taxi data is available from http://www.nyc.gov/html/tlc/html/about/trip_record_data.shtml. The idea is to evaluate methods that are required for stream processing mobility data with the intent of providing people with transport. For example, taxis (or also autonomous cars, buses, etc.) constantly send location updates and if they are free or not. This data has to be cleaned, processed, and probably / partially stored in order to make it queryable for people looking for transport.

## Data

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

From the pickup and dropoff locations, a route is computed using the Open Source Routing Machine (www.project-osrm.org). As no taxi ids are given in these datasets, and it is not known where taxis drive between served routes, a simple model is applied to generate the data:

1. For each route, we check if there is any taxi that is either in init state (i.e., has never served a customer before) or is free and can reach the pickup location before the pickup time.
2. We randomly choose one of these taxis (if no taxi is available, this route is simply skipped), and route to the pickup location. In case the taxi would arrive way too early, we let it cruise around randomly for a while. 
3. All the routes are entered into a PostGIS database, both the one with customers from the CSV file, as well as the one driving to the pickup location (or potential "idle cruising" routes).

### Notes About Data

* It seems the taxi data **does not contain** lat/lon after June 2016. So probably better to use data from before, as this
is a hypothetical example anyways.
* Other interesting repositories and blogs:
  * https://github.com/toddwschneider/nyc-taxi-data
  * http://minimaxir.com/2015/11/nyc-ggplot2-howto/


## Stream Generation

The stream is now created by periodically taking a set of routes from the PostGIS database, and transforming them into individual trackpoints / stream packets. This means that for example every 15 seconds all routes somehow intersecting with the *next 15 second window* are pulled from the database, and transformed into a number of smaller location packets. These location packets are stored in a queue.

A second part of the program simply constantly pipes out location and other packets from the queue. 
 
## Known Simulator Problems

* Taxi movements do not line up, i.e., sometimes a taxi arrives later than it starts from a certain point.
This can be resolved by actually routing to the pickup location, and checking if it's feasible.
Otherwise, one can choose another candidate. This will increase running time though.
* Taxis do not necessarily stay in vicinity. I saw a taxi that happily drove back to Manhattan from the airport, 
even though realistically, it would probably wait for a pickup at the airport. Maybe we could introduce a random waiting period?
* Sometimes taxis will get ordered to go somewhere (I imagine quite frequently). They are not able to pick up
someone else during this time. This can be modeled by simply randomly make them drive to a pickup location on order, i.e.,
by not being free during this time. 
