package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

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

	_, err := database.New(ctx, connectionURL)
	if err != nil {
		os.Exit(postgresConnectionFailed)
	}
	fmt.Println("Successfully connected to postgres")

	fmt.Println("Serving...")
	router := setupRouter()
	router.Run(":8080")

	os.Exit(successCode)
}
