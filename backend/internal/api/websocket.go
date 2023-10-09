package api

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

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
		}()

		// begin connection by sending all geolocations
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

		// listen to updates and send new geolocations as they occur
		err = repo.ListenToGeolocationInserted(c.Request.Context(), func(geolocation *database.DeviceGeolocation) error {
			json := GeolocationsWebSocketMessage{
				Geolocations: []*database.DeviceGeolocation{geolocation},
			}

			muWriter.Lock()
			ws.SetWriteDeadline(time.Now().Add(writeWait))
			err = ws.WriteJSON(json)
			muWriter.Unlock()

			if err != nil {
				fmt.Printf("error writing json to websocket: %v\n", err)
				return err
			}
			return nil
		})
		if err != nil {
			fmt.Printf("error listening to geolocation inserted: %v\n", err)
			return
		}
	}
}

func getLatestGeolocations(ctx context.Context, repo database.Repo) ([]*database.DeviceGeolocation, error) {
	geolocations := []*database.DeviceGeolocation{}
	page := 1
	for {
		geolocationsPage, err := repo.GetLatestGeolocations(ctx, filters.PageOptions{
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
	}
	return geolocations, nil
}
