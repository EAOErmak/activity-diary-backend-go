package foods

import (
	"net/http"
	"strconv"
	"strings"

	"activitydiary/api/internal/models"
	"activitydiary/api/internal/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Service struct {
	db *gorm.DB
}

type FoodDTO struct {
	ID                  uint    `json:"id"`
	DictionaryItemID    uint    `json:"dictionaryItemId"`
	DictionaryItemLabel string  `json:"dictionaryItemLabel"`
	Protein             float64 `json:"protein"`
	Fat                 float64 `json:"fat"`
	Carbs               float64 `json:"carbs"`
	Callories           float64 `json:"callories"`
}

type UpsertRequest struct {
	DictionaryItemID uint    `json:"dictionaryItemId"`
	Protein          float64 `json:"protein"`
	Fat              float64 `json:"fat"`
	Carbs            float64 `json:"carbs"`
	Callories        float64 `json:"callories"`
}

func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

func (s *Service) ListGeneralFoods(q string) ([]FoodDTO, error) {
	query := s.db.Table("general_foods AS gf").
		Joins("JOIN dictionary_items d ON d.id = gf.dictionary_item_id")
	if strings.TrimSpace(q) != "" {
		query = query.Where("LOWER(d.label) LIKE ?", strings.ToLower(strings.TrimSpace(q))+"%")
	}
	var items []FoodDTO
	if err := query.Select("gf.id, gf.dictionary_item_id, d.label AS dictionary_item_label, gf.protein, gf.fat, gf.carbs, gf.callories").
		Order("d.label ASC").Scan(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (s *Service) Get(id uint) (*FoodDTO, error) {
	var item FoodDTO
	err := s.db.Table("general_foods AS gf").
		Joins("JOIN dictionary_items d ON d.id = gf.dictionary_item_id").
		Select("gf.id, gf.dictionary_item_id, d.label AS dictionary_item_label, gf.protein, gf.fat, gf.carbs, gf.callories").
		Where("gf.id = ?", id).
		Scan(&item).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (s *Service) Create(req UpsertRequest) (*FoodDTO, error) {
	item := models.GeneralFood{
		DictionaryItemID: req.DictionaryItemID,
		Protein:          req.Protein,
		Fat:              req.Fat,
		Carbs:            req.Carbs,
		Callories:        req.Callories,
	}
	if err := s.db.Create(&item).Error; err != nil {
		return nil, err
	}
	return s.Get(item.ID)
}

func (s *Service) Update(id uint, req UpsertRequest) (*FoodDTO, error) {
	var item models.GeneralFood
	if err := s.db.First(&item, id).Error; err != nil {
		return nil, err
	}
	item.DictionaryItemID = req.DictionaryItemID
	item.Protein = req.Protein
	item.Fat = req.Fat
	item.Carbs = req.Carbs
	item.Callories = req.Callories
	if err := s.db.Save(&item).Error; err != nil {
		return nil, err
	}
	return s.Get(id)
}

func (s *Service) Delete(id uint) error {
	return s.db.Delete(&models.GeneralFood{}, id).Error
}

type Handler struct {
	service serviceAPI
}

type serviceAPI interface {
	ListGeneralFoods(q string) ([]FoodDTO, error)
	Get(id uint) (*FoodDTO, error)
	Create(req UpsertRequest) (*FoodDTO, error)
	Update(id uint, req UpsertRequest) (*FoodDTO, error)
	Delete(id uint) error
}

func NewHandler(service serviceAPI) *Handler {
	return &Handler{service: service}
}

func (h *Handler) ListGeneralFoods(c *gin.Context) {
	items, err := h.service.ListGeneralFoods(c.Query("q"))
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "Failed to load foods")
		return
	}
	response.OK(c, http.StatusOK, items)
}

func (h *Handler) GetGeneralFood(c *gin.Context) {
	id := uint(parseInt(c.Param("id"), 0))
	item, err := h.service.Get(id)
	if err != nil {
		response.Error(c, http.StatusNotFound, "Food not found")
		return
	}
	response.OK(c, http.StatusOK, item)
}

func (h *Handler) CreateGeneralFood(c *gin.Context) {
	var req UpsertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid food request")
		return
	}
	item, err := h.service.Create(req)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Failed to create general food")
		return
	}
	response.OK(c, http.StatusCreated, item)
}

func (h *Handler) UpdateGeneralFood(c *gin.Context) {
	var req UpsertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "Invalid food request")
		return
	}
	id := uint(parseInt(c.Param("id"), 0))
	item, err := h.service.Update(id, req)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "Failed to update general food")
		return
	}
	response.OK(c, http.StatusOK, item)
}

func (h *Handler) DeleteGeneralFood(c *gin.Context) {
	id := uint(parseInt(c.Param("id"), 0))
	if err := h.service.Delete(id); err != nil {
		response.Error(c, http.StatusBadRequest, "Failed to delete general food")
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
