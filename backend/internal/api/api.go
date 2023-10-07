package api

import (
	"context"
	"net/http"

	"github.com/NinjaPerson24119/MapProject/backend/internal/database"
	"github.com/gin-gonic/gin"
)

type GeolocationAPI interface {
	GetLatestGeolocations(ctx context.Context, page int, pageSize int) ([]*database.Device, error)
	AddDevice(ctx context.Context, device *database.Device) (string, error)
	AddGeolocation(ctx context.Context, geolocation *database.DeviceGeolocation) error
}

type AddDeviceResponse struct {
	DeviceID string `json:"device_id"`
}

type GetLatestGeolocationsRequest struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
}

type GetLatestGeolocationsResponse struct {
	Geolocations []*database.Device `json:"geolocations"`
}

func RouterWithGeolocationAPI(router *gin.Engine, repo database.Repo) {
	router.POST("/device", func(c *gin.Context) {
		var request database.Device
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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

		c.Status(http.StatusCreated)
	})

	router.GET("/geolocations", func(c *gin.Context) {
		var request GetLatestGeolocationsRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if request.Page < 1 || request.Page < 1 || request.Page > 1000 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid page or pageSize"})
		}

		geolocations, err := s.repo.GetLatestGeolocations(ctx, page, pageSize)
		if err != nil {
			return nil, err
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})
}

type GeolocationAPIImpl struct {
	repo database.Repo
}

func New(repo database.Repo) *GeolocationAPIImpl {
	return &GeolocationAPIImpl{
		repo: repo,
	}
}

func (s *GeolocationAPIImpl) GetLatestGeolocations(ctx context.Context, page int, pageSize int) ([]*database.Device, error) {

	return geolocations, nil
}

func (s *GeolocationAPIImpl) AddDevice(ctx context.Context, device *database.Device) (string, error) {
	id, err := s.repo.InsertDevice(ctx, device)
	if err != nil {
		return "", err
	}

	return id, nil
}

func (s *GeolocationAPIImpl) AddGeolocation(ctx context.Context, geolocation *database.DeviceGeolocation) error {
	err := s.repo.InsertGeolocation(ctx, geolocation)
	if err != nil {
		return err
	}

	return nil
}
