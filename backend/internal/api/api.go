package api

import (
	"net/http"

	"github.com/NinjaPerson24119/MapProject/backend/internal/database"
	"github.com/NinjaPerson24119/MapProject/backend/internal/filters"
	"github.com/gin-gonic/gin"
)

type AddDeviceResponse struct {
	DeviceID string `json:"device_id"`
}

type ListDevicesRequest struct {
	Paging filters.PageOptions `json:"paging"`
}

type GetDevicesResponse struct {
	Devices []*database.Device `json:"devices"`
}

type GetLatestGeolocationsRequest struct {
	Paging filters.PageOptions `json:"paging"`
	// TODO: maybe add filters later by (lat,lng) and radius
}

type GetLatestGeolocationsResponse struct {
	Geolocations []*database.DeviceGeolocation `json:"geolocations"`
}

func RouterWithGeolocationAPI(router *gin.Engine, repo database.Repo) {
	router.POST("/device/create", func(c *gin.Context) {
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

	router.POST("/device/list", func(c *gin.Context) {
		var request ListDevicesRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if request.Paging.Page < 1 || request.Paging.PageSize < 1 || request.Paging.PageSize > 1000 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid page or page_size"})
			return
		}

		devices, err := repo.ListDevices(c.Request.Context(), request.Paging)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		resp := GetDevicesResponse{
			Devices: devices,
		}
		c.JSON(http.StatusOK, resp)
	})

	router.POST("/geolocation/create", func(c *gin.Context) {
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

	router.POST("/geolocation/:deviceID", func(c *gin.Context) {
		deviceID := c.Param("deviceID")
		geolocation, err := repo.GetLatestGeolocation(c.Request.Context(), deviceID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if geolocation == nil {
			c.Status(http.StatusNotFound)
			return
		}
		c.JSON(http.StatusOK, geolocation)
	})

	router.POST("/geolocation/list", func(c *gin.Context) {
		var request GetLatestGeolocationsRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if request.Paging.Page < 1 || request.Paging.PageSize < 1 || request.Paging.PageSize > 1000 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid page or page_size"})
			return
		}

		geolocations, err := repo.GetLatestGeolocations(c.Request.Context(), request.Paging)
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

	router.GET("/geolocation/stream", geolocationsWebSocketGenerator(repo))
}
