package taxisite

import (
	"taxistream/base"
	"github.com/gorilla/websocket"
	"fmt"
)

type Streamer struct {
	WebsocketChannel  *websocket.Conn
	TaxiupdateChannel chan TaxiUpdate
}

type TaxiUpdate struct {
	TaxiId int32   `json:"taxiId"`
	Lon    float64 `json:"lon"`
	Lat    float64 `json:"lat"`
}

func setUpStreamer(conf base.Configuration) *Streamer {
	taxiupdates := make(chan TaxiUpdate, int32(conf.TargetSpeedPerSecond)+1)
	streamer := Streamer{nil, taxiupdates}
	go func() {
		for {
			u := <-taxiupdates
			fmt.Println("Gotten taxi update.")
			if streamer.WebsocketChannel != nil {
				fmt.Println("Sending taxi update via channel.")
				streamer.WebsocketChannel.WriteJSON(u)
			}
		}
	}()
	return &streamer
}
