package tags

import (
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

type TagDTO struct {
	ID     uint   `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
}

type TagMetricDTO struct {
	TagID           uint   `json:"tagId"`
	MetricNameID    uint   `json:"metricNameId"`
	MetricNameLabel string `json:"metricNameLabel"`
}

type CreateRequest struct {
	Name string `json:"name"`
}

type UpdateTagMetricsRequest struct {
	MetricNameIDs []uint `json:"metricNameIds"`
}

func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

func (s *Service) EnsureTags(names []string) error {
	for _, name := range names {
		trimmed := strings.TrimSpace(name)
		if trimmed == "" {
			continue
		}
		if err := s.db.Where("LOWER(name) = LOWER(?)", trimmed).FirstOrCreate(&models.Tag{}, models.Tag{Name: trimmed, Status: "PROPOSED"}).Error; err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) ListPublic(q string) ([]TagDTO, error) {
	var tags []models.Tag
	query := s.db.Model(&models.Tag{}).Where("status = ?", "APPROVED")
	if strings.TrimSpace(q) != "" {
		query = query.Where("LOWER(name) LIKE ?", strings.ToLower(strings.TrimSpace(q))+"%")
	}
	if err := query.Order("name ASC").Find(&tags).Error; err != nil {
		return nil, err
	}
	return mapTags(tags), nil
}

func (s *Service) ListMetricsByTags(tagIDs []uint, page, limit int, q string) (*common.PageResponse[map[string]interface{}], error) {
	var total int64
	query := s.db.Table("tag_metric_links AS l").
		Joins("JOIN dictionary_items d ON d.id = l.metric_name_id").
		Where("l.tag_id IN ?", tagIDs).Distinct("d.id, d.label")
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

func (s *Service) ListAdmin(page, size int, q string) (*common.Slice[TagDTO], error) {
	var tags []models.Tag
	query := s.db.Model(&models.Tag{})
	if strings.TrimSpace(q) != "" {
		query = query.Where("LOWER(name) LIKE ?", strings.ToLower(strings.TrimSpace(q))+"%")
	}
	if err := query.Order("id DESC").Offset(page * size).Limit(size).Find(&tags).Error; err != nil {
		return nil, err
	}
	dtos := mapTags(tags)
	return &common.Slice[TagDTO]{
		Content:          dtos,
		Number:           page,
		Size:             size,
		First:            page == 0,
		Last:             len(dtos) < size,
		NumberOfElements: len(dtos),
		Empty:            len(dtos) == 0,
	}, nil
}

func (s *Service) Create(name string) (*TagDTO, error) {
	tag := models.Tag{Name: strings.TrimSpace(name), Status: "PROPOSED"}
	if err := s.db.Create(&tag).Error; err != nil {
		return nil, err
	}
	dto := mapTags([]models.Tag{tag})[0]
	return &dto, nil
}

func (s *Service) Update(id uint, name string) (*TagDTO, error) {
	var tag models.Tag
	if err := s.db.First(&tag, id).Error; err != nil {
		return nil, err
	}
	tag.Name = strings.TrimSpace(name)
	if err := s.db.Save(&tag).Error; err != nil {
		return nil, err
	}
	dto := mapTags([]models.Tag{tag})[0]
	return &dto, nil
}

func (s *Service) SetStatus(id uint, status string) error {
	return s.db.Model(&models.Tag{}).Where("id = ?", id).Update("status", status).Error
}

func (s *Service) Delete(id uint) error {
	return s.db.Delete(&models.Tag{}, id).Error
}

func (s *Service) ListTagMetrics(tagID uint) ([]TagMetricDTO, error) {
	rows := []TagMetricDTO{}
	if err := s.db.Table("tag_metric_links AS l").
		Joins("JOIN dictionary_items d ON d.id = l.metric_name_id").
		Where("l.tag_id = ?", tagID).
		Select("l.tag_id, l.metric_name_id, d.label AS metric_name_label").
		Order("d.label ASC").
		Scan(&rows).Error; err != nil {
		return nil, err
	}
	return rows, nil
}

func (s *Service) ReplaceTagMetrics(tagID uint, metricNameIDs []uint) ([]TagMetricDTO, error) {
	returnList := []TagMetricDTO{}
	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("DELETE FROM tag_metric_links WHERE tag_id = ?", tagID).Error; err != nil {
			return err
		}
		for _, id := range metricNameIDs {
			if err := tx.Exec("INSERT INTO tag_metric_links (tag_id, metric_name_id) VALUES (?, ?) ON CONFLICT DO NOTHING", tagID, id).Error; err != nil {
				return err
			}
		}
		rows := []TagMetricDTO{}
		if err := tx.Table("tag_metric_links AS l").
			Joins("JOIN dictionary_items d ON d.id = l.metric_name_id").
			Where("l.tag_id = ?", tagID).
			Select("l.tag_id, l.metric_name_id, d.label AS metric_name_label").
			Order("d.label ASC").
			Scan(&rows).Error; err != nil {
			return err
		}
		returnList = rows
		return nil
	})
	return returnList, err
}

func mapTags(tags []models.Tag) []TagDTO {
	result := make([]TagDTO, 0, len(tags))
	for _, tag := range tags {
		result = append(result, TagDTO{ID: tag.ID, Name: tag.Name, Status: tag.Status})
	}
	return result
}

type Handler struct {
	service serviceAPI
}

type serviceAPI interface {
	ListPublic(q string) ([]TagDTO, error)
	ListMetricsByTags(tagIDs []uint, page, limit int, q string) (*common.PageResponse[map[string]interface{}], error)
	ListAdmin(page, size int, q string) (*common.Slice[TagDTO], error)
	Create(name string) (*TagDTO, error)
	Update(id uint, name string) (*TagDTO, error)
	SetStatus(id uint, status string) error
	Delete(id uint) error
	ListTagMetrics(tagID uint) ([]TagMetricDTO, error)
	ReplaceTagMetrics(tagID uint, metricNameIDs []uint) ([]TagMetricDTO, error)
}

func NewHandler(service serviceAPI) *Handler {
	return &Handler{service: service}
}

func (h *Handler) ListPublic(c *gin.Context) {
	items, err := h.service.ListPublic(c.Query("q"))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to load tags")
		return
	}
	response.OK(c, http.StatusOK, items)
}

func (h *Handler) ListMetricsByTags(c *gin.Context) {
	rawIDs := c.QueryArray("tagIds")
	ids := make([]uint, 0, len(rawIDs))
	for _, raw := range rawIDs {
		ids = append(ids, uint(parseInt(raw, 0)))
	}
	data, err := h.service.ListMetricsByTags(ids, parseInt(c.Query("page"), 0), parseInt(c.Query("limit"), 10), c.Query("q"))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to load tag metrics")
		return
	}
	response.OK(c, http.StatusOK, data)
}

func (h *Handler) ListAdmin(c *gin.Context) {
	data, err := h.service.ListAdmin(parseInt(c.Query("page"), 0), parseInt(c.Query("size"), 20), c.Query("q"))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to load admin tags")
		return
	}
	response.OK(c, http.StatusOK, data)
}

func (h *Handler) CreateAdmin(c *gin.Context) {
	var req CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid tag request")
		return
	}
	tag, err := h.service.Create(req.Name)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Failed to create tag")
		return
	}
	response.OK(c, http.StatusCreated, tag)
}

func (h *Handler) UpdateAdmin(c *gin.Context) {
	var req CreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid tag request")
		return
	}
	tag, err := h.service.Update(uint(parseInt(c.Param("id"), 0)), req.Name)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Failed to update tag")
		return
	}
	response.OK(c, http.StatusOK, tag)
}

func (h *Handler) Approve(c *gin.Context)   { h.setStatus(c, "APPROVED") }
func (h *Handler) Reject(c *gin.Context)    { h.setStatus(c, "REJECTED") }
func (h *Handler) Deprecate(c *gin.Context) { h.setStatus(c, "DEPRECATED") }

func (h *Handler) setStatus(c *gin.Context, status string) {
	if err := h.service.SetStatus(uint(parseInt(c.Param("id"), 0)), status); err != nil {
		response.Error(c, http.StatusBadRequest, "Failed to update tag status")
		return
	}
	response.OK(c, http.StatusOK, gin.H{"updated": true})
}

func (h *Handler) DeleteAdmin(c *gin.Context) {
	if err := h.service.Delete(uint(parseInt(c.Param("id"), 0))); err != nil {
		response.Error(c, http.StatusBadRequest, "Failed to delete tag")
		return
	}
	response.OK(c, http.StatusOK, gin.H{"deleted": true})
}

func (h *Handler) ListTagMetrics(c *gin.Context) {
	rows, err := h.service.ListTagMetrics(uint(parseInt(c.Param("id"), 0)))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to load tag metrics")
		return
	}
	response.OK(c, http.StatusOK, rows)
}

func (h *Handler) ReplaceTagMetrics(c *gin.Context) {
	var req UpdateTagMetricsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid tag metrics request")
		return
	}
	rows, err := h.service.ReplaceTagMetrics(uint(parseInt(c.Param("id"), 0)), req.MetricNameIDs)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Failed to replace tag metrics")
		return
	}
	response.OK(c, http.StatusOK, rows)
}

func parseInt(raw string, fallback int) int {
	value, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	return value
}
