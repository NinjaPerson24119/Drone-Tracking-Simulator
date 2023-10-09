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

type ListLatestGeolocationsRequest struct {
	Paging filters.PageOptions `json:"paging"`
	// TODO: maybe add filters later by (lat,lng) and radius
}

type ListLatestGeolocationsResponse struct {
	Geolocations []*database.DeviceGeolocation `json:"geolocations"`
}

type GetMultiLatestGeolocationsRequest struct {
	DeviceIDs []string `json:"device_ids"`
}

type GetMultiLatestGeolocationsResponse struct {
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

	router.POST("/geolocation/getMulti", func(c *gin.Context) {
		var request GetMultiLatestGeolocationsRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		geolocation, err := repo.GetMultiLatestGeolocations(c.Request.Context(), request.DeviceIDs)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		resp := GetMultiLatestGeolocationsResponse{
			Geolocations: geolocation,
		}
		c.JSON(http.StatusOK, resp)
	})

	router.POST("/geolocation/list", func(c *gin.Context) {
		var request ListLatestGeolocationsRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if request.Paging.Page < 1 || request.Paging.PageSize < 1 || request.Paging.PageSize > 1000 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid page or page_size"})
			return
		}

		geolocations, err := repo.ListLatestGeolocations(c.Request.Context(), request.Paging)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if len(geolocations) == 0 {
			geolocations = []*database.DeviceGeolocation{}
		}
		resp := ListLatestGeolocationsResponse{
			Geolocations: geolocations,
		}
		c.JSON(http.StatusOK, resp)
	})

	//router.GET("/geolocation/stream", geolocationsWebSocketGenerator(repo))
}
