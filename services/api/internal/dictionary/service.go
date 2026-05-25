package dictionary

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"activitydiary/api/internal/common"
	"activitydiary/api/internal/models"
	"activitydiary/api/internal/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Service struct {
	db *gorm.DB
}

type ItemDTO struct {
	ID          uint    `json:"id"`
	Type        string  `json:"type"`
	Label       string  `json:"label"`
	Active      bool    `json:"active"`
	AllowedRole *string `json:"allowedRole"`
	ParentID    *uint   `json:"parentId"`
	CreatedAt   string  `json:"createdAt"`
	UpdatedAt   string  `json:"updatedAt"`
}

type AdminUpsertRequest struct {
	Type        string  `json:"type"`
	Label       string  `json:"label"`
	AllowedRole *string `json:"allowedRole"`
	Active      *bool   `json:"active"`
}

func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

func (s *Service) ListAll() ([]ItemDTO, error) {
	var items []models.DictionaryItem
	if err := s.db.Order("type, label").Find(&items).Error; err != nil {
		return nil, err
	}
	return mapItems(items), nil
}

func (s *Service) ListMetricUnits(metricNameID uint, page, limit int, q string) (*common.PageResponse[map[string]interface{}], error) {
	var total int64
	query := s.db.Table("metric_name_unit_links AS l").
		Joins("JOIN dictionary_items d ON d.id = l.metric_unit_id").
		Where("l.metric_name_id = ?", metricNameID)
	if q != "" {
		query = query.Where("LOWER(d.label) LIKE ?", strings.ToLower(q)+"%")
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}
	rows := []struct {
		ID    uint
		Label string
	}{}
	if err := query.Select("d.id, d.label").Order("d.label ASC").Offset(page * limit).Limit(limit).Scan(&rows).Error; err != nil {
		return nil, err
	}
	items := make([]map[string]interface{}, 0, len(rows))
	for _, row := range rows {
		items = append(items, map[string]interface{}{"id": row.ID, "label": row.Label})
	}
	return &common.PageResponse[map[string]interface{}]{
		Items:         items,
		Page:          page,
		Limit:         limit,
		TotalElements: total,
		TotalPages:    common.TotalPages(total, limit),
		HasNext:       int64((page+1)*limit) < total,
		HasPrevious:   page > 0,
	}, nil
}

func (s *Service) ListAdmin(itemType string, page, limit int, q string) (*common.PageResponse[ItemDTO], error) {
	var total int64
	query := s.db.Model(&models.DictionaryItem{}).Where("type = ?", itemType)
	if q != "" {
		query = query.Where("LOWER(label) LIKE ?", strings.ToLower(q)+"%")
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}
	var items []models.DictionaryItem
	if err := query.Order("active DESC, label ASC").Offset(page * limit).Limit(limit).Find(&items).Error; err != nil {
		return nil, err
	}
	return &common.PageResponse[ItemDTO]{
		Items:         mapItems(items),
		Page:          page,
		Limit:         limit,
		TotalElements: total,
		TotalPages:    common.TotalPages(total, limit),
		HasNext:       int64((page+1)*limit) < total,
		HasPrevious:   page > 0,
	}, nil
}

func (s *Service) Search(q string) ([]ItemDTO, error) {
	var items []models.DictionaryItem
	query := s.db.Model(&models.DictionaryItem{})
	if strings.TrimSpace(q) != "" {
		query = query.Where("LOWER(label) LIKE ?", strings.ToLower(strings.TrimSpace(q))+"%")
	}
	if err := query.Order("label ASC").Limit(25).Find(&items).Error; err != nil {
		return nil, err
	}
	return mapItems(items), nil
}

func (s *Service) Create(req AdminUpsertRequest) (*ItemDTO, error) {
	item := models.DictionaryItem{
		Type:        req.Type,
		Label:       strings.TrimSpace(req.Label),
		Active:      true,
		AllowedRole: req.AllowedRole,
	}
	if req.Active != nil {
		item.Active = *req.Active
	}
	if err := s.db.Create(&item).Error; err != nil {
		return nil, err
	}
	dto := mapItems([]models.DictionaryItem{item})[0]
	return &dto, nil
}

