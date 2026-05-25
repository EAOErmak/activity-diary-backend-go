package diary

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"activitydiary/api/internal/auth"
	"activitydiary/api/internal/common"
	"activitydiary/api/internal/models"
	"activitydiary/api/internal/response"
	"activitydiary/api/internal/tags"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Service struct {
	db         *gorm.DB
	tagService *tags.Service
}

type MetricValueInput struct {
	UnitID uint    `json:"unitId"`
	Value  float64 `json:"value"`
}

type MetricInput struct {
	ID           uint               `json:"id"`
	MetricTypeID uint               `json:"metricTypeId"`
	Values       []MetricValueInput `json:"values"`
}

type UpsertRequest struct {
	WhenStarted *time.Time    `json:"whenStarted"`
	WhenEnded   *time.Time    `json:"whenEnded"`
	Mood        *int          `json:"mood"`
	Description *string       `json:"description"`
	Tags        []string      `json:"tags"`
	Status      string        `json:"status"`
	Metrics     []MetricInput `json:"metrics"`
}

type ListItem struct {
	ID          uint       `json:"id"`
	WhenStarted *time.Time `json:"whenStarted"`
	WhenEnded   *time.Time `json:"whenEnded"`
	Status      string     `json:"status"`
	FirstTag    *string    `json:"firstTag"`
}

type MetricValueDTO struct {
	UnitID   uint    `json:"unitId"`
	UnitName string  `json:"unitName"`
	Value    float64 `json:"value"`
}

type MetricDTO struct {
	ID             uint             `json:"id"`
	MetricTypeID   uint             `json:"metricTypeId"`
	MetricTypeName string           `json:"metricTypeName"`
	Values         []MetricValueDTO `json:"values"`
}

type EntryDTO struct {
	ID          uint        `json:"id"`
	WhenStarted *time.Time  `json:"whenStarted"`
	WhenEnded   *time.Time  `json:"whenEnded"`
	Duration    *int        `json:"duration"`
	Mood        *int        `json:"mood"`
	Description *string     `json:"description"`
	Status      string      `json:"status"`
	FirstTag    *string     `json:"firstTag"`
	UserID      uint        `json:"userId"`
	Metrics     []MetricDTO `json:"metrics"`
	CreatedAt   string      `json:"createdAt"`
	UpdatedAt   string      `json:"updatedAt"`
}

func NewService(db *gorm.DB, tagService *tags.Service) *Service {
	return &Service{db: db, tagService: tagService}
}

func (s *Service) ListMine(userID uint, page, size int, status, now, from, to string, tagNames []string) (*common.Page[ListItem], error) {
	query := s.db.Model(&models.DiaryEntry{}).Where("user_id = ?", userID)
	if status != "" {
		if status == "OVERDUE" {
			comparisonTime := time.Now().UTC()
			if now != "" {
				if parsed, err := time.Parse(time.RFC3339, now); err == nil {
					comparisonTime = parsed
				}
			}
			query = query.Where("status = ? AND when_started < ?", "PLANNED", comparisonTime)
		} else {
			query = query.Where("status = ?", status)
		}
	}
	if from != "" {
		query = query.Where("when_started >= ?", from)
	}
	if to != "" {
		query = query.Where("when_started <= ?", to)
	}
	if len(tagNames) > 0 {
		query = query.Joins("JOIN diary_entry_tags det ON det.diary_entry_id = diary_entries.id").
			Joins("JOIN tags t ON t.id = det.tag_id").
			Where("t.name IN ?", tagNames).
			Group("diary_entries.id")
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}
	var entries []models.DiaryEntry
	if err := query.Order("COALESCE(when_started, created_at) DESC").Offset(page * size).Limit(size).Find(&entries).Error; err != nil {
		return nil, err
	}
	content := make([]ListItem, 0, len(entries))
	for _, entry := range entries {
		firstTag := s.lookupFirstTag(entry.ID)
		content = append(content, ListItem{
			ID:          entry.ID,
			WhenStarted: entry.WhenStarted,
			WhenEnded:   entry.WhenEnded,
			Status:      resolveStatus(entry, now),
			FirstTag:    firstTag,
		})
	}
	return &common.Page[ListItem]{
		Content:       content,
		TotalElements: total,
		TotalPages:    common.TotalPages(total, size),
		Size:          size,
		Number:        page,
	}, nil
}

func (s *Service) Get(userID, id uint) (*EntryDTO, error) {
	var entry models.DiaryEntry
	if err := s.db.Preload("Metrics.Values").First(&entry, "id = ? AND user_id = ?", id, userID).Error; err != nil {
		return nil, err
	}
	return s.buildEntryDTO(entry), nil
}

