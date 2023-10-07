package api

import (
	"net/http"

	"github.com/NinjaPerson24119/MapProject/backend/internal/database"
	"github.com/gin-gonic/gin"
)

type AddDeviceResponse struct {
	DeviceID string `json:"device_id"`
}

type GetLatestGeolocationsRequest struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
	// maybe add filters later by (lat,lng) and radius
}

type GetLatestGeolocationsResponse struct {
	Geolocations []*database.DeviceGeolocation `json:"geolocations"`
}

func RouterWithGeolocationAPI(router *gin.Engine, repo database.Repo) {
	router.POST("/device", func(c *gin.Context) {
		var request database.Device
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if request.Name == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "missing name"})
			return
		}

		id, err := repo.InsertDevice(c.Request.Context(), &request)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		resp := AddDeviceResponse{
			DeviceID: id,
		}
		c.JSON(http.StatusCreated, resp)
	})

	router.POST("/geolocation", func(c *gin.Context) {
		var request database.DeviceGeolocation
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if request.DeviceID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "missing device_id"})
			return
		}
		if request.Latitude < -90 || request.Latitude > 90 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid latitude"})
			return
		}
		if request.Longitude < -180 || request.Longitude > 180 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid longitude"})
			return
		}
		if request.EventTime.IsZero() {
			c.JSON(http.StatusBadRequest, gin.H{"error": "missing event_time"})
		}
		err := repo.InsertGeolocation(c.Request.Context(), &request)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.Status(http.StatusCreated)
	})

	router.POST("/geolocations", func(c *gin.Context) {
		var request GetLatestGeolocationsRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if request.Page < 1 || request.PageSize > 1000 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid page or page_size"})
			return
		}

		geolocations, err := repo.GetLatestGeolocations(c.Request.Context(), request.Page, request.PageSize)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if len(geolocations) == 0 {
			geolocations = []*database.DeviceGeolocation{}
		}
		resp := GetLatestGeolocationsResponse{
			Geolocations: geolocations,
		}
		c.JSON(http.StatusOK, resp)
	})
}
