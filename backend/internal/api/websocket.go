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

func handleCloseError(err error, whenMessage string) (isClosed bool) {
	if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
		fmt.Printf("websocket closed during %s: %s", whenMessage, err)
		return true
	} else {
		fmt.Printf("error %s: %v\n", whenMessage, err)
		return false
	}
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

		// ping pong
		pongWait := 5 * time.Second
		pingPeriod := (pongWait * 9) * 10
		go func() {
			ws.SetReadDeadline(time.Now().Add(pongWait))
			ws.SetPongHandler(func(string) error {
				ws.SetReadDeadline(time.Now().Add(pongWait))
				return nil
			})
			for {
				if _, _, err := ws.NextReader(); err != nil {
					ws.Close()
					break
				}
			}
		}()

		writeWait := 3 * time.Second
		go func() {
			for {
				time.Sleep(pingPeriod)

				muWriter.Lock()
				ws.SetWriteDeadline(time.Now().Add(writeWait))
				err := ws.WriteMessage(websocket.PingMessage, nil)
				muWriter.Unlock()

				if err != nil {
					closed := handleCloseError(err, "pinging websocket")
					if closed {
						return
					}
				}
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
				closed := handleCloseError(err, "writing geolocation to websocket")
				if closed {
					return err
				}
				return nil
			}
			return nil
		})
		if err != nil {
			fmt.Printf("error listening to geolocation inserted: %v\n", err)
			return
		}
	}
}