func (s *Service) Create(userID uint, req UpsertRequest) (*EntryDTO, error) {
	if err := s.tagService.EnsureTags(req.Tags); err != nil {
		return nil, err
	}
	entry := models.DiaryEntry{
		UserID:      userID,
		WhenStarted: req.WhenStarted,
		WhenEnded:   req.WhenEnded,
		Mood:        req.Mood,
		Description: req.Description,
		Status:      deriveStatus(req),
	}
	if req.WhenStarted != nil && req.WhenEnded != nil {
		duration := int(req.WhenEnded.Sub(*req.WhenStarted).Minutes())
		entry.Duration = &duration
	}

	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&entry).Error; err != nil {
			return err
		}
		if err := saveTags(tx, entry.ID, req.Tags); err != nil {
			return err
		}
		return saveMetrics(tx, entry.ID, req.Metrics)
	})
	if err != nil {
		return nil, err
	}
	return s.Get(userID, entry.ID)
}

func (s *Service) Update(userID, id uint, req UpsertRequest) (*EntryDTO, error) {
	var entry models.DiaryEntry
	if err := s.db.First(&entry, "id = ? AND user_id = ?", id, userID).Error; err != nil {
		return nil, err
	}
	if req.WhenStarted != nil {
		entry.WhenStarted = req.WhenStarted
	}
	if req.WhenEnded != nil {
		entry.WhenEnded = req.WhenEnded
	}
	if req.Mood != nil {
		entry.Mood = req.Mood
	}
	if req.Description != nil {
		entry.Description = req.Description
	}
	if req.Status != "" {
		entry.Status = req.Status
	} else {
		entry.Status = deriveStatus(req)
	}
	if entry.WhenStarted != nil && entry.WhenEnded != nil {
		duration := int(entry.WhenEnded.Sub(*entry.WhenStarted).Minutes())
		entry.Duration = &duration
	}

	err := s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(&entry).Error; err != nil {
			return err
		}
		if len(req.Tags) > 0 {
			if err := s.tagService.EnsureTags(req.Tags); err != nil {
				return err
			}
			if err := tx.Exec("DELETE FROM diary_entry_tags WHERE diary_entry_id = ?", entry.ID).Error; err != nil {
				return err
			}
			if err := saveTags(tx, entry.ID, req.Tags); err != nil {
				return err
			}
		}
		if err := tx.Exec("DELETE FROM diary_entry_metric_values WHERE diary_entry_metric_id IN (SELECT id FROM diary_entry_metrics WHERE diary_entry_id = ?)", entry.ID).Error; err != nil {
			return err
		}
		if err := tx.Exec("DELETE FROM diary_entry_metrics WHERE diary_entry_id = ?", entry.ID).Error; err != nil {
			return err
		}
		return saveMetrics(tx, entry.ID, req.Metrics)
	})
	if err != nil {
		return nil, err
	}
	return s.Get(userID, entry.ID)
}

func (s *Service) Delete(userID, id uint) error {
	result := s.db.Model(&models.DiaryEntry{}).Where("id = ? AND user_id = ?", id, userID).Update("status", "DELETED")
	if result.RowsAffected == 0 {
		return errors.New("not found")
	}
	return result.Error
}

func saveTags(tx *gorm.DB, entryID uint, names []string) error {
	for _, name := range names {
		var tag models.Tag
		if err := tx.Where("LOWER(name) = LOWER(?)", strings.TrimSpace(name)).First(&tag).Error; err != nil {
			return err
		}
		if err := tx.Create(&models.DiaryEntryTag{DiaryEntryID: entryID, TagID: tag.ID}).Error; err != nil {
			return err
		}
	}
	return nil
}

