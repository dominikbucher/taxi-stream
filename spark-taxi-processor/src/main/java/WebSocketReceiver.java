import org.apache.spark.storage.StorageLevel;
import org.apache.spark.streaming.receiver.Receiver;
import org.java_websocket.client.WebSocketClient;
import org.java_websocket.handshake.ServerHandshake;

import java.net.URI;
import java.net.URISyntaxException;
import java.nio.ByteBuffer;

public class WebSocketReceiver extends Receiver<String> implements Runnable {
    String url = "ws://129.132.127.249:8080/ws";

    private Thread thread = null;
    private WebSocketClient client = null;

    WebSocketReceiver(StorageLevel storageLevel) {
        super(storageLevel);
    }

    public void onStart() {
        try {
            client = new EmptyClient(new URI(url));
            thread = new Thread(this);
            thread.start();
        } catch (URISyntaxException e) {
            e.printStackTrace();
        }
    }

    public void onStop() {
        client.close();
        thread.interrupt();
    }

    public void run() {
        client.connect();
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
            store(message);
            //System.out.println("received message: " + message);
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