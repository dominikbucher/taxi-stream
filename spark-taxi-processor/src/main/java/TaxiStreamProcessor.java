import com.google.gson.Gson;
import com.vividsolutions.jts.geom.Coordinate;
import com.vividsolutions.jts.geom.GeometryFactory;
import com.vividsolutions.jts.geom.Point;
import org.apache.spark.SparkConf;
import org.apache.spark.api.java.JavaRDD;
import org.apache.spark.api.java.JavaSparkContext;
import org.apache.spark.sql.Row;
import org.apache.spark.sql.SQLContext;
import org.apache.spark.sql.SparkSession;
import org.apache.spark.sql.types.StructType;
import org.apache.spark.storage.StorageLevel;
import org.apache.spark.streaming.Seconds;
import org.apache.spark.streaming.api.java.JavaReceiverInputDStream;
import org.apache.spark.streaming.api.java.JavaStreamingContext;
import org.datasyslab.geospark.enums.FileDataSplitter;
import org.datasyslab.geospark.enums.GridType;
import org.datasyslab.geospark.enums.IndexType;
import org.datasyslab.geospark.spatialOperator.KNNQuery;
import org.datasyslab.geospark.spatialRDD.PointRDD;
import org.json4s.jackson.Json;

import java.util.ArrayList;
import java.util.List;

import static org.apache.spark.sql.types.DataTypes.*;

public class TaxiStreamProcessor {
    public class TaxiUpdate {
        public double lon;
        public double lat;
        public int taxiId;

        public TaxiUpdate() {

        }
    }

    public static void main(String[] args) throws InterruptedException, Exception {
        System.setProperty("hadoop.home.dir", "C:\\Programs\\Hadoop-adds");
        System.setProperty("spark.sql.warehouse.dir", "file:///C:/spark-warehouse");

        //SparkConf conf = new SparkConf().setMaster("local[*]").setAppName("TaxiStreamProcessor");
        SparkSession spark = SparkSession.builder().appName("TaxiStreamProcessor").master("local[*]").getOrCreate();
        SQLContext sqlContext = spark.sqlContext();
        JavaSparkContext jsc = JavaSparkContext.fromSparkContext(spark.sparkContext()); //new JavaSparkContext(spark);
        JavaStreamingContext ssc = new JavaStreamingContext(jsc, Seconds.apply(1));

        GeometryFactory geometryFactory = new GeometryFactory();
        //Gson gson = new Gson();

        JavaReceiverInputDStream<String> stream = ssc.receiverStream(new WebSocketReceiver(StorageLevel.MEMORY_ONLY()));

        // From https://github.com/SiefSeif/GeoSpark/blob/patch-1/Using%20GeoSpark%20with%20Spark%20Streaming.md.
        String pointRDDInputLocation = "data/clients.csv";
        int pointRDDOffset = 0; // The point long/lat starts from Column 0
        FileDataSplitter pointRDDSplitter = FileDataSplitter.CSV;
        PointRDD clientRequests = new PointRDD(jsc, pointRDDInputLocation, pointRDDOffset, pointRDDSplitter, true);

        clientRequests.analyze();
        clientRequests.spatialPartitioning(GridType.QUADTREE);
        clientRequests.buildIndex(IndexType.RTREE, false);

        stream.foreachRDD((JavaRDD<String> rdd) -> {
            //JavaRDD<Point> taxiPoints =
            rdd.collect().forEach(record -> {
                System.out.println(record);
                try {

                    //TaxiUpdate taxiUpdate = gson.fromJson(record, TaxiUpdate.class);
                    StructType schema = (new StructType()).add("lon", DoubleType).add("lat", DoubleType).add("taxiId", IntegerType);
                    List<String> records = new ArrayList<>();
                    records.add(record);
                    Row r = sqlContext.read().schema(schema).json(jsc.parallelize(records)).first();
                    double lon = r.getDouble(0);
                    double lat = r.getDouble(1);
                    int taxiId = r.getInt(2);

                    System.out.println(lon);

                    List<Point> clients = KNNQuery.SpatialKnnQuery(clientRequests, geometryFactory
                            .createPoint(new Coordinate(lon, lat)), 1, true);
                    System.out.println(taxiId + ": " + clients);

                    //return geometryFactory.createPoint(new Coordinate(lon, lat));
                } catch (Exception e) {
                    System.out.println("Error: " + e);
                    e.printStackTrace();
                    //return null;
                }
            });

            /*if (!rdd.isEmpty()) {
                PointRDD taxiPointsRDD = new PointRDD(taxiPoints, StorageLevel.NONE());
                taxiPointsRDD.spatialPartitioning(GridType.QUADTREE);

                taxiPointsRDD.buildIndex(IndexType.RTREE, true);
                taxiPointsRDD.indexedRDD.persist(StorageLevel.MEMORY_ONLY());
                clientRequests.spatialPartitionedRDD.persist(StorageLevel.MEMORY_ONLY());
            }*/
        });


        //stream.print();
        ssc.start();
        ssc.awaitTermination();
        ssc.stop();
    }
}
