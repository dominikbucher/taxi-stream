import org.apache.spark.SparkConf;
import org.apache.spark.storage.StorageLevel;
import org.apache.spark.streaming.Seconds;
import org.apache.spark.streaming.api.java.JavaReceiverInputDStream;
import org.apache.spark.streaming.api.java.JavaStreamingContext;

public class TaxiStreamProcessor {
    public static void main(String[] args) throws InterruptedException {
        System.setProperty("hadoop.home.dir", "C:\\Programs\\Hadoop-adds");

        SparkConf conf = new SparkConf().setMaster("local[*]").setAppName("TaxiStreamProcessor");
        JavaStreamingContext ssc = new JavaStreamingContext(conf, Seconds.apply(1));
        JavaReceiverInputDStream stream = ssc.receiverStream(new WebSocketReceiver(StorageLevel.MEMORY_ONLY()));

        stream.print();
        ssc.start();
        ssc.awaitTermination();
        ssc.stop();
    }
}
