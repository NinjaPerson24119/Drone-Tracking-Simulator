package api

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/NinjaPerson24119/MapProject/backend/internal/database"
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
		writeWait := 10 * time.Second

		// ping pong
		pongWait := 60 * time.Second
		pingPeriod := (pongWait * 9) * 10
		go func() {
			ws.SetReadDeadline(time.Now().Add(pongWait))
			ws.SetPongHandler(func(string) error {
				fmt.Print("pong\n")
				ws.SetReadDeadline(time.Now().Add(pongWait))
				return nil
			})
			for {
				if _, _, err := ws.ReadMessage(); err != nil {
					fmt.Printf("error reading next reader from websocket: %v\n", err)
					ws.Close()
					break
				}
			}
		}()
		go func() {
			for {
				time.Sleep(pingPeriod)

				muWriter.Lock()
				ws.SetWriteDeadline(time.Now().Add(writeWait))
				err := ws.WriteMessage(websocket.PingMessage, nil)
				muWriter.Unlock()

				if err != nil {
					fmt.Printf("error writing ping message to websocket: %v\n", err)
					break
				}
				fmt.Print("ping\n")
			}
		}()

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
