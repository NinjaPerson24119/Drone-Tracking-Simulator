package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/NinjaPerson24119/MapProject/backend/internal/api"
	"github.com/NinjaPerson24119/MapProject/backend/internal/constants"
	"github.com/NinjaPerson24119/MapProject/backend/internal/database"
	"github.com/NinjaPerson24119/MapProject/backend/internal/simulator"
	"github.com/gin-gonic/gin"
)

func setupBaseRouter() *gin.Engine {
	router := gin.Default()

	router.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	return router
}

func main() {
	ctx := context.Background()

	connectionURL := os.Getenv("POSTGRES_CONNECTION_URL")
	repo, err := database.New(ctx, connectionURL)
	if err != nil {
		os.Exit(postgresConnectionFailed)
	}
	defer repo.Close()
	fmt.Println("Connected to postgres")

	// Edmonton legislature
	latitude := 53.5357
	longitude := -113.5068
	simulator := simulator.New(repo, constants.SimulatedDevices, latitude, longitude, 0.25/3, 10, 0.025/4)
	ctxWithCancel, cancel := context.WithCancel(ctx)
	defer cancel()
	go simulator.Run(ctxWithCancel)

	router := setupBaseRouter()
	api.RouterWithGeolocationAPI(router, repo)
	router.Run(":8080")

	os.Exit(successCode)
}
