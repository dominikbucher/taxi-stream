package taxisite

import (
	"container/ring"
	"taxistream/base"
	"github.com/gorilla/websocket"
	"time"
	"fmt"
)

// The streamer simply takes the messages produced by the trackpoint preparation component
// and pushes them out to interested parties.
type Streamer struct {
	WebsocketChannel  map[*websocket.Conn]bool
	TaxiupdateChannel *chan []byte
	ChannelUpdates    *ring.Ring
}

// Computes the average value in a ring of float64 values.
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

// Sets up the streamer and lets it listen to potential updates on a channel.
func setUpStreamer(conf base.Configuration) *Streamer {
	websocketChannels := make(map[*websocket.Conn]bool, 0)
	taxiupdates := make(chan []byte, int32(conf.TargetSpeedPerSecond*conf.TrackpointPrepWindowSize*1.1))
	channelUpdates := ring.New(1000)
	streamer := Streamer{websocketChannels, &taxiupdates, channelUpdates}

	throughput := conf.TargetSpeedPerSecond
	backoff := 1000000000.0 / conf.TargetSpeedPerSecond
	lastSent := time.Now()
	statsCounter := 0
	reset := true

	go func() {
		for {
			u := <-taxiupdates
			if len(streamer.WebsocketChannel) > 0 {
				for c := range streamer.WebsocketChannel{
					c.WriteMessage(websocket.TextMessage, u)
				}
			}
			if reset {
				lastSent = time.Now()
				reset = false
			}
			streamer.ChannelUpdates = streamer.ChannelUpdates.Next()
			streamer.ChannelUpdates.Value = float64(time.Now().Sub(lastSent).Nanoseconds())
			lastSent = time.Now()

			statsCounter += 1
			if (statsCounter % 1000) == 0 {
				fmt.Println("Sent 1000:", ringAverage(streamer.ChannelUpdates), throughput, backoff, len(taxiupdates))
			}
			timePerMessage := ringAverage(streamer.ChannelUpdates)
			throughput = 1000000000.0 / timePerMessage
			processingTime := timePerMessage - backoff
			targetProcTime := processingTime * conf.TargetSpeedPerSecond
			backoff = (1000000000 - targetProcTime) / conf.TargetSpeedPerSecond

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
