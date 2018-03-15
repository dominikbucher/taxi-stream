package taxisite

import (
	"net/http"
	"github.com/gorilla/websocket"
	"taxistream/base"
	"log"
)

var streamer *Streamer = nil

// Exposes some endpoints to interact with the streaming application.
func ExposeEndpoints(conf base.Configuration) {
	streamer = setUpStreamer(conf)
	setUpTrackpointPrep(conf, *streamer)

	http.Handle("/", http.FileServer(http.Dir("./taxisite/static")))
	http.HandleFunc("/ws", wsHandler)
	http.ListenAndServe(":8080", nil)
}

// Upgrades a connection to WebSockets.
func wsHandler(w http.ResponseWriter, r *http.Request) {
	/*if r.Header.Get("Origin") != "http://"+r.Host && !strings.Contains(r.Host, "localhost") {
		http.Error(w, "Origin not allowed", 403)
		return
	}*/
	conn, err := websocket.Upgrade(w, r, w.Header(), 1024, 1024)
	if err != nil {
		http.Error(w, "Could not open websocket connection", http.StatusBadRequest)
	}

	streamer.WebsocketChannel[conn] = true
	log.Println("Serving", len(streamer.WebsocketChannel), "sockets.")
	go handleWs(conn)
}

// Handles a websocket, in particular, closes it after the client goes offline.
func handleWs(conn *websocket.Conn) {
	defer conn.Close()
	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsCloseError(err) || websocket.IsUnexpectedCloseError(err) {
				delete(streamer.WebsocketChannel, conn)
				log.Println("Serving", len(streamer.WebsocketChannel), "sockets.")
			} else {
				log.Println("Error when reading from WebSocket channel.")
				delete(streamer.WebsocketChannel, conn)
				log.Println("Serving", len(streamer.WebsocketChannel), "sockets.")
			}
			log.Println(err)
			return
		}

		// Simply write the message back to the sender (pong).
		if err := conn.WriteMessage(messageType, p); err != nil {
			log.Println(err)
			return
		}
	}
}
