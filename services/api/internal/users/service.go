package users

import (
	"activitydiary/api/internal/auth"
	"activitydiary/api/internal/models"
	"activitydiary/api/internal/response"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Service struct {
	db *gorm.DB
}

type UserDTO struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	FullName string `json:"fullName"`
	Role     string `json:"role"`
	Enabled  bool   `json:"enabled"`
}

func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

func (s *Service) Me(userID uint) (*UserDTO, error) {
	var user models.User
	if err := s.db.First(&user, userID).Error; err != nil {
		return nil, err
	}
	return &UserDTO{
		ID:       user.ID,
		Username: user.Username,
		FullName: user.FullName,
		Role:     user.Role,
		Enabled:  user.Enabled,
	}, nil
}

type Handler struct {
	service serviceAPI
}

type serviceAPI interface {
	Me(userID uint) (*UserDTO, error)
}

func NewHandler(service serviceAPI) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Me(c *gin.Context) {
	user, err := h.service.Me(auth.UserIDFromContext(c))
	if err != nil {
		response.Error(c, http.StatusNotFound, "User not found")
		return
	}
	response.OK(c, http.StatusOK, user)
}