func saveMetrics(tx *gorm.DB, entryID uint, metrics []MetricInput) error {
	for _, metric := range metrics {
		row := models.DiaryEntryMetric{DiaryEntryID: entryID, MetricTypeID: metric.MetricTypeID}
		if err := tx.Create(&row).Error; err != nil {
			return err
		}
		for _, value := range metric.Values {
			if err := tx.Create(&models.DiaryEntryMetricValue{
				DiaryEntryMetricID: row.ID,
				UnitID:             value.UnitID,
				Value:              value.Value,
			}).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

func deriveStatus(req UpsertRequest) string {
	if req.Status != "" {
		return req.Status
	}
	now := time.Now().UTC()
	if req.WhenEnded != nil {
		return "FINISHED"
	}
	if req.WhenStarted != nil && req.WhenStarted.Before(now) {
		return "ACTIVE"
	}
	return "PLANNED"
}

func resolveStatus(entry models.DiaryEntry, nowRaw string) string {
	if entry.Status != "PLANNED" {
		return entry.Status
	}
	reference := time.Now().UTC()
	if nowRaw != "" {
		if parsed, err := time.Parse(time.RFC3339, nowRaw); err == nil {
			reference = parsed
		}
	}
	if entry.WhenStarted != nil && entry.WhenStarted.Before(reference) {
		return "OVERDUE"
	}
	return entry.Status
}

func (s *Service) lookupFirstTag(entryID uint) *string {
	var result struct{ Name string }
	if err := s.db.Table("diary_entry_tags AS det").
		Joins("JOIN tags t ON t.id = det.tag_id").
		Where("det.diary_entry_id = ?", entryID).
		Order("t.name ASC").
		Limit(1).
		Scan(&result).Error; err != nil || result.Name == "" {
		return nil
	}
	return &result.Name
}

func (s *Service) buildEntryDTO(entry models.DiaryEntry) *EntryDTO {
	metrics := make([]MetricDTO, 0, len(entry.Metrics))
	for _, metric := range entry.Metrics {
		var metricName struct{ Label string }
		_ = s.db.Table("dictionary_items").Select("label").Where("id = ?", metric.MetricTypeID).Scan(&metricName).Error
		values := make([]MetricValueDTO, 0, len(metric.Values))
		for _, value := range metric.Values {
			var unit struct{ Label string }
			_ = s.db.Table("dictionary_items").Select("label").Where("id = ?", value.UnitID).Scan(&unit).Error
			values = append(values, MetricValueDTO{
				UnitID:   value.UnitID,
				UnitName: unit.Label,
				Value:    value.Value,
			})
		}
		metrics = append(metrics, MetricDTO{
			ID:             metric.ID,
			MetricTypeID:   metric.MetricTypeID,
			MetricTypeName: metricName.Label,
			Values:         values,
		})
	}
	return &EntryDTO{
		ID:          entry.ID,
		WhenStarted: entry.WhenStarted,
		WhenEnded:   entry.WhenEnded,
		Duration:    entry.Duration,
		Mood:        entry.Mood,
		Description: entry.Description,
		Status:      entry.Status,
		FirstTag:    s.lookupFirstTag(entry.ID),
		UserID:      entry.UserID,
		Metrics:     metrics,
		CreatedAt:   entry.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:   entry.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

type Handler struct {
	service serviceAPI
}

type serviceAPI interface {
	ListMine(userID uint, page, size int, status, now, from, to string, tagNames []string) (*common.Page[ListItem], error)
	Get(userID, id uint) (*EntryDTO, error)
	Create(userID uint, req UpsertRequest) (*EntryDTO, error)
	Update(userID, id uint, req UpsertRequest) (*EntryDTO, error)
	Delete(userID, id uint) error
}

func NewHandler(service serviceAPI) *Handler {
	return &Handler{service: service}
}

func (h *Handler) ListMine(c *gin.Context) {
	data, err := h.service.ListMine(
		auth.UserIDFromContext(c),
		parseInt(c.Query("page"), 0),
		parseInt(c.Query("size"), 8),
		c.Query("status"),
		c.Query("now"),
		c.Query("from"),
		c.Query("to"),
		c.QueryArray("tags"),
	)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to load diary entries")
		return
	}
	response.OK(c, http.StatusOK, data)
}

func (h *Handler) GetByID(c *gin.Context) {
	data, err := h.service.Get(auth.UserIDFromContext(c), uint(parseInt(c.Param("id"), 0)))
	if err != nil {
		response.Error(c, http.StatusNotFound, "Diary entry not found")
		return
	}
	response.OK(c, http.StatusOK, data)
}

func (h *Handler) Create(c *gin.Context) {
	var req UpsertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid diary entry request")
		return
	}
	data, err := h.service.Create(auth.UserIDFromContext(c), req)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Failed to create diary entry")
		return
	}
	response.OK(c, http.StatusCreated, data)
}

func (h *Handler) Update(c *gin.Context) {
	var req UpsertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid diary entry request")
		return
	}
	data, err := h.service.Update(auth.UserIDFromContext(c), uint(parseInt(c.Param("id"), 0)), req)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Failed to update diary entry")
		return
	}
	response.OK(c, http.StatusOK, data)
}

func (h *Handler) Delete(c *gin.Context) {
	if err := h.service.Delete(auth.UserIDFromContext(c), uint(parseInt(c.Param("id"), 0))); err != nil {
		response.Error(c, http.StatusNotFound, "Diary entry not found")
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
