package api

import (
	"fmt"
	"net/http"

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
	if closeErr, ok := err.(*websocket.CloseError); ok {
		fmt.Printf("websocket closed during %s: %s", whenMessage, closeErr.Error())
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

		// ping pong
		/*
			go func() {
				for {
					msgType, bytes, err := ws.ReadMessage()
					if err != nil {
						isClosed := handleCloseError(err, "reading ping from websocket")
						if isClosed {
							return
						}
						continue
					}
					if msgType != websocket.TextMessage {
						continue
					}
					if string(bytes) == "ping" {
						err = ws.WriteMessage(websocket.TextMessage, []byte("pong"))
						if err != nil {
							isClosed := handleCloseError(err, "writing pong to websocket")
							if isClosed {
								return
							}
							continue
						}
					}
				}
			}()
		*/

		// relay geolocation inserts
		err = repo.ListenToGeolocationInserted(c.Request.Context(), func(geolocation *database.DeviceGeolocation) error {
			json := GeolocationsWebSocketMessage{
				Geolocations: []*database.DeviceGeolocation{geolocation},
			}
			err = ws.WriteJSON(json)
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
