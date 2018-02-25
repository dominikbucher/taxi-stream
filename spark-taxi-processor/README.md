# Spark Stream Processor for Taxi Data

This is a self-contained Apache Spark (https://spark.apache.org) application for handling taxi data.
It automatically hooks up to a WebSocket hosted at ws://localhost:8080/ws and simply prints the incoming messages.
It uses the old Spark DStream API. Maybe the newer streaming API would be better.
Spark Streaming is using a batched approach, in this case a batch window of 1 second was chosen.
This comes from Spark's background as a purely Big Data processing platform (Big Data in the sense of MapReduce offline processing).

## Running 
As this is a self-contained program (Spark runs embedded and not on its own cluster), there might be a problem with `winutils` on Windows. 
To resolve, put `winutils.exe` from https://github.com/srccodes/hadoop-common-2.2.0-bin/archive/master.zip into a folder of your choosing and set the Hadoop path respectively:

```java
System.setProperty("hadoop.home.dir", "c:\\winutil\\");
```

`winutils.exe` must reside in `c:\winutil\bin` in this case.