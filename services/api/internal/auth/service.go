package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"activitydiary/api/internal/models"
	"activitydiary/api/internal/response"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	ErrInvalidCredentials  = errors.New("invalid email or password")
	ErrInvalidRefreshToken = errors.New("invalid refresh token")
)

type JWTManager struct {
	secret          []byte
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
}

type Claims struct {
	UserID uint   `json:"userId"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

type Service struct {
	db  *gorm.DB
	jwt *JWTManager
}

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required,min=6"`
	FullName string `json:"fullName" binding:"required"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}

type AuthResponse struct {
	AccessToken       string `json:"accessToken"`
	RefreshToken      string `json:"refreshToken"`
	Username          string `json:"username"`
	UserID            uint   `json:"userId"`
	Role              string `json:"role"`
	TwoFactorRequired bool   `json:"twoFactorRequired"`
}

func NewJWTManager(secret string, accessTTL, refreshTTL time.Duration) *JWTManager {
	return &JWTManager{
		secret:          []byte(secret),
		accessTokenTTL:  accessTTL,
		refreshTokenTTL: refreshTTL,
	}
}

func NewService(db *gorm.DB, jwtManager *JWTManager) *Service {
	return &Service{db: db, jwt: jwtManager}
}

func (s *Service) Register(req RegisterRequest) (map[string]string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := models.User{
		Email:        req.Email,
		Username:     req.Username,
		FullName:     req.FullName,
		PasswordHash: string(hash),
		Role:         "USER",
		Enabled:      true,
	}
	if err := s.db.Create(&user).Error; err != nil {
		return nil, err
	}

	return map[string]string{"message": "Registration successful"}, nil
}

func (s *Service) Login(req LoginRequest) (*AuthResponse, error) {
	var user models.User
	if err := s.db.Where("email = ?", req.Email).First(&user).Error; err != nil {
		return nil, ErrInvalidCredentials
	}
	if !user.Enabled || bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)) != nil {
		return nil, ErrInvalidCredentials
	}
	return s.issueTokens(&user)
}

func (s *Service) Refresh(req RefreshRequest) (*AuthResponse, error) {
	tokenHash := hashToken(req.RefreshToken)
	var stored models.RefreshToken
	if err := s.db.Where("token_hash = ? AND revoked = FALSE", tokenHash).First(&stored).Error; err != nil {
		return nil, ErrInvalidRefreshToken
	}
	if time.Now().After(stored.ExpiresAt) {
		return nil, ErrInvalidRefreshToken
	}

	var claims Claims
	token, err := jwt.ParseWithClaims(req.RefreshToken, &claims, func(token *jwt.Token) (interface{}, error) {
		return s.jwt.secret, nil
	})
	if err != nil || !token.Valid {
		return nil, ErrInvalidRefreshToken
	}

	var user models.User
	if err := s.db.First(&user, claims.UserID).Error; err != nil {
		return nil, ErrInvalidRefreshToken
	}

	if err := s.db.Model(&stored).Update("revoked", true).Error; err != nil {
		return nil, err
	}

	return s.issueTokens(&user)
}

func (s *Service) issueTokens(user *models.User) (*AuthResponse, error) {
	accessToken, err := s.jwt.createToken(user.ID, user.Role, s.jwt.accessTokenTTL)
	if err != nil {
		return nil, err
	}
	refreshToken, err := s.jwt.createToken(user.ID, user.Role, s.jwt.refreshTokenTTL)
	if err != nil {
		return nil, err
	}

	if err := s.db.Create(&models.RefreshToken{
		UserID:    user.ID,
		TokenHash: hashToken(refreshToken),
		Revoked:   false,
		ExpiresAt: time.Now().Add(s.jwt.refreshTokenTTL),
	}).Error; err != nil {
		return nil, err
	}

	return &AuthResponse{
		AccessToken:       accessToken,
		RefreshToken:      refreshToken,
		Username:          user.Username,
		UserID:            user.ID,
		Role:              user.Role,
		TwoFactorRequired: false,
	}, nil
}

func (m *JWTManager) createToken(userID uint, role string, ttl time.Duration) (string, error) {
	claims := Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secret)
}

func (m *JWTManager) Parse(tokenString string) (*Claims, error) {
	var claims Claims
	token, err := jwt.ParseWithClaims(tokenString, &claims, func(token *jwt.Token) (interface{}, error) {
		return m.secret, nil
	})
	if err != nil || !token.Valid {
		return nil, errors.New("invalid token")
	}
	return &claims, nil
}

func hashToken(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}

type Handler struct {
	service handlerService
}

type handlerService interface {
	Register(req RegisterRequest) (map[string]string, error)
	Login(req LoginRequest) (*AuthResponse, error)
	Refresh(req RefreshRequest) (*AuthResponse, error)
}

func NewHandler(service handlerService) *Handler {
	return &Handler{service: service}
}

func UserIDFromContext(c *gin.Context) uint {
	value, _ := c.Get("userID")
	if id, ok := value.(uint); ok {
		return id
	}
	return 0
}

func (h *Handler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "Invalid register request")
		return
	}
	data, err := h.service.Register(req)
	if err != nil {
		response.Error(c, 400, "Failed to register user")
		return
	}
	response.OK(c, 201, data)
}

func (h *Handler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "Invalid login request")
		return
	}
	data, err := h.service.Login(req)
	if err != nil {
		response.Error(c, 401, "Invalid email or password")
		return
	}
	response.OK(c, 200, data)
}

func (h *Handler) Refresh(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "Invalid refresh request")
		return
	}
	data, err := h.service.Refresh(req)
	if err != nil {
		response.Error(c, 401, "Invalid refresh token")
		return
	}
	response.OK(c, 200, data)
}
