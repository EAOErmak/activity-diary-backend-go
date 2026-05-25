package handlers

import (
	"net/http"
	"strconv"

	"activitydiary/analytics/internal/service"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *service.Service
}

func New(service *service.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) GetChartTypes(c *gin.Context) {
	tagID := uint(parseInt(c.Param("tagId"), 0))
	data, err := h.service.GetChartTypes(tagID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to load chart types"})
		return
	}
	c.JSON(http.StatusOK, data)
}

func (h *Handler) GetChart(c *gin.Context) {
	userID := uint(parseInt(c.GetHeader("X-User-ID"), 0))
	tagID := uint(parseInt(c.Query("tagId"), 0))
	chartType := c.Query("chartType")
	if err := h.service.EnsureAllowedChartType(tagID, chartType); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}
	data, err := h.service.GetChart(userID, tagID, chartType, c.Query("dateFrom"), c.Query("dateTo"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"message": "failed to load chart"})
		return
	}
	c.JSON(http.StatusOK, data)
}

func parseInt(raw string, fallback int) int {
	value, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	return value
}
