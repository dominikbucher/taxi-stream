# Flink Stream Processor for Taxi Data

This is a self-contained Apache Flink (https://flink.apache.org) application for handling taxi data.
It automatically hooks up to a WebSocket hosted at ws://localhost:8080/ws and simply prints the incoming messages.
In contrary to Apache Spark, Flink uses a purely streaming approach, i.e., by default nothing is windowed or batched. 

It might be better to use Apache Kafka (https://kafka.apache.org) instead, as Flink basically adds some guarantees and another API on top of Kafka. 
This is not quite clear yet.