func (s *Service) Update(id uint, req AdminUpsertRequest) (*ItemDTO, error) {
	var item models.DictionaryItem
	if err := s.db.First(&item, id).Error; err != nil {
		return nil, err
	}
	if req.Label != "" {
		item.Label = strings.TrimSpace(req.Label)
	}
	if req.Active != nil {
		item.Active = *req.Active
	}
	item.AllowedRole = req.AllowedRole
	if err := s.db.Save(&item).Error; err != nil {
		return nil, err
	}
	dto := mapItems([]models.DictionaryItem{item})[0]
	return &dto, nil
}

func mapItems(items []models.DictionaryItem) []ItemDTO {
	result := make([]ItemDTO, 0, len(items))
	for _, item := range items {
		result = append(result, ItemDTO{
			ID:          item.ID,
			Type:        item.Type,
			Label:       item.Label,
			Active:      item.Active,
			AllowedRole: item.AllowedRole,
			ParentID:    item.ParentID,
			CreatedAt:   item.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"),
			UpdatedAt:   item.UpdatedAt.UTC().Format("2006-01-02T15:04:05Z"),
		})
	}
	return result
}

type Handler struct {
	service serviceAPI
}

type serviceAPI interface {
	ListAll() ([]ItemDTO, error)
	ListMetricUnits(metricNameID uint, page, limit int, q string) (*common.PageResponse[map[string]interface{}], error)
	ListAdmin(itemType string, page, limit int, q string) (*common.PageResponse[ItemDTO], error)
	Search(q string) ([]ItemDTO, error)
	Create(req AdminUpsertRequest) (*ItemDTO, error)
	Update(id uint, req AdminUpsertRequest) (*ItemDTO, error)
}

func NewHandler(service serviceAPI) *Handler {
	return &Handler{service: service}
}

func (h *Handler) ListAll(c *gin.Context) {
	items, err := h.service.ListAll()
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to load dictionary items")
		return
	}
	response.OK(c, http.StatusOK, items)
}

func (h *Handler) ListMetricUnits(c *gin.Context) {
	metricNameID, _ := strconv.Atoi(c.Param("metricNameId"))
	page := parseInt(c.Query("page"), 0)
	limit := parseInt(c.Query("limit"), 10)
	data, err := h.service.ListMetricUnits(uint(metricNameID), page, limit, c.Query("q"))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to load metric units")
		return
	}
	response.OK(c, http.StatusOK, data)
}

func (h *Handler) ListAdmin(c *gin.Context) {
	itemType := strings.ToUpper(c.Param("type"))
	page := parseInt(c.Query("page"), 0)
	limit := parseInt(c.Query("limit"), 10)
	data, err := h.service.ListAdmin(itemType, page, limit, c.Query("q"))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, fmt.Sprintf("Failed to load %s items", itemType))
		return
	}
	response.OK(c, http.StatusOK, data)
}

func (h *Handler) SearchAdmin(c *gin.Context) {
	items, err := h.service.Search(c.Query("q"))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to search dictionary")
		return
	}
	response.OK(c, http.StatusOK, items)
}

func (h *Handler) CreateAdmin(c *gin.Context) {
	var req AdminUpsertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid dictionary request")
		return
	}
	item, err := h.service.Create(req)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Failed to create dictionary item")
		return
	}
	response.OK(c, http.StatusCreated, item)
}

func (h *Handler) UpdateAdmin(c *gin.Context) {
	var req AdminUpsertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid dictionary request")
		return
	}
	id := uint(parseInt(c.Param("id"), 0))
	item, err := h.service.Update(id, req)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Failed to update dictionary item")
		return
	}
	response.OK(c, http.StatusOK, item)
}

func parseInt(raw string, fallback int) int {
	value, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	return value
}
