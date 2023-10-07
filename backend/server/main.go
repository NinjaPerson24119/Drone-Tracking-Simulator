package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/NinjaPerson24119/MapProject/backend/internal/api"
	"github.com/NinjaPerson24119/MapProject/backend/internal/database"
	"github.com/gin-gonic/gin"
)

func setupRouter() *gin.Engine {
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

	router := setupRouter()
	api.RouterWithGeolocationAPI(router, repo)
	router.Run(":8080")

	os.Exit(successCode)
}
