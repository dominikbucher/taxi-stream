package taxisite

import (
	"container/ring"
	"taxistream/base"
	"github.com/gorilla/websocket"
	"time"
	"fmt"
	"encoding/csv"
	"os"
	"strconv"
	"strings"
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
	taxiupdates := make(chan []byte, int32(conf.TargetSpeedPerSecond*conf.TrackpointPrepWindowSize*2))
	channelUpdates := ring.New(100)
	streamer := Streamer{websocketChannels, &taxiupdates, channelUpdates}

	throughput := conf.TargetSpeedPerSecond
	backoff := 30.0 //1000000000.0 / conf.TargetSpeedPerSecond
	lastSent := time.Now()
	statsCounter := 0
	reset := true

	go func() {
		time.Sleep(3 * time.Second)
		var writer *csv.Writer = nil
		if conf.Log {
			file, _ := os.Create("data/application-metrics.csv")
			defer file.Close()
			writer = csv.NewWriter(file)
			defer writer.Flush()
			line := []string{"throughput", "backoff", "timediff", "taxiupdates"}
			writer.Write(line)
		}

		for {
			u := <-taxiupdates
			if float64(len(taxiupdates)) > 0.95 * float64(cap(taxiupdates)) && strings.Contains(string(u), "\"lon\"") {
				// The channel is almost full... this is a hack to simply let off some steam.
				// TODO Remove this hack.
			} else {
				if len(streamer.WebsocketChannel) > 0 {
					for c := range streamer.WebsocketChannel {
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
				processingTime := streamer.ChannelUpdates.Value.(float64) - backoff
				targetProcTime := processingTime * conf.TargetSpeedPerSecond
				backoff = (1000000000 - targetProcTime) / conf.TargetSpeedPerSecond

				if conf.Log {
					line := []string{strconv.FormatFloat(throughput, 'f', 5, 64),
						strconv.FormatFloat(backoff, 'f', 5, 64),
						strconv.FormatFloat(streamer.ChannelUpdates.Value.(float64), 'f', 5, 64),
						strconv.Itoa(len(taxiupdates))}
					writer.Write(line)
				}

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
		}
	}()
	return &streamer
}
