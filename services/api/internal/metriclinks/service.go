package metriclinks

import (
	"net/http"
	"strconv"

	"activitydiary/api/internal/common"
	"activitydiary/api/internal/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Service struct {
	db *gorm.DB
}

type LinkRequest struct {
	MetricNameID uint `json:"metricNameId"`
	MetricUnitID uint `json:"metricUnitId"`
}

type LinkResponse struct {
	ID    uint   `json:"id"`
	Label string `json:"label"`
}

func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

func (s *Service) ListUnits(metricNameID uint, page, limit int) (*common.PageResponse[LinkResponse], error) {
	var total int64
	query := s.db.Table("metric_name_unit_links AS l").
		Joins("JOIN dictionary_items d ON d.id = l.metric_unit_id").
		Where("l.metric_name_id = ?", metricNameID)
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}
	items := []LinkResponse{}
	if err := query.Select("d.id, d.label").Order("d.label ASC").Offset(page * limit).Limit(limit).Scan(&items).Error; err != nil {
		return nil, err
	}
	return &common.PageResponse[LinkResponse]{
		Items:         items,
		Page:          page,
		Limit:         limit,
		TotalElements: total,
		TotalPages:    common.TotalPages(total, limit),
		HasNext:       int64((page+1)*limit) < total,
		HasPrevious:   page > 0,
	}, nil
}

func (s *Service) Create(req LinkRequest) (*LinkResponse, error) {
	if err := s.db.Exec("INSERT INTO metric_name_unit_links (metric_name_id, metric_unit_id) VALUES (?, ?) ON CONFLICT DO NOTHING", req.MetricNameID, req.MetricUnitID).Error; err != nil {
		return nil, err
	}
	resp := LinkResponse{}
	if err := s.db.Table("dictionary_items").Select("id, label").Where("id = ?", req.MetricUnitID).Scan(&resp).Error; err != nil {
		return nil, err
	}
	return &resp, nil
}

func (s *Service) Delete(metricNameID, metricUnitID uint) error {
	return s.db.Exec("DELETE FROM metric_name_unit_links WHERE metric_name_id = ? AND metric_unit_id = ?", metricNameID, metricUnitID).Error
}

type Handler struct {
	service serviceAPI
}

type serviceAPI interface {
	ListUnits(metricNameID uint, page, limit int) (*common.PageResponse[LinkResponse], error)
	Create(req LinkRequest) (*LinkResponse, error)
	Delete(metricNameID, metricUnitID uint) error
}

func NewHandler(service serviceAPI) *Handler {
	return &Handler{service: service}
}

func (h *Handler) ListUnitsByMetricName(c *gin.Context) {
	metricNameID := uint(parseInt(c.Param("metricNameId"), 0))
	page := parseInt(c.Query("page"), 0)
	limit := parseInt(c.Query("limit"), 10)
	data, err := h.service.ListUnits(metricNameID, page, limit)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to load metric links")
		return
	}
	response.OK(c, http.StatusOK, data)
}

func (h *Handler) Create(c *gin.Context) {
	var req LinkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid metric link request")
		return
	}
	data, err := h.service.Create(req)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Failed to create metric link")
		return
	}
	response.OK(c, http.StatusCreated, data)
}

func (h *Handler) Delete(c *gin.Context) {
	metricNameID := uint(parseInt(c.Query("metricNameId"), 0))
	metricUnitID := uint(parseInt(c.Query("metricUnitId"), 0))
	if err := h.service.Delete(metricNameID, metricUnitID); err != nil {
		response.Error(c, http.StatusBadRequest, "Failed to delete metric link")
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
