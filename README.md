# Taxistream Generator

This program uses taxi data from New York to create (half-simulated) data streams.

## Various

* It seems the taxi data **does not contain** lat/lon after June 2016. So probably better to use data from before, as this
is a hypothetical example anyways.
* Other interesting repositories and blogs:
  * https://github.com/toddwschneider/nyc-taxi-data
  * http://minimaxir.com/2015/11/nyc-ggplot2-howto/
  
## Simulator Problems

* Taxi movements do not line up, i.e., sometimes a taxi arrives later than it starts from a certain point.
This can be resolved by actually routing to the pickup location, and checking if it's feasible.
Otherwise, one can choose another candidate. This will increase running time though.
* Taxis do not necessarily stay in vicinity. I saw a taxi that happily drove back to Manhattan from the airport, 
even though realistically, it would probably wait for a pickup at the airport. Maybe we could introduce a random waiting period?
* Sometimes taxis will get ordered to go somewhere (I imagine quite frequently). They are not able to pick up
someone else during this time. This can be modeled by simply randomly make them drive to a pickup location on order, i.e.,
by not being free during this time. 