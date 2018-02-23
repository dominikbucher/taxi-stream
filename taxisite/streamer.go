package taxisite

import (
	"container/ring"
	"taxistream/base"
	"github.com/gorilla/websocket"
	"time"
	"fmt"
)

type Streamer struct {
	WebsocketChannel  *websocket.Conn
	TaxiupdateChannel *chan []byte
	ChannelUpdates    *ring.Ring
}

func ringAverage(ring *ring.Ring) float64 {
	avg := 0.0
	els := 0.0
	ring.Do(func(p interface{}) {
		if r, ok := p.(float64); ok {
			avg += r
			els += 1.0
		}
	})
	return avg / els
}

func setUpStreamer(conf base.Configuration) *Streamer {
	taxiupdates := make(chan []byte, int32(conf.TargetSpeedPerSecond*conf.TrackpointPrepWindowSize*1.1))
	channelUpdates := ring.New(1000)
	streamer := Streamer{nil, &taxiupdates, channelUpdates}

	throughput := conf.TargetSpeedPerSecond
	backoff := 1000000000.0 / conf.TargetSpeedPerSecond
	lastSent := time.Now()
	statsCounter := 0
	reset := true

	go func() {
		for {
			u := <-taxiupdates
			if streamer.WebsocketChannel != nil {
				streamer.WebsocketChannel.WriteMessage(websocket.TextMessage, u)
			}
			if reset {
				lastSent = time.Now()
				reset = false
			}
			streamer.ChannelUpdates = streamer.ChannelUpdates.Next()
			streamer.ChannelUpdates.Value = float64(time.Now().Sub(lastSent).Nanoseconds())
			lastSent = time.Now()
			//fmt.Println("Sent 1:", throughput, backoff, len(taxiupdates))

			statsCounter += 1
			if (statsCounter % 1000) == 0 {
				fmt.Println("Sent 1000:", ringAverage(streamer.ChannelUpdates), throughput, backoff, len(taxiupdates))
			}
			timePerMessage := ringAverage(streamer.ChannelUpdates)
			throughput = 1000000000.0 / timePerMessage
			processingTime := timePerMessage - backoff
			targetProcTime := processingTime * conf.TargetSpeedPerSecond
			backoff = (1000000000 - targetProcTime) / conf.TargetSpeedPerSecond

			//scale := throughput / conf.TargetSpeedPerSecond
			//backoff = backoff * scale
			//fmt.Println(ringAverage(streamer.ChannelUpdates), throughput, backoff, len(taxiupdates))
			/*
			if ring1SecAgo, ok := streamer.ChannelUpdates.Next().Value.(time.Time); ok {
				timeDiff := math.Max(1, float64(time.Now().Sub(ring1SecAgo).Nanoseconds()))
				throughput = 1000000000.0 / timeDiff * conf.TargetSpeedPerSecond
				backoff = backoff + 1000000000.0/(throughput-conf.TargetSpeedPerSecond)
			}*/
			// If we exhaust the channel, reset to the target speed.
			if len(taxiupdates) == 0 {
				backoff = 1000000000.0 / conf.TargetSpeedPerSecond
				streamer.ChannelUpdates = ring.New(1000)
				reset = true
			}
			if backoff > 0 {
				time.Sleep(time.Duration(backoff) * time.Nanosecond)
			}
		}
	}()
	return &streamer
}
