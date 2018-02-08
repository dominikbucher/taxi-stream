package taxisite

import (
	"net/http"
	"github.com/gorilla/websocket"
	"fmt"
	"strings"
	"taxistream/base"
)

type msg struct {
	Num int
}

var streamer *Streamer = nil

func ExposeEndpoints(conf base.Configuration) {
	streamer = setUpStreamer(conf)
	setUpTrackpointPrep(conf, *streamer)

	http.Handle("/", http.FileServer(http.Dir("./taxisite/static")))
	http.HandleFunc("/ws", wsHandler)
	http.ListenAndServe(":8080", nil)
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Println(r.Host)
	if r.Header.Get("Origin") != "http://"+r.Host && !strings.Contains(r.Host, "localhost") {
		http.Error(w, "Origin not allowed", 403)
		return
	}
	conn, err := websocket.Upgrade(w, r, w.Header(), 1024, 1024)
	if err != nil {
		http.Error(w, "Could not open websocket connection", http.StatusBadRequest)
	}

	streamer.WebsocketChannel = conn
	go echo(conn)
}

func echo(conn *websocket.Conn) {
	for {
		m := msg{}

		err := conn.ReadJSON(&m)
		if err != nil {
			fmt.Println("Error reading json.", err)
		}

		fmt.Printf("Got message: %#v\n", m)

		if err = conn.WriteJSON(m); err != nil {
			fmt.Println(err)
		}
	}
}
