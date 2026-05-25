package auth

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type mockHandlerService struct{}

func (mockHandlerService) Register(req RegisterRequest) (map[string]string, error) {
	return map[string]string{"message": "Registration successful"}, nil
}
func (mockHandlerService) Login(req LoginRequest) (*AuthResponse, error) {
	if req.Password != "password" {
		return nil, errors.New("bad password")
	}
	return &AuthResponse{AccessToken: "a", RefreshToken: "r", Username: "admin", UserID: 1, Role: "ADMIN", TwoFactorRequired: false}, nil
}
func (mockHandlerService) Refresh(req RefreshRequest) (*AuthResponse, error) {
	return &AuthResponse{AccessToken: "a2", RefreshToken: "r2", Username: "admin", UserID: 1, Role: "ADMIN", TwoFactorRequired: false}, nil
}

func TestLoginSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/login", NewHandler(mockHandlerService{}).Login)
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBufferString(`{"email":"admin@example.com","password":"password"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"success":true`)
	assert.Contains(t, rec.Body.String(), `"username":"admin"`)
}

func TestLoginInvalidPassword(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.POST("/login", NewHandler(mockHandlerService{}).Login)
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBufferString(`{"email":"admin@example.com","password":"bad"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Contains(t, rec.Body.String(), `"success":false`)
}
