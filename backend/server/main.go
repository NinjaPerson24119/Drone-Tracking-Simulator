package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/NinjaPerson24119/MapProject/backend/internal/api"
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
	fmt.Println("Connected to postgres")

	// Edmonton legislature
	latitude := 53.5357
	longitude := -113.5068
	simulator := simulator.New(repo, 100, latitude, longitude, 0.1, 60, 0.0001)
	ctxWithCancel, cancel := context.WithCancel(ctx)
	defer cancel()
	go simulator.Run(ctxWithCancel)

	router := setupBaseRouter()
	// TODO: this should take a service instead of a repo, but right now the service layer would be just a relay
	api.RouterWithGeolocationAPI(router, repo)
	router.Run(":8080")

	os.Exit(successCode)
}
