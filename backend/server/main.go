package main

import (
	"context"
	"fmt"
	"os"

	"github.com/NinjaPerson24119/MapProject/backend/internal/database"
)

func main() {
	ctx := context.Background()
	connectionURL := os.Getenv("POSTGRES_CONNECTION_URL")

	_, err := database.New(ctx, connectionURL)
	if err != nil {
		os.Exit(postgresConnectionFailed)
	}

	fmt.Println("Successfully connected to postgres")
	os.Exit(successCode)
}
