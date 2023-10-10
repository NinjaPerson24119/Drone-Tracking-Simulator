package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/NinjaPerson24119/MapProject/backend/internal/api"
	"github.com/NinjaPerson24119/MapProject/backend/internal/constants"
	"github.com/NinjaPerson24119/MapProject/backend/internal/database"
	"github.com/NinjaPerson24119/MapProject/backend/internal/simulator"
	"github.com/gin-gonic/gin"

	"runtime/pprof"
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

	// profiler
	cpuProfileFile, err := os.Create("cpu.pprof")
	if err != nil {
		os.Exit(-1)
	}
	defer cpuProfileFile.Close()
	pprof.StartCPUProfile(cpuProfileFile)
	defer pprof.StopCPUProfile()
	memProfileFile, err := os.Create("mem.pprof")
	if err != nil {
		os.Exit(-1)
	}
	defer memProfileFile.Close()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		sig := <-sigChan
		fmt.Printf("Received signal: %v\n", sig)
		pprof.StopCPUProfile()
		cpuProfileFile.Close()
		memProfileFile.Close()
		os.Exit(0)
	}()

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
	simulator := simulator.New(repo, constants.SimulatedDevices, latitude, longitude, 0.25, 30, 0.025)
	ctxWithCancel, cancel := context.WithCancel(ctx)
	defer cancel()
	go simulator.Run(ctxWithCancel)

	router := setupBaseRouter()
	api.RouterWithGeolocationAPI(router, repo)
	router.Run(":8080")

	os.Exit(successCode)
}
