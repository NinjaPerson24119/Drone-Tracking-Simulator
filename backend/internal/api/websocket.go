package api

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/NinjaPerson24119/MapProject/backend/internal/constants"
	"github.com/NinjaPerson24119/MapProject/backend/internal/database"
	"github.com/NinjaPerson24119/MapProject/backend/internal/filters"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type GeolocationsWebSocketMessage struct {
	Geolocations []*database.DeviceGeolocation `json:"geolocations"`
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func geolocationsWebSocketGenerator(repo database.Repo) func(c *gin.Context) {
	return func(c *gin.Context) {
		ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer ws.Close()
		var wsClosed atomic.Bool
		wsClosed.Store(false)

		fmt.Print("websocket connection opened\n")

		// it is safe to have one reader and one writer concurrently
		muWriter := sync.Mutex{}
		writeWait := 3 * time.Second

		// ping pong
		pongWait := 10 * time.Second
		pingPeriod := (pongWait * 9) / 10
		if pingPeriod >= pongWait {
			fmt.Printf("ping period is greater than pong wait: %v >= %v\n", pingPeriod, pongWait)
			c.Status(http.StatusInternalServerError)
			return
		}
		go func() {
			ws.SetReadDeadline(time.Now().Add(pongWait))
			ws.SetPongHandler(func(string) error {
				ws.SetReadDeadline(time.Now().Add(pongWait))
				return nil
			})
			for {
				messageType, bytes, err := ws.ReadMessage()
				if err != nil {
					fmt.Printf("error reading from websocket: %v\n", err)
					ws.Close()
					break
				}
				if messageType == websocket.TextMessage {
					if string(bytes) == "ping" {
						muWriter.Lock()
						ws.SetWriteDeadline(time.Now().Add(writeWait))
						err := ws.WriteMessage(websocket.TextMessage, []byte("pong"))
						muWriter.Unlock()

						if err != nil {
							fmt.Printf("error writing user-level pong message to websocket: %v\n", err)
							break
						}
					}
				}
			}
			fmt.Print("propagated close over atomic bool")
			wsClosed.Store(true)
		}()
		go func() {
			for {
				muWriter.Lock()
				ws.SetWriteDeadline(time.Now().Add(writeWait))
				err := ws.WriteMessage(websocket.PingMessage, nil)
				muWriter.Unlock()

				if err != nil {
					fmt.Printf("error writing ping message to websocket: %v\n", err)
					break
				}
				time.Sleep(pingPeriod)
			}
			wsClosed.Store(true)
		}()

		// buffer geolocations
		muFlaggedDeviceIDs := sync.Mutex{}
		flaggedDeviceIDs := map[string]bool{}
		bufferSize := constants.SimulatedDevices
		bufferPeriod := time.Second / 2
		timeAtLastSend := time.Now()
		checkPeriod := time.Millisecond * 10
		go func() {
			for {
				fmt.Print("checking for flagged geolocations\n")
				if wsClosed.Load() {
					fmt.Print("websocket closed while processing flagged geolocations")
					return
				}

				// wait until there are enough flagged geolocations or enough time has passed
				muFlaggedDeviceIDs.Lock()
				flaggedDeviceIDsLength := len(flaggedDeviceIDs)
				muFlaggedDeviceIDs.Unlock()
				if (flaggedDeviceIDsLength < bufferSize && time.Since(timeAtLastSend) < bufferPeriod) || flaggedDeviceIDsLength == 0 {
					fmt.Printf("not enough flagged geolocations: %v < %v, and not enough time elapsed: %v < %v\n", flaggedDeviceIDsLength, bufferSize, time.Since(timeAtLastSend), bufferPeriod)
					continue
				}

				// get flagged geolocations as a list of IDs, then clear the flag
				geolocationIDs := []string{}
				muFlaggedDeviceIDs.Lock()
				for k := range flaggedDeviceIDs {
					geolocationIDs = append(geolocationIDs, k)
				}
				flaggedDeviceIDs = map[string]bool{}
				muFlaggedDeviceIDs.Unlock()
				fmt.Printf("got %v flagged geolocations and reset flags\n", len(geolocationIDs))

				// get flagged geolocations from the database
				geolocations, err := repo.GetMultiLatestGeolocations(c.Request.Context(), geolocationIDs)
				if err != nil {
					fmt.Printf("error getting flagged geolocations: %v\n", err)
					return
				}
				// not found geolocations will be returned from GetMulti as nil
				geolocationsWithoutNil := []*database.DeviceGeolocation{}
				for _, g := range geolocations {
					if g != nil {
						geolocationsWithoutNil = append(geolocationsWithoutNil, g)
					} else {
						fmt.Printf("geolocation not found after notification: %v\n", g)
					}
				}
				if len(geolocationsWithoutNil) == 0 {
					fmt.Printf("getmulti only returned nil geolocations\n")
					continue
				}

				// send flagged geolocations to the websocket
				muWriter.Lock()
				ws.SetWriteDeadline(time.Now().Add(writeWait))
				json := GeolocationsWebSocketMessage{
					Geolocations: geolocationsWithoutNil,
				}
				err = ws.WriteJSON(json)
				muWriter.Unlock()
				if err != nil {
					fmt.Printf("error writing json to websocket: %v\n", err)
					return
				}
				timeAtLastSend = time.Now()
				fmt.Printf("sent %v geolocations to websocket\n", len(geolocations))

				// avoid hammering the locks
				time.Sleep(checkPeriod)
			}
		}()

		// begin connection by sending all geolocations
		fmt.Print("sending complete geolocations update to websocket\n")
		geolocations, err := getLatestGeolocations(c.Request.Context(), repo)
		if err != nil {
			c.Status(http.StatusInternalServerError)
			return
		}
		json := GeolocationsWebSocketMessage{
			Geolocations: geolocations,
		}
		muWriter.Lock()
		ws.SetWriteDeadline(time.Now().Add(writeWait))
		err = ws.WriteJSON(json)
		muWriter.Unlock()
		if err != nil {
			fmt.Printf("error writing json to websocket: %v\n", err)
			return
		}
		fmt.Print("sent complete geolocations update to websocket\n")

		// listen to updates and send new geolocations as they occur
		err = repo.ListenToGeolocationInserted(c.Request.Context(), func(deviceID string) error {
			muFlaggedDeviceIDs.Lock()
			flaggedDeviceIDs[deviceID] = true
			muFlaggedDeviceIDs.Unlock()
			fmt.Printf("flagged geolocation inserted: %v\n", deviceID)

			if wsClosed.Load() {
				return fmt.Errorf("websocket closed while handling geolocation inserted")
			}
			return nil
		})
		if err != nil {
			fmt.Printf("error listening to geolocation inserted: %v\n", err)
			return
		}
		fmt.Print("websocket connection closed\n")
		wsClosed.Store(true)
	}
}

func getLatestGeolocations(ctx context.Context, repo database.Repo) ([]*database.DeviceGeolocation, error) {
	geolocations := []*database.DeviceGeolocation{}
	page := 1
	for {
		fmt.Printf("getting latest geolocations page %v\n", page)
		geolocationsPage, err := repo.ListLatestGeolocations(ctx, filters.PageOptions{
			Page:     page,
			PageSize: 100,
		})
		if err != nil {
			return nil, fmt.Errorf("error getting latest geolocations: %v\n", err)
		}
		if len(geolocationsPage) == 0 {
			break
		}
		geolocations = append(geolocations, geolocationsPage...)
		if len(geolocations) >= constants.SimulatedDevices {
			break
		}
		page++
	}
	return geolocations[:constants.SimulatedDevices], nil
}
