package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"activitydiary/api/internal/auth"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestGetCurrentUserWithoutTokenReturns401(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	manager := auth.NewJWTManager("secret", 0, 0)
	router.GET("/user/me", Auth(manager), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/user/me", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Contains(t, rec.Body.String(), `"success":false`)
}
