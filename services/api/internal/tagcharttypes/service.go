package tagcharttypes

import (
	"net/http"
	"strconv"

	"activitydiary/api/internal/models"
	"activitydiary/api/internal/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Service struct {
	db *gorm.DB
}

type LinkDTO struct {
	TagID     uint   `json:"tagId"`
	ChartType string `json:"chartType"`
}

type CreateRequest struct {
	TagID     uint   `json:"tagId"`
	ChartType string `json:"chartType"`
}

func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

func (s *Service) ListByTag(tagID uint) ([]LinkDTO, error) {
	var rows []models.TagChartTypeLink
	if err := s.db.Where("tag_id = ?", tagID).Order("chart_type ASC").Find(&rows).Error; err != nil {
		return nil, err
	}
	result := make([]LinkDTO, 0, len(rows))
	for _, row := range rows {
		result = append(result, LinkDTO{TagID: row.TagID, ChartType: row.ChartType})
	}
	return result, nil
}

func (s *Service) Create(req CreateRequest) (*LinkDTO, error) {
	row := models.TagChartTypeLink{TagID: req.TagID, ChartType: req.ChartType}
	if err := s.db.Where("tag_id = ? AND chart_type = ?", req.TagID, req.ChartType).FirstOrCreate(&row).Error; err != nil {
		return nil, err
	}
	return &LinkDTO{TagID: row.TagID, ChartType: row.ChartType}, nil
}

func (s *Service) Delete(tagID uint, chartType string) error {
	return s.db.Where("tag_id = ? AND chart_type = ?", tagID, chartType).Delete(&models.TagChartTypeLink{}).Error
}

type Handler struct {
	service serviceAPI
}

type serviceAPI interface {
	ListByTag(tagID uint) ([]LinkDTO, error)
	Create(req CreateRequest) (*LinkDTO, error)
	Delete(tagID uint, chartType string) error
}

func NewHandler(service serviceAPI) *Handler {
	return &Handler{service: service}
}

func (h *Handler) ListByTag(c *gin.Context) {
	tagID := uint(parseInt(c.Param("tagId"), 0))
	items, err := h.service.ListByTag(tagID)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to load tag chart types")
		return
	}
	response.OK(c, http.StatusOK, items)
}

func (h *Handler) Create(c *gin.Context) {
	var req CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid tag chart type request")
		return
	}
	item, err := h.service.Create(req)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Failed to create tag chart type link")
		return
	}
	response.OK(c, http.StatusCreated, item)
}

func (h *Handler) Delete(c *gin.Context) {
	tagID := uint(parseInt(c.Query("tagId"), 0))
	if err := h.service.Delete(tagID, c.Query("chartType")); err != nil {
		response.Error(c, http.StatusBadRequest, "Failed to delete tag chart type link")
		return
	}
	response.OK(c, http.StatusOK, gin.H{"deleted": true})
}

func parseInt(raw string, fallback int) int {
	value, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	return value
}
