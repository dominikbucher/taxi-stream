package ch.ethz.gis;

import org.apache.flink.configuration.Configuration;
import org.apache.flink.streaming.api.functions.source.RichSourceFunction;
import org.java_websocket.client.WebSocketClient;
import org.java_websocket.handshake.ServerHandshake;

import java.net.URI;
import java.net.URISyntaxException;
import java.nio.ByteBuffer;

public class WebSocketSource extends RichSourceFunction<String> {
    String url = "ws://localhost:8080/ws";

    private volatile boolean run = false;

    SourceContext<String> context;
    private WebSocketClient client = null;

    @Override
    public void open(Configuration parameters) {
        try {
            client = new EmptyClient(new URI(url));
            run = true;
        } catch (URISyntaxException e) {
            e.printStackTrace();
        }
    }

    @Override
    public void cancel() {
        run = false;
        client.close();
    }

    @Override
    public void run(SourceContext<String> sourceContext) throws Exception {
        this.context = sourceContext;
        client.connect();

        while (run) {
            Thread.sleep(1);
        }
    }

    public class EmptyClient extends WebSocketClient {
        public EmptyClient(URI serverURI) {
            super(serverURI);
        }

        @Override
        public void onOpen(ServerHandshake handshakedata) {
            send("Hello, it is me. Mario :)");
            System.out.println("new connection opened");
        }

        @Override
        public void onClose(int code, String reason, boolean remote) {
            System.out.println("closed with exit code " + code + " additional info: " + reason);
        }

        @Override
        public void onMessage(String message) {
            //System.out.println("received message: " + message);
            context.collect(message);
        }

        @Override
        public void onMessage(ByteBuffer message) {
            System.out.println("received ByteBuffer");
        }

        @Override
        public void onError(Exception ex) {
            System.err.println("an error occurred:" + ex);
        }
    }
}