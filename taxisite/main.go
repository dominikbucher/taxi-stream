package taxisite

import (
	"net/http"
	"github.com/gorilla/websocket"
	"taxistream/base"
	"log"
	"time"
	"encoding/json"
	"math/rand"
	"strconv"
)

var streamer *Streamer = nil
var clientRequestStreamer *ClientRequestStreamer = nil

type ClientRequestStreamer struct {
	WebsocketChannel     map[*websocket.Conn]bool
	MaxClients           int
	ClientRequestsPerSec float64
}

// Exposes some endpoints to interact with the streaming application.
func ExposeEndpoints(conf base.Configuration) {
	streamer = setUpStreamer(conf)
	setUpTrackpointPrep(conf, *streamer)

	clientRequestStreamer = &ClientRequestStreamer{make(map[*websocket.Conn]bool, 0),
		conf.MaxClients, conf.ClientRequestsPerSec}

	http.Handle("/", http.FileServer(http.Dir("./taxisite/static")))
	http.HandleFunc("/ws", wsHandler)
	http.HandleFunc("/ws-clients", wsHandlerClients)
	http.ListenAndServe(":" + strconv.Itoa(conf.WebSocketPort), nil)
}

// Upgrades a connection to WebSockets.
func wsHandler(w http.ResponseWriter, r *http.Request) {
	// We removed this to make connecting from Spark easier.
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

// Upgrades a connection to WebSockets, in this case for clients.
func wsHandlerClients(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Upgrade(w, r, w.Header(), 1024, 1024)
	if err != nil {
		http.Error(w, "Could not open websocket connection", http.StatusBadRequest)
	}

	// Here, we simply generate client requests occasionally.
	clientRequestStreamer.WebsocketChannel[conn] = true
	log.Println("Serving", len(clientRequestStreamer.WebsocketChannel), "client sockets.")
	go handleWsClients(conn)
	if len(clientRequestStreamer.WebsocketChannel) == 1 {
		go writeOccasionalClientRequest(clientRequestStreamer)
	} // Otherwise this already runs.
}

// Handles a websocket, in particular, closes it after the client goes offline.
func handleWsClients(conn *websocket.Conn) {
	defer conn.Close()
	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsCloseError(err) || websocket.IsUnexpectedCloseError(err) {
				delete(clientRequestStreamer.WebsocketChannel, conn)
				log.Println("Serving", len(clientRequestStreamer.WebsocketChannel), "client sockets.")
			} else {
				log.Println("Error when reading from WebSocket channel.")
				delete(clientRequestStreamer.WebsocketChannel, conn)
				log.Println("Serving", len(clientRequestStreamer.WebsocketChannel), "client sockets.")
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

type ClientRequestUpdate struct {
	ClientId  int     `json:"clientId"`
	OrigLon   float64 `json:"origLon"`
	OrigLat   float64 `json:"origLat"`
	DestLon   float64 `json:"destLon"`
	DestLat   float64 `json:"destLat"`
	WillShare bool    `json:"willShare"`
}

// Random boolean generator.
func randbool() bool {
	return rand.Float32() < 0.5
}

// Generates a random latitude in New York.
func randlat() float64 {
	return 40.61 + rand.Float64()*0.21
}

// Generates a random longitude in New York.
func randlon() float64 {
	return -74.02 + rand.Float64()*0.26
}

// Occasionally writes a client request on the WebSocket.
func writeOccasionalClientRequest(clientRequestStreamer *ClientRequestStreamer) {
	for {
		if len(clientRequestStreamer.WebsocketChannel) > 0 {
			msg, _ := json.Marshal(ClientRequestUpdate{rand.Intn(clientRequestStreamer.MaxClients),
				randlon(), randlat(), randlon(), randlat(), randbool()})
			for c := range clientRequestStreamer.WebsocketChannel {
				c.WriteMessage(websocket.TextMessage, msg)
			}
		} else {
			return
		}
		time.Sleep(time.Duration(1000000000.0/clientRequestStreamer.ClientRequestsPerSec) * time.Nanosecond)
	}
}